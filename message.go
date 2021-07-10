package luxtbot

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	TextMsgSeg      = "text"
	FaceMsgSeg      = "face"
	ImageMsgSeg     = "image"
	RecordMsgSeg    = "record"
	VideoMsgSeg     = "video"
	AtMsgSeg        = "at"
	RPSMsgSeg       = "rps"
	DiceMsgSeg      = "dice"
	ShakeMsgSeg     = "shake"
	PokeMsgSeg      = "poke"
	AnonymousMsgSeg = "anonymous"
	ShareMsgSeg     = "share"
	ContactMsgSeg   = "contact"
	LocationMsgSeg  = "location"
	MusicMsgSeg     = "music"
	ReplyMsgSeg     = "reply"
	ForwardMsgSeg   = "forward"
	NodeMsgSeg      = "node"
	XmlMsgSeg       = "xml"
	JsonMsgSeg      = "json"
)

const (
	QEmojiBabble      = 13
	QEmojiFrog        = 170
	QEmojiFirecracker = 137
	QEmojiCrab        = 184
)

const AtAll = int64(-1)

type MsgBuilder interface {
	GetMsg() (interface{}, error)
}

type TextMsg struct {
	buf bytes.Buffer
}

func MakeTextMsg() *TextMsg {
	return &TextMsg{}
}

func (tm *TextMsg) NewLine() *TextMsg {
	tm.buf.WriteByte('\n')
	return tm
}

func (tm *TextMsg) AddText(text string) *TextMsg {
	tm.buf.WriteString(text)
	return tm
}

func (tm *TextMsg) AddImg(file, url string) *TextMsg {
	tm.buf.WriteString("[CQ:image,file='")
	tm.buf.WriteString(file)
	tm.buf.WriteString("']")
	return tm
}

func (tm *TextMsg) AddFace(faceNo int) *TextMsg {
	tm.buf.WriteString("[CQ:face,id=")
	tm.buf.WriteString(strconv.Itoa(faceNo))
	tm.buf.WriteByte(']')

	return tm
}

func (tm *TextMsg) AddAt(uid int64) *TextMsg {
	tm.buf.WriteString("[CQ:at,qq=")
	if uid == AtAll {
		tm.buf.WriteString("all")
	} else {
		tm.buf.WriteString(strconv.FormatInt(uid, 10))
	}
	tm.buf.WriteByte(']')
	return tm
}

func (tm *TextMsg) AddRecord(file, url string) *TextMsg {
	tm.buf.WriteString("[CQ:record,file='")
	tm.buf.WriteString(file)
	tm.buf.WriteString("']")
	return tm
}

func (tm *TextMsg) AddRPS() *TextMsg {
	tm.buf.WriteString("[CQ:rps]")
	return tm
}

func (tm *TextMsg) AddDice() *TextMsg {
	tm.buf.WriteString("[CQ:dice]")
	return tm
}

func (tm *TextMsg) AddShake() *TextMsg {
	tm.buf.WriteString("[CQ:shake]")
	return tm
}

func (tm *TextMsg) GetMsg() (interface{}, error) {
	if tm.buf.Len() == 0 {
		return "", errors.New("消息为空。")
	}
	return tm.buf.String(), nil
}

type MsgSeg struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

type ArrayMsg struct {
	Segs []MsgSeg
	Len  int
}

func MakeArrayMsg(size int) *ArrayMsg {
	msg := ArrayMsg{
		Segs: make([]MsgSeg, 0, size),
		Len:  0,
	}
	return &msg
}

func (am *ArrayMsg) initNewSeg(segType string) map[string]string {
	data := make(map[string]string)
	mSeg := MsgSeg{
		Type: segType,
		Data: data,
	}
	am.Segs = append(am.Segs, mSeg)
	am.Len++
	return am.Segs[len(am.Segs)-1].Data
}

func (am *ArrayMsg) AddText(text string) *ArrayMsg {
	data := am.initNewSeg(TextMsgSeg)
	data["text"] = text
	return am
}

func (am *ArrayMsg) AddImg(file, url string) *ArrayMsg {
	data := am.initNewSeg(ImageMsgSeg)
	data["file"] = file
	data["url"] = url
	return am
}

