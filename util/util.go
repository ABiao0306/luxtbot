package util

import (
	"github.com/rs/xid"
)

func GetEchoStr() string {
	echo := xid.New()
	return echo.String()
}

func Trim(s string) string {
	for i, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		return s[i:]
	}
	return s
}
