package luxtbot

type Rule = func(e *Event, bInfo BotInfo) bool

var (
    // 消息发送者是否是管理员
    IsAdmin = func(e *Event, bInfo BotInfo) bool {
        for _, admin := range bInfo.Admins {
            if e.UserID == admin {
                return true
            }
        }
        return false
    }

    // 发送者是否是超级管理员
    IsSAdmin = func(e *Event, bInfo BotInfo) bool {
        for _, admin := range Conf.SAdmins {
            if e.UserID == admin {
                return true
            }
        }
        return false
    }

    IsGroupMsg = func(e *Event, bInfo BotInfo) bool {
        return e.MessageType == MsgTypeGroup
    }

    IsPrivateMsg = func(e *Event, bInfo BotInfo) bool {
        return e.MessageType == MsgTypePrivate
    }
)