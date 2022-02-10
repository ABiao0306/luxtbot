package luxtbot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ABiao0306/luxtbot/util"
)

var (
	BackenChain  []BackenUnit
	MsgChain     []MessageUnit
	CmdChain     []CommandUnit
	NoticeChain  []NoticeUnit
	RequestChain []RequestUnit

	PluginList []*Plugin
)

const (
	defaultPluginInfo = "写该插件的人很懒，没有留下任何信息！"
)

type Plugin struct {
	ID            int
	Name          string
	Enable        bool
	HelpInfo      string
	IsAdminPlugin bool
}

func NewPlugin(id int) *Plugin {
	var p Plugin
	p.ID = id
	p.Enable = true
	p.HelpInfo = defaultPluginInfo
	PluginList = append(PluginList, &p)
	return &p
}

func (p *Plugin) SetName(name string) *Plugin {
	p.Name = name
	return p
}

func (p *Plugin) SetHelpInfo(helpInfo string) *Plugin {
	p.HelpInfo = helpInfo
	return p
}

func (p *Plugin) SetAdminPlugin() *Plugin {
	p.IsAdminPlugin = true
	return p
}

func (p *Plugin) AddBackenUnit() *BackenUnit {
	var bp BackenUnit
	bp.Plg = p
	return &bp
}

func (p *Plugin) AddMessageUnit() *MessageUnit {
	var mp MessageUnit
	mp.Plg = p
	return &mp
}

func (p *Plugin) AddCommandUnit() *CommandUnit {
	var cp CommandUnit
	cp.Plg = p

	return &cp
}

func (p *Plugin) AddNoticeUnit() *NoticeUnit {
	var np NoticeUnit
	np.Plg = p
	return &np
}

func (p *Plugin) AddRequestUnit() *RequestUnit {
	var rp RequestUnit
	rp.Plg = p
	return &rp
}

type BackenUnit struct {
	Plg   *Plugin
	Init  func()
	Start func(bInfos []BotInfo)
}

func (bp *BackenUnit) SetInitFunc(f func()) *BackenUnit {
	bp.Init = f
	return bp
}

func (bp *BackenUnit) SetStartFunc(f func(bInfos []BotInfo)) *BackenUnit {
	bp.Start = f
	return bp
}

func (bp *BackenUnit) AddToBackenChain() {
	BackenChain = append(BackenChain, *bp)
}

type MessageUnit struct {
	Rule    *Rule
	Plg     *Plugin
	Process func(e *Event, bInfo BotInfo)
}

func (mp *MessageUnit) SetProcessor(f func(e *Event, bInfo BotInfo)) *MessageUnit {
	mp.Process = f
	return mp
}

func (mp *MessageUnit) SetRule(rule *Rule) *MessageUnit {
	mp.Rule = rule
	return mp
}

func (mp *MessageUnit) AddToMsgChain() {
	MsgChain = append(MsgChain, *mp)
}

// 命令以 ~$#或者@bot 开头
const ConmandPrefix = "~$#"

type CommandUnit struct {
	Plg     *Plugin
	Rule    *Rule
	Cmd     string
	Aliases []string
	Process func(e *Event, params []string, bInfo BotInfo)
}

func (cp *CommandUnit) SetCommand(cmd string) *CommandUnit {
	cp.Cmd = cmd
	return cp
}

func (cp *CommandUnit) AddAliases(aliases ...string) *CommandUnit {
	if len(cp.Aliases) == 0 {
		cp.Aliases = aliases
	} else {
		for _, alias := range aliases {
			cp.Aliases = append(cp.Aliases, alias)
		}
	}
	return cp
}

func (cp *CommandUnit) SetProcessor(f func(e *Event, params []string, bInfo BotInfo)) *CommandUnit {
	cp.Process = f
	return cp
}

func (cp *CommandUnit) SetRule(rule *Rule) *CommandUnit {
	cp.Rule = rule
	return cp
}

func (cp *CommandUnit) AddToCmdChain() {
	CmdChain = append(CmdChain, *cp)
}

func (cp *CommandUnit) matchCmd(cmd, qq string, botID int64) bool {
	if qq != "" && strconv.FormatInt(botID, 10) != qq {
		return false
	}
	if cmd == cp.Cmd {
		return true
	}
	for _, alias := range cp.Aliases {
		if cmd == alias {
			return true
		}
	}
	return false
}

// @return cmd, qqnum
func parseCmd(e *Event) (string, string, error) {
	msg := e.GetArrayMsg()
	cmd, text, qq := "", "", ""
	if len(msg) <= 0 {
		return "", "", nil
	}
	if msg[0].Type == AtMsgSeg {
		qq = msg[0].Data["qq"]
		if len(msg) < 2 {
			return cmd, qq, nil
		}
		text = util.Trim(msg[1].Data["text"])
		// LBLogger.Debugln("text is ", text)
		args := strings.SplitN(text, " ", 2)
		cmd = args[0]
	} else if msg[0].Type == TextMsgSeg {
		text = msg[0].Data["text"]
		if text[0] == '~' || text[0] == '$' || text[0] == '#' {
			args := strings.SplitN(text, " ", 2)
			cmd = args[0][1:]
		}
	}
	return cmd, qq, nil
}

