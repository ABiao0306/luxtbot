package luxtbot

var (
	BackenChain  []BackenPlugin
	MsgChain     []MessagePlugin
	CmdChain     []CommandPlugin
	NoticeChain  []NoticePlugin
	RequestChain []RequestPlugin
)

const (
	defaultPluginInfo = "写该插件的人很懒，没有留下任何信息！"
)

type Plugin struct {
	ID       int
	Name     string
	Enable   bool
	HelpInfo string
	Must     []Rule 
	Or       []Rule
}

func NewPlugin(id int) *Plugin {
	var p Plugin
	p.ID = id
	p.Enable = true
	p.HelpInfo = defaultPluginInfo
	return &p
}

/**
 * @description: All in *Must* is match and (One of *Or* is math or len(Or) == 0)
 * @param *Event e
 * @param *BotContext bInfo
 * @return result bool
 */
func (p *Plugin) CheckRules(e *Event, bInfo BotInfo) bool {
	for _, rule := range p.Must {
		if !rule(e, bInfo) {
			return false
		}
	}
	for _, rule := range p.Or {
		if rule(e, bInfo) {
			return true
		}
	}
	return len(p.Or) == 0
}

func (p *Plugin) SetName(name string) *Plugin {
	p.Name = name
	return p
}

func (p *Plugin) SetHelpInfo(name string) *Plugin {
	p.Name = name
	return p
}

func (p *Plugin) AddMustRules(rs ...Rule) *Plugin {
	for _, r := range rs {
		p.Must = append(p.Must, r)
	}
	return p
}

func (p *Plugin) AddOrRules(rs ...Rule) *Plugin {
	for _, r := range rs {
		p.Or = append(p.Or, r)
	}
	return p
}

func (p *Plugin) Up2BackenPlugin() *BackenPlugin {
	var bp BackenPlugin
	bp.Plg = *p
	return &bp
}

func (p *Plugin) Up2MessagePlugin() *MessagePlugin {
	var mp MessagePlugin
	mp.Plg = *p
	return &mp
}

func (p *Plugin) Up2CommandPlugin() *CommandPlugin {
	var cp CommandPlugin
	cp.Plg = *p
	return &cp
}

func (p *Plugin) Up2NoticePlugin() *NoticePlugin {
	var np NoticePlugin
	np.Plg = *p
	return &np
}

func (p *Plugin) Up2RequestPlugin() *RequestPlugin {
	var rp RequestPlugin
	rp.Plg = *p
	return &rp
}

type BackenPlugin struct {
	Plg   Plugin
	Init  func()
	Start func(bots BotCtxs)
}

func (bp *BackenPlugin) SetInitFunc(f func()) *BackenPlugin {
	bp.Init = f
	return bp
}

func (bp *BackenPlugin) SetStartFunc(f func(bots []BotContext)) *BackenPlugin {
	bp.Start = f
	return bp
}

func (bp *BackenPlugin) AddToBackenChain() {
	BackenChain = append(BackenChain, *bp)
}

type MessagePlugin struct {
	Plg     Plugin
	Process func(e *Event, bInfo BotInfo)
}

func (mp *MessagePlugin) SetProcessor(f func(e *Event, bInfo BotInfo)) *MessagePlugin {
	mp.Process = f
	return mp
}

func (mp *MessagePlugin) AddToMsgChain() {
	MsgChain = append(MsgChain, *mp)
}

type CommandPlugin struct {
	Plg     Plugin
	Cmd     string
	Aliases []string
	Process func(e *Event, bInfo BotInfo)
}

func (cp *CommandPlugin) SetCommand(cmd string) *CommandPlugin {
	cp.Cmd = cmd
	return cp
}

func (cp *CommandPlugin) AddAliases(aliases ...string) *CommandPlugin {
	if len(cp.Aliases) == 0 {
		cp.Aliases = aliases
	} else {
		for _, alias := range aliases {
			cp.Aliases = append(cp.Aliases, alias)
		}
	}
	return cp
}

func (cp *CommandPlugin) SetProcessor(f func(e *Event, bInfo BotInfo)) *CommandPlugin {
	cp.Process = f
	return cp
}

func (cp *CommandPlugin) AddToCmdChain() {
	CmdChain = append(CmdChain, *cp)
}

type NoticePlugin struct {
	Plg     Plugin
	Process func(e *Event, bInfo BotInfo)
}

func (np *NoticePlugin) SetProcessor(f func(e *Event, bInfo BotInfo)) *NoticePlugin {
	np.Process = f
	return np
}

func (np *NoticePlugin) AddToNoticeChain() {
	NoticeChain = append(NoticeChain, *np)
}

type RequestPlugin struct {
	Plg     Plugin
	Process func(e *Event, bInfo BotInfo)
}

func (rp *RequestPlugin) SetProcessor(f func(e *Event, bInfo BotInfo)) *RequestPlugin {
	rp.Process = f
	return rp
}

func (rp *RequestPlugin) AddToRequestChain() {
	RequestChain = append(RequestChain, *rp)
}