func (am *ArrayMsg) AddFace(faceID int) *ArrayMsg {
	data := am.initNewSeg(FaceMsgSeg)
	data["id"] = strconv.Itoa(faceID)
	return am
}

func (am *ArrayMsg) AddAt(uid int64) *ArrayMsg {
	data := am.initNewSeg(AtMsgSeg)
	if uid == AtAll {
		data["qq"] = "all"
	} else {
		data["qq"] = strconv.FormatInt(uid, 10)
	}
	return am
}

func (am *ArrayMsg) AddRecord(file, url string) *ArrayMsg {
	data := am.initNewSeg(RecordMsgSeg)
	data["file"] = file
	data["url"] = url
	return am
}

func (am *ArrayMsg) AddRPS() *ArrayMsg {
	am.initNewSeg(RPSMsgSeg)
	return am
}

func (am *ArrayMsg) AddDice() *ArrayMsg {
	am.initNewSeg(DiceMsgSeg)
	return am
}

func (am *ArrayMsg) AddShake() *ArrayMsg {
	am.initNewSeg(ShakeMsgSeg)
	return am
}

func (am *ArrayMsg) GetMsg() (interface{}, error) {
	if am.Len != len(am.Segs) {
		return nil, errors.New("消息段数与预期不符，检查创建方式是否正确")
	}
	return am.Segs, nil
}

func ParseTextMsg(segs []MsgSeg) string {
	tm := MakeTextMsg()
	for _, seg := range segs {
		switch seg.Type {
		case TextMsgSeg:
			tm.AddText(seg.Data["text"])
		case FaceMsgSeg:
			id, err := strconv.Atoi(seg.Data["id"])
			if err != nil {
				LBLogger.WithField("face id", seg.Data["id"]).Debugln("表情ID格式错误")
				break
			}
			tm.AddFace(id)
		case ImageMsgSeg:
			tm.AddImg(seg.Data["file"], seg.Data["url"])
		case RecordMsgSeg:
			tm.AddRecord(seg.Data["file"], seg.Data["url"])
		case AtMsgSeg:
			qqStr := seg.Data["qq"]
			if qqStr == "all" {
				tm.AddAt(AtAll)
			}
			qq, err := strconv.ParseInt(qqStr, 10, 64)
			if err != nil {
				LBLogger.WithField("QQ", qqStr).Debugln("AT QQ号错误")
				break
			}
			tm.AddAt(qq)
		case RPSMsgSeg:
			tm.AddRPS()
		case DiceMsgSeg:
			tm.AddDice()
		case ShakeMsgSeg:
			tm.AddShake()
		}
	}
	return tm.buf.String()
}

var SegRegex = regexp.MustCompile(`\[CQ:[a-z]{2,9}[^\]]*\]`)

func ParseMsgSegs(msg string) []MsgSeg {
	re := SegRegex.FindAllStringIndex(msg, 10)
	segs := make([]MsgSeg, 0, 5)
	l := 0
	for _, part := range re {
		if part[0] != l {
			seg := parseMsgSeg(msg[l:part[0]])
			segs = append(segs, seg)
		}
		seg := parseMsgSeg(msg[part[0]:part[1]])
		segs = append(segs, seg)
		l = part[1]
	}
	if l <= len(msg)-1 {
		seg := parseMsgSeg(msg[l:])
		segs = append(segs, seg)
	}
	return segs
}

func parseMsgSeg(segStr string) MsgSeg {
	if segStr[0] != '[' {
		data := make(map[string]string)
		data["text"] = segStr
		return MsgSeg{
			Type: TextMsgSeg,
			Data: data,
		}
	}
	strs := strings.Split(segStr, ",")
	seg := MsgSeg{
		Data: make(map[string]string),
	}
	if len(strs) == 1 {
		seg.Type = strs[0][4 : len(strs[0])-1]
	} else {
		seg.Type = strs[0][4:len(strs[0])]
		pair := strings.Split(strs[len(strs)-1], "=")
		seg.Data[pair[0]] = pair[1][:len(pair[1])-1]
	}
	for i := 1; i < len(strs)-1; i++ {
		pair := strings.Split(strs[1], "=")
		seg.Data[pair[0]] = pair[1]
	}
	return seg
}
