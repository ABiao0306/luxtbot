package luxtbot

import (
	"errors"
	"reflect"
	"time"

	lutil "github.com/ABiao0306/luxtbot/util"
)

const (
	MsgTypeArray  = "array"
	MsgTypeString = "string"
)

const (
	GroupMsgAction   = "send_group_msg"
	PrivateMsgAction = "send_private_msg"
)

const (
	MsgTypePrivate = "private"
	MsgTypeGroup   = "group"

	MessageEvent = "message"
	NoticeEvent  = "notice"
	RequestEvent = "request"
	MetaEvent    = "meta_event"

	Lifecycle = "lifecycle"
	Heartbeat = "heartbeat"

	SendTimeout = "timeout"
	BotIDMiss   = "miss"
)

type Event struct {
	Font          int         `json:"font"`
	GroupID       int64       `json:"group_id"`
	Message       interface{} `json:"message"`
	MessageID     int         `json:"message_id"`
	MessageSeq    int         `json:"message_seq"`
	MessageType   string      `json:"message_type"`
	PostType      string      `json:"post_type"`
	RawMessage    string      `json:"raw_message"`
	TempSource    int         `json:"temp_source"`
	SelfID        int64       `json:"self_id"`
	Sender        Sender      `json:"sender"`
	SubType       string      `json:"sub_type"`
	Time          int         `json:"time"`
	UserID        int64       `json:"user_id"`
	MetaEventType string      `json:"meta_event_type"`
}

var (
	reflectTypeOfString  = reflect.TypeOf("")
	reflectTypeOfMsgSegs = reflect.TypeOf([]MsgSeg{})
)

func (e *Event) GetArrayMsg() []MsgSeg {
	if reflect.TypeOf(e.Message) == reflectTypeOfString {
		msgSegs := ParseMsgSegs(e.Message.(string))
		return msgSegs
	}
	return e.Message.([]MsgSeg)
}

func (e *Event) GetTextMsg() string {
	if reflect.TypeOf(e.Message) == reflectTypeOfString {
		return e.Message.(string)
	}
	text := ParseTextMsg(e.Message.([]MsgSeg))
	return text
}

type Sender struct {
	Age      int    `json:"age"`
	Area     string `json:"area"`
	Card     string `json:"card"`
	Level    string `json:"level"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Sex      string `json:"sex"`
	Title    string `json:"title"`
	UserID   int64  `json:"user_id"`
}

type ApiPost struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
	Echo   string      `json:"echo"`
}

func (api ApiPost) Do(botID int64, needEcho bool) (string, error) {
	if needEcho {
		api.Echo = lutil.GetEchoStr()
	} else {
		api.Echo = ""
	}
	bCtx, err := getBotCtxByID(botID)
	if err != nil {
		LBLogger.WithField("BotID", botID).Warnln(err)
		return "", err
	}
	select {
	case bCtx.OutChan <- api:
		return api.Echo, nil
	case <-time.After(time.Second * 5):
		return "", errors.New("发送消息超时。")
	}
}

func makeApi(action string, params interface{}) ApiPost {
	api := ApiPost{
		Action: action,
		Params: params,
		Echo:   "",
	}
	return api
}

type PrivateMsg struct {
	Message    interface{} `json:"message"`
	AutoEscape bool        `json:"auto_escape"`
	UserID     int64       `json:"user_id"`
	GroupID    int64       `json:"group_id"`
}

type GroupMsg struct {
	Message    interface{} `json:"message"`
	AutoEscape bool        `json:"auto_escape"`
	GroupID    int64       `json:"group_id"`
}

type ApiResp struct {
	Data struct {
		MessageID int `json:"message_id"`
	} `json:"data"`
	Echo    string `jons:"echo"`
	Retcode int    `json:"retcode"`
	Status  string `json:"status"`
}

func MakeGroupMsg(msgBuilder MsgBuilder, groupID int64) ApiPost {
	msgData, err := msgBuilder.GetMsg()
	if err != nil {
		LBLogger.WithField("目标群", groupID).Infof("检查到消息异常，将放弃该条消息：%v", err)
	}
	msg := GroupMsg{
		GroupID:    groupID,
		AutoEscape: false,
		Message:    msgData,
	}
	return makeApi(GroupMsgAction, msg)
}

func MakePrivateMsg(msgBuilder MsgBuilder, userID int64) ApiPost {
	msgData, err := msgBuilder.GetMsg()
	if err != nil {
		LBLogger.WithField("目标QQ", userID).Infof("检查到消息异常，将放弃该条消息：%v", err)
	}
	msg := PrivateMsg{
		UserID:     userID,
		AutoEscape: false,
		Message:    msgData,
	}
	return makeApi(PrivateMsgAction, msg)
}
