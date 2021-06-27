package main

import "luxtbot"

func main() {
	luxtbot.Init("../dev/config.yml")
	luxtbot.Start()
	select {}
}
