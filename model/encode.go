package model

func Encode(msg LogEntry) []byte {
	return msg.Data
}

