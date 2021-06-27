package luxtbot

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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
)

type eventContext struct {
	e    *Event
	bCtx *BotContext
}

type apiRespContext struct {
	resp *ApiResp
	bCtx *BotContext
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
		go connCQServer(&bCtx, ReconnTimes)
		go heartCheck(&bCtx)
	}
}

func InitEventDispatcher() {
	for {
		eCtx := <-cqEventChan
		e, bCtx := eCtx.e, eCtx.bCtx
		switch e.PostType {
		case MessageEvent:
			for _, mp := range MsgChain {
				if !mp.Plg.Enable || !mp.Plg.CheckRules(e, *bCtx.BotInfo) {
					continue
				}
				mp.Process(e, *bCtx.BotInfo)
			}
			for _, cp := range CmdChain {
				if !cp.Plg.Enable || !cp.Plg.CheckRules(e, *bCtx.BotInfo) {
					continue
				}
				cp.Process(e, *bCtx.BotInfo)
			}
		case NoticeEvent:
			for _, np := range NoticeChain {
				if !np.Plg.Enable || !np.Plg.CheckRules(e, *bCtx.BotInfo) {
					continue
				}
				np.Process(e, *bCtx.BotInfo)
			}
		case RequestEvent:
			for _, mp := range MsgChain {
				if !mp.Plg.Enable || !mp.Plg.CheckRules(e, *bCtx.BotInfo) {
					continue
				}
				mp.Process(e, *bCtx.BotInfo)
			}
		case MetaEvent:
			processMateEvent(e, bCtx)
		}
	}
}

func InitBackenPlugin() {
	for _, bp := range BackenChain {
		bp.Init()
		go bp.Start(bots)
	}
}

func connCQServer(bCtx *BotContext, try int) {
	header := make(http.Header)
	header.Add("Authorization", "dasdasddasdasd Ym90LXNhaWtv")
	header.Add("Content-Type", "application/json; charset=utf-8")
	var (
		conn *ws.Conn
		err  error
	)
	for i := 0; i < try; i++ {
		logrus.WithField("BotName", bCtx.BotInfo.Name).Printf("尝试连接第%v次", i+1)
		conn, _, err = ws.DefaultDialer.Dial("ws://"+bCtx.BotInfo.Host+":"+strconv.Itoa(bCtx.BotInfo.Port), header)
		if err != nil {
			logrus.WithField("BotName", bCtx.BotInfo.Name).Warnln("连接CQ server 失败。将在3秒后重试。")
			time.Sleep(time.Second * 3)
			continue
		}
		bCtx.Conn = conn
		bCtx.IsReady = true
		logrus.WithField("BotName", bCtx.BotInfo.Name).Infoln("Bot已经上线。")
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
	logrus.WithField("BotName", bCtx.BotInfo.Name).Warnln("将在15秒之后开启心跳检测，超时时间为：", timeout)
	time.Sleep(time.Second * 15)
	for {
		select {
		case flag := <-bCtx.FlagChan:
			{
				if flag == heartOn {
					continue
				} else if flag == idMismatch {
					closeConn(bCtx)
					logrus.WithField("BotName", bCtx.BotInfo.Name).Warnln("已禁用该Bot")
					return
				}
			}
		case <-time.After(time.Second * time.Duration(timeout)):
			logrus.WithField("BotName", bCtx.BotInfo.Name).Infoln("Bot心跳检测失败，尝试重新连接")
			// 关闭原来的连接
			closeConn(bCtx)
			connCQServer(bCtx, ReconnTimes)
			break
		}
	}
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
			logrus.WithField("BotName", bCtx.BotInfo.Name).Debugln("Bot消息读取异常")
			closeConn(bCtx)
			break
		}
		if len(data) == 0 {
			continue
		}
		// logrus.WithField("BotName", bCtx.BotInfo.Name).Debugln("recieve: ", string(data))
		var e Event
		_ = json.Unmarshal(data, &e)
		eCtx := eventContext{
			e:    &e,
			bCtx: bCtx,
		}
		select {
		case cqEventChan <- eCtx:
		case <-bCtx.CloseChan:
			return
		}
	}
}

func sendData(bCtx *BotContext) {
	for {
		if !bCtx.IsReady {
			break
		}
		api := <-bCtx.OutChan
		err := bCtx.Conn.WriteJSON(api)
		if err != nil {
			logrus.WithField("BotName", bCtx.BotInfo.Name).Debugln("Bot消息发送异常")
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
	logrus.WithField("BotName", bCtx.BotInfo.Name).Debugln("正在尝试关闭已有连接")
	err := bCtx.Conn.Close()
	if err != nil {
		logrus.WithField("BotName", bCtx.BotInfo.Name).Debugln("关闭连接异常！尚存在数据未读取", err)
	}
	bCtx.IsReady = false
	bCtx.Conn = nil
}

func processMateEvent(e *Event, bCtx *BotContext) {
	switch e.MetaEventType {
	case Lifecycle:
		if bCtx.BotInfo.BotID != e.SelfID {
			le := logrus.WithField("BotName", bCtx.BotInfo.Name).WithField("BotId", bCtx.BotInfo.BotID).WithField("CQ-Server-Id", e.SelfID)
			le.Warnln("CQ Server 账号ID与所配置的ID无法匹配，即将禁用该Bot")
			bCtx.FlagChan <- idMismatch
		}
	case Heartbeat:
		bCtx.FlagChan <- heartOn
	}
}
