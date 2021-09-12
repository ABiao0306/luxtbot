package luxtbot

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
)

const (
	TokenPrefix = "WhYhAvEsPaCe "
	ReconnTimes = 3

	heartOn    = 'O'
	idMismatch = 'X'
)

var (
	cqEventChan = make(chan eventContext, 10)
	cqRespChan  = make(chan apiRespContext, 10)

	onConnectChain    []OnconnectHook
	disConnectChain   []DisconnectHook
	eventInChain      []EventInHook
	beforeApiOutChain []BeforeApiOutHook
)

type eventContext struct {
	e    *Event
	bCtx *BotContext
}

type apiRespContext struct {
	resp *ApiResp
	bCtx *BotContext
}

func InitPluginList() {
	sort.SliceStable(PluginList, func(l, r int) bool {
		return PluginList[l].ID < PluginList[r].ID
	})
	for _, plg := range PluginList {
		LBLogger.Infoln("启动插件：", plg.ID, plg.Name)
	}
}

func InitBackenPlugin() {
	for _, bp := range BackenChain {
		if bp.Init != nil {
			bp.Init()
		}
	}
}

func InitBotCtxs() {
	for _, bot := range Conf.BotInfos {
		bCtx := BotContext{
			Conn:      nil,
			CloseLock: new(sync.Mutex),
			OutChan:   make(chan ApiPost, 10),
			FlagChan:  make(chan byte, 3),
			CloseChan: make(chan byte, 1),
			IsReady:   false,
			BotInfo:   &bot,
		}
		bots = append(bots, bCtx)
	}
}

func RunBots() {
	for i := range bots {
		go connCQServer(&bots[i], ReconnTimes)
		go heartCheck(&bots[i])
	}
}

func RunEventDispatcher() {
	go func() {
		for {
			eCtx := <-cqEventChan
			e, bCtx := eCtx.e, eCtx.bCtx
			switch e.PostType {
			case MessageEvent:
				for _, mp := range MsgChain {
					if !mp.Plg.Enable || !mp.Rule.CheckRules(e, *bCtx.BotInfo) {
						continue
					}
					go mp.Process(e, *bCtx.BotInfo)
				}
				cmd, qq := parseCmd(e)
				if cmd == "" {
					continue
				}
				for _, cp := range CmdChain {
					if !cp.Plg.Enable || !cp.matchCmd(cmd, qq, *&bCtx.BotInfo.BotID) || !cp.Rule.CheckRules(e, *bCtx.BotInfo) {
						continue
					}
					go cp.Process(e, parseParams(e), *bCtx.BotInfo)
				}
			case NoticeEvent:
				for _, np := range NoticeChain {
					if !np.Plg.Enable || !np.Rule.CheckRules(e, *bCtx.BotInfo) {
						continue
					}
					go np.Process(e, *bCtx.BotInfo)
				}
			case RequestEvent:
				for _, mp := range MsgChain {
					if !mp.Plg.Enable || !mp.Rule.CheckRules(e, *bCtx.BotInfo) {
						continue
					}
					go mp.Process(e, *bCtx.BotInfo)
				}
			case MetaEvent:
				processMateEvent(e, bCtx)
			}
		}
	}()
}

func RunRespDispatcher(poolSize int) {
	go func() {
		callBackPool = make(map[string]EchoCallback, poolSize)
		for {
			select {
			case respCtx := <-cqRespChan:
				if respCtx.resp.Echo != "" {
					doEchoCallback(respCtx.resp, respCtx.bCtx)
				}
			}
		}
	}()
}

type EchoCallback func(apiResp *ApiResp, bInfo BotInfo)

var callBackPool map[string]EchoCallback

func AddEchoCallback(echo string, callback EchoCallback) {
	callBackPool[echo] = callback
}

func doEchoCallback(apiResp *ApiResp, bCtx *BotContext) {
	if len(apiResp.Echo) == 0 {
		return
	}
	callback := callBackPool[apiResp.Echo]
	if callback == nil {
		LBLogger.WithField("BotName", bCtx.BotInfo.Name).WithField("Echo", apiResp.Echo).Infoln("找不到api回复回调函数。")
		return
	}
	go callback(apiResp, *bCtx.BotInfo)
	callBackPool[apiResp.Echo] = nil
}

func RunBackenPlugin() {
	for _, bp := range BackenChain {
		if bp.Start == nil {
			LBLogger.WithField("Plugin", bp.Plg.Name).Debugln("the start func of backen plugin is nil!")
			continue
		}
		go bp.Start(Conf.BotInfos)
	}
}

func connCQServer(bCtx *BotContext, try int) {
	header := make(http.Header)
	header.Add("Authorization", TokenPrefix + bCtx.BotInfo.Token)
	header.Add("Content-Type", "application/json; charset=utf-8")
	var (
		conn *ws.Conn
		err  error
	)
	for i := 0; i < try; i++ {
		LBLogger.WithField("BotName", bCtx.BotInfo.Name).Printf("尝试连接第%v次", i+1)
		conn, _, err = ws.DefaultDialer.Dial("ws://"+bCtx.BotInfo.Host+":"+strconv.Itoa(bCtx.BotInfo.Port), header)
		if err != nil {
			LBLogger.WithField("BotName", bCtx.BotInfo.Name).Warnln("连接CQ server 失败。将在3秒后重试。")
			time.Sleep(time.Second * 3)
			continue
		}
		bCtx.Conn = conn
		bCtx.IsReady = true
		for _, hook := range onConnectChain {
			err = hook(*bCtx.BotInfo)
			if err != nil {
				LBLogger.WithField("BotName", bCtx.BotInfo.Name).Infoln(err)
				closeConn(bCtx)
				return
			}
		}
		LBLogger.WithField("BotName", bCtx.BotInfo.Name).Infoln("Bot已经上线。")
		go receiveData(bCtx)
		go sendData(bCtx)
		break
	}
}

