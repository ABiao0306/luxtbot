package luxtbot


type Rule struct {
	Must []Judge
	Or   []Judge
}

func NewRule() *Rule{
    return &Rule {
        Must: make([]Judge, 0),
        Or: make([]Judge, 0),
    }
}

/**
 * @description: All in *Must* is match and (One of *Or* is math or len(Or) == 0)
 * @param *Event e
 * @param *BotContext bInfo
 * @return result bool
 */
func (r *Rule) CheckRules(e *Event, bInfo BotInfo) bool {
    if r == nil {
        return true
    }
	for _, rule := range r.Must {
		if !rule(e, bInfo) {
			return false
		}
	}
	for _, rule := range r.Or {
		if rule(e, bInfo) {
			return true
		}
	}
	return len(r.Or) == 0
}

func (r *Rule) AddMustRules(rs ...Judge) *Rule {
	for _, j := range rs {
		r.Must = append(r.Must, j)
	}
	return r
}

func (r *Rule) AddOrRules(rs ...Judge) *Rule {
	for _, j := range rs {
		r.Or = append(r.Or, j)
	}
	return r
}

type Judge = func(e *Event, bInfo BotInfo) bool

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