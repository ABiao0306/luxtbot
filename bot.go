package luxtbot

import (
	"errors"
	"io/ioutil"
	"os"
	"sync"

	ws "github.com/gorilla/websocket"
	yaml "gopkg.in/yaml.v2"
)

var (
	Conf Config
	bots BotCtxs
)

type Config struct {
	BotInfos         []BotInfo `yaml:"bots"`
	LogConf          LogConf   `yaml:"log"`
	SAdmins          []int64   `yaml:"s-admin"`
	CallbackPoolSize int       `yaml:"callback-pool-size"`
}

type BotCtxs = []BotContext

type BotInfo struct {
	BotID       int64   `yaml:"id"`
	Host        string  `yaml:"host"`
	Port        int     `yaml:"port"`
	Name        string  `yaml:"name"`
	Token       string  `yaml:"access-token"`
	Timeout     int     `yaml:"time-out"`
	Admins      []int64 `yaml:"admins"`
	MessageType string  `yaml:"message-type"`
}

type BotContext struct {
	Conn      *ws.Conn
	CloseLock *sync.Mutex
	OutChan   chan ApiPost
	FlagChan  chan byte
	CloseChan chan byte
	IsReady   bool
	BotInfo   *BotInfo
}

func Init(conf string) {
	var (
		file *os.File
		err  error
		data []byte
	)
	file, err = os.Open(conf)
	if err != nil {
		panic("打开配置文件失败：" + conf)
	}
	data, err = ioutil.ReadAll(file)
	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		panic("打开配置文件失败：" + conf)
	}
	InitLogConf()
	InitPluginList()
	InitBotCtxs()
}

func Start() {
	RunBots()
	RunEventDispatcher()
	RunRespDispatcher(Conf.CallbackPoolSize)
	RunBackenPlugin()
	select {}
}

func getBotCtxByID(botID int64) (*BotContext, error) {
	for _, bCtx := range bots {
		if bCtx.BotInfo.BotID == botID {
			return &bCtx, nil
		}
	}
	return nil, errors.New("无法找到该ID的Bot实例。")
}
