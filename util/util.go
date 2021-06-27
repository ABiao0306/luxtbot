package util

import (
	"strconv"
	"time"
)

func RunTimedTask(dura time.Duration, task func(...interface{}), args ...interface{}) {
	ticker := time.NewTicker(dura)
	go func() {
		for _ = range ticker.C {
			task(args...)
		}
	}()
}

func GetEchoStr() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
