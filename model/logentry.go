package model

type LogEntry struct {
	Data []byte
	Origin string
	Clock uint64
}

type Log struct {
	Entry *LogEntry
	Next *Log
}

