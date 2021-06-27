package luxtbot

import (
	"bytes"
	"strconv"
)

type MsgBuilder interface {
	GetMsg() interface{}
}

type TextMsg struct {
	buf bytes.Buffer
}

func MkTextMsg() *TextMsg {
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

const (
	QEmojiBabble      = 13
	QEmojiFrog        = 170
	QEmojiFirecracker = 137
	QEmojiCrab        = 184
)

func (tm *TextMsg) AddFace(faceNo int) *TextMsg {
	tm.buf.WriteString("[CQ:face,id=")
	tm.buf.WriteString(strconv.Itoa(faceNo))
	tm.buf.WriteByte(']')

	return tm
}

const (
	AtAll = int64(-1)
)

func (tm *TextMsg) AddAt(uid int64, name string) *TextMsg {
	tm.buf.WriteString("[CQ:at,qq=")
	tm.buf.WriteString(strconv.FormatInt(uid, 10))
	tm.buf.WriteString(",name=")
	tm.buf.WriteString(name)
	tm.buf.WriteByte(']')
	return tm
}

func (tm *TextMsg) AddVoice(file, url string) *TextMsg {
	tm.buf.WriteString("[CQ:record,file='")
	tm.buf.WriteString(file)
	tm.buf.WriteString("']")
	return tm
}

func (tm *TextMsg) GetMsg() interface{} {
	return tm.buf.String()
}
