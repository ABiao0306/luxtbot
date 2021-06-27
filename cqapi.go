package luxtbot

import (
	"errors"
	"luxtbot/util"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	SendGroupMsgApi   = "send_group_msg"
	SendPrivateMsgApi = "send_private_msg"
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
	Font          int    `json:"font"`
	GroupID       int64  `json:"group_id"`
	Message       string `json:"message"`
	MessageID     int    `json:"message_id"`
	MessageSeq    int    `json:"message_seq"`
	MessageType   string `json:"message_type"`
	PostType      string `json:"post_type"`
	RawMessage    string `json:"raw_message"`
	TempSource    int    `json:"temp_source"`
	SelfID        int64  `json:"self_id"`
	Sender        Sender `json:"sender"`
	SubType       string `json:"sub_type"`
	Time          int    `json:"time"`
	UserID        int64  `json:"user_id"`
	MetaEventType string `json:"meta_event_type"`
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

func (api ApiPost) Send(botID int64) (string, error) {
	bCtx, err := getBotCtxByID(botID)
	if err != nil {
		logrus.WithField("BotID", botID).Warnln(err)
		return BotIDMiss, err
	}
	select {
	case bCtx.OutChan <- api:
		return api.Echo, nil
	case <-time.After(time.Second * 5):
		return SendTimeout, errors.New("发送消息超时")
	}
}

func useApi(action string, params interface{}) ApiPost {
	echo := util.GetEchoStr()
	api := ApiPost{
		Action: action,
		Params: params,
		Echo:   echo,
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
	msg := GroupMsg{
		GroupID:    groupID,
		AutoEscape: false,
		Message:    msgBuilder.GetMsg(),
	}
	return useApi(SendGroupMsgApi, msg)
}

func MakePrivateMsg(msgBuilder MsgBuilder, userID int64) ApiPost {
	msg := PrivateMsg{
		UserID:     userID,
		AutoEscape: false,
		Message:    msgBuilder.GetMsg(),
	}
	return useApi(SendGroupMsgApi, msg)
}