func parseParams(e *Event) []string {
	msg := e.GetArrayMsg()
	var text string
	var args []string
	if msg[0].Type == AtMsgSeg {
		text = util.Trim(msg[1].Data["text"])
		// LBLogger.Debugln("text is ", text)
		args = strings.Split(text, " ")
	} else if msg[0].Type == TextMsgSeg {
		text = msg[0].Data["text"]
		if text[0] == '~' || text[0] == '$' || text[0] == '#' {
			args = strings.SplitN(text, " ", 2)
		}
	}
	if len(args) <= 1 {
		return []string{}
	}
	return args[1:]
}

type NoticeUnit struct {
	Rule    *Rule
	Plg     *Plugin
	Process func(e *Event, bInfo BotInfo)
}

func (np *NoticeUnit) SetProcessor(f func(e *Event, bInfo BotInfo)) *NoticeUnit {
	np.Process = f
	return np
}

func (np *NoticeUnit) SetRule(rule *Rule) *NoticeUnit {
	np.Rule = rule
	return np
}

func (np *NoticeUnit) AddToNoticeChain() {
	NoticeChain = append(NoticeChain, *np)
}

type RequestUnit struct {
	Rule    *Rule
	Plg     *Plugin
	Process func(e *Event, bInfo BotInfo)
}

func (rp *RequestUnit) SetProcessor(f func(e *Event, bInfo BotInfo)) *RequestUnit {
	rp.Process = f
	return rp
}

func (rp *RequestUnit) SetRule(rule *Rule) *RequestUnit {
	rp.Rule = rule
	return rp
}

func (rp *RequestUnit) AddToRequestChain() {
	RequestChain = append(RequestChain, *rp)
}

func InitDefaultPluginManager(plgId int) {
	plg := NewPlugin(plgId).SetName("Luxtbot插件管理").SetHelpInfo("Luxtbot默认插件管理").SetAdminPlugin()
	rule := NewRule().AddMustRules(IsAdmin)
	addQueryUnit(plg, rule)
	addOnOffUnit(plg, plgId, rule)
}

func addQueryUnit(plg *Plugin, rule *Rule) {
	plg.AddCommandUnit().SetCommand("help").AddAliases("???", "插件信息").SetRule(rule).SetProcessor(func(e *Event, params []string, bInfo BotInfo) {
		size := len(PluginList)
		msg := MakeArrayMsg(size)
		for _, plg := range PluginList {
			state := "ON"
			if !plg.Enable {
				state = "OFF"
			}
			msgText := fmt.Sprintf("%d. %v: %v - %v \n", plg.ID, plg.Name, plg.HelpInfo, state)
			msg.AddText(msgText)
		}
		sendMsg(msg, e, bInfo)
	}).AddToCmdChain()
}

func addOnOffUnit(plg *Plugin, selfID int, rule *Rule) {
	plg.AddCommandUnit().SetCommand("plgon").AddAliases("开启插件", "启用插件").SetRule(rule).SetProcessor(func(e *Event, params []string, bInfo BotInfo) {
		msg := MakeArrayMsg(1)
		var err error
		plg, err = searchPugin(params[0])
		if err != nil {
			msg.AddText(err.Error())
		} else if plg.ID == selfID {
			msg.AddText(fmt.Sprintln("无法禁用或开启插件管理插件！"))
		} else {
			plg.Enable = true
			msg.AddText(fmt.Sprintln("已开启插件：", plg.ID, plg.Name))
		}
		sendMsg(msg, e, bInfo)
	}).AddToCmdChain()
	plg.AddCommandUnit().SetCommand("plgoff").AddAliases("关闭插件", "禁用插件").SetRule(rule).SetProcessor(func(e *Event, params []string, bInfo BotInfo) {
		msg := MakeArrayMsg(1)
		var err error
		plg, err = searchPugin(params[0])
		if err != nil {
			msg.AddText(err.Error())
		} else if plg.ID == selfID {
			msg.AddText(fmt.Sprintln("无法禁用或开启插件管理插件！"))
		} else {
			plg.Enable = false
			msg.AddText(fmt.Sprintln("已关闭插件：", plg.ID, plg.Name))
		}
		sendMsg(msg, e, bInfo)
	}).AddToCmdChain()
}

func searchPugin(idStr string) (*Plugin, error) {
	var (
		id  int
		err error
	)
	id, err = strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New(fmt.Sprintln("插件id格式错误：", id))
	}
	l, r, m := 0, len(PluginList), 0
	for l <= r {
		m = (l + r) / 2
		if PluginList[m].ID > id {
			r = m - 1
		} else if PluginList[m].ID < id {
			l = m + 1
		} else {
			return PluginList[m], nil
		}
	}
	return nil, errors.New(fmt.Sprintln("未找到插件：", id))
}

func sendMsg(msg MsgBuilder, e *Event, bInfo BotInfo) {
	var api ApiPost
	if e.MessageType == MsgTypePrivate {
		api = MakePrivateMsg(msg, e.UserID)
	} else if e.MessageType == MsgTypeGroup {
		api = MakeGroupMsg(msg, e.GroupID)
	}
	_, err := api.Do(bInfo.BotID, false)
	if err != nil {
		LBLogger.WithField("BotName", bInfo.Name).Warnln(err)
	}
}
