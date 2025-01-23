package netreactors

import "log"

type Log struct {
	IsOpenDebugLog bool
}

var Dlog = &Log{
	IsOpenDebugLog: false,
}

// var IsOpenDebugLog bool = false

func (d *Log) Printf(format string, v ...any) {
	if d.IsOpenDebugLog {
		log.Printf(format, v...)
	}
}

func (d *Log) TurnOnLog() {
	Dlog.IsOpenDebugLog = true
}

func (d *Log) TurnOffLog() {
	Dlog.IsOpenDebugLog = false
}
