package qron

import "log"

const (
	lvlError = iota
	lvlInfo
	lvlDebug
)

var verb int

func SetVerbose(lvl int) {
	verb = lvl
}

func writeLog(lvl int, msg string) {
	if verb >= lvl {
		log.Println(map[int]string{
			lvlError: "[ERR]",
			lvlInfo:  "[INF]",
			lvlDebug: "[DBG]",
		}[lvl], msg)
	}
}
