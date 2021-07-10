package main

import (
	"github.com/ABiao0306/luxtbot"
)

func main() {
	luxtbot.MakeBeforeApiOutHook(func(apiPost *luxtbot.ApiPost, bInfo luxtbot.BotInfo) error {
		if apiPost.Action == luxtbot.GroupMsgAction {
			var groupMsg = (apiPost.Params).(luxtbot.GroupMsg)
			groupMsg.GroupID = 123
		}
		return nil
	}).AddToHookChain()
	luxtbot.InitDefaultPluginManager(0)
	luxtbot.Init("config-file-path.yml")
	luxtbot.Start()
	select {}
}