const (
	DefaultTimeout = 10
	OffHeartCheck  = 0
)

func heartCheck(bCtx *BotContext) {
	if bCtx.BotInfo.Timeout == OffHeartCheck {
		return
	}
	timeout := bCtx.BotInfo.Timeout
	if timeout < DefaultTimeout {
		timeout = DefaultTimeout
	}
	LBLogger.WithField("BotName", bCtx.BotInfo.Name).Infoln("将在15秒之后开启心跳检测，超时时间为：", timeout)
	time.Sleep(time.Second * 15)
	for {
		select {
		case flag := <-bCtx.FlagChan:
			{
				if flag == heartOn {
					continue
				} else if flag == idMismatch {
					closeConn(bCtx)
					LBLogger.WithField("BotName", bCtx.BotInfo.Name).Warnln("已禁用该Bot")
					return
				}
			}
		case <-time.After(time.Second * time.Duration(timeout)):
			LBLogger.WithField("BotName", bCtx.BotInfo.Name).Infoln("Bot心跳检测失败，尝试重新连接")
			// 关闭原来的连接
			closeConn(bCtx)
			connCQServer(bCtx, ReconnTimes)
			break
		}
	}
}

const (
	dataFormatError = -1
	dataTypeEvent   = 1
	dataTypeResp    = 2
)

func parseData(data []byte, msgType string) (interface{}, int, error) {
	var e Event
	switch msgType {
	case MsgTypeArray:
		e.Message = make([]MsgSeg, 0)
	case MsgTypeString:
		e.Message = ""
	}
	err := json.Unmarshal(data, &e)
	if err != nil || e.PostType == "" || e.Time == 0 {
		var resp ApiResp
		err = json.Unmarshal(data, &resp)
		if err != nil {
			return nil, dataFormatError, errors.New("读入数据解析失败，数据格式异常。")
		}
		return &resp, dataTypeResp, nil
	}
	return &e, dataTypeEvent, nil
}

func receiveData(bCtx *BotContext) {
	for {
		var (
			err  error
			data []byte
		)
		if !bCtx.IsReady {
			break
		}
		_, data, err = bCtx.Conn.ReadMessage()
		if err != nil {
			LBLogger.WithField("BotName", bCtx.BotInfo.Name).Debugln("Bot消息读取异常")
			closeConn(bCtx)
			break
		}
		if len(data) == 0 {
			continue
		}
		var (
			result interface{}
			dt     int
		)
		result, dt, err = parseData(data, bCtx.BotInfo.MessageType)
		if err != nil {
			LBLogger.WithField("BotName", bCtx.BotInfo.Name).WithField("Data", string(data)).Warningln(err)
			continue
		}
		switch dt {
		case dataTypeEvent:
			e, hookPass := (result.(*Event)), true
			// LBLogger.Debugln("receive msg: ", *e)
			for _, hook := range eventInChain {
				err = hook(e, *bCtx.BotInfo)
				if err != nil {
					LBLogger.WithField("BotName", bCtx.BotInfo.Name).Infoln(err)
					hookPass = false
					break
				}
			}
			if !hookPass {
				break
			}
			eCtx := eventContext{
				e:    e,
				bCtx: bCtx,
			}
			select {
			case cqEventChan <- eCtx:
			case <-bCtx.CloseChan:
				return
			}
		case dataTypeResp:
			rCtx := apiRespContext{
				resp: result.(*ApiResp),
				bCtx: bCtx,
			}
			select {
			case cqRespChan <- rCtx:
			case <-bCtx.CloseChan:
				return
			}
		}
	}
}

func sendData(bCtx *BotContext) {
	for {
		if !bCtx.IsReady {
			break
		}
		var (
			hookPass = true
			err      error
		)
		api := <-bCtx.OutChan
		for _, hook := range beforeApiOutChain {
			err = hook(&api, *bCtx.BotInfo)
			if err != nil {
				LBLogger.WithField("BotName", bCtx.BotInfo.Name).Infoln(err)
				hookPass = false
				break
			}
		}
		if !hookPass {
			continue
		}
		err = bCtx.Conn.WriteJSON(api)
		if err != nil {
			LBLogger.WithField("BotName", bCtx.BotInfo.Name).Debugln("Bot消息发送异常")
			closeConn(bCtx)
			bCtx.OutChan <- api
			break
		}
	}
}

func closeConn(bCtx *BotContext) {
	bCtx.CloseLock.Lock()
	defer bCtx.CloseLock.Unlock()
	if !bCtx.IsReady {
		return
	}
	LBLogger.WithField("BotName", bCtx.BotInfo.Name).Debugln("正在尝试关闭已有连接")
	err := bCtx.Conn.Close()
	if err != nil {
		LBLogger.WithField("BotName", bCtx.BotInfo.Name).Debugln("关闭连接异常！尚存在数据未读取", err)
	}
	bCtx.IsReady = false
	bCtx.Conn = nil
	for _, hook := range disConnectChain {
		hook(*bCtx.BotInfo)
	}
}

func processMateEvent(e *Event, bCtx *BotContext) {
	switch e.MetaEventType {
	case Lifecycle:
		if bCtx.BotInfo.BotID != e.SelfID {
			le := LBLogger.WithField("BotName", bCtx.BotInfo.Name).WithField("BotId", bCtx.BotInfo.BotID).WithField("CQ-Server-Id", e.SelfID)
			le.Warnln("CQ Server 账号ID与所配置的ID无法匹配，即将禁用该Bot")
			bCtx.FlagChan <- idMismatch
		}
	case Heartbeat:
		bCtx.FlagChan <- heartOn
	}
}
