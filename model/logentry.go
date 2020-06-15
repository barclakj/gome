package model

type LogEntry struct {
	Data []byte
	Origin string
	Seq uint64
	Uuid string
	Ts uint64
	RemoteTs uint64
}

type Log struct {
	Entry *LogEntry
	Next *Log
}

