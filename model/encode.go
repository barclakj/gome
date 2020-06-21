package model

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(oid string, previousHash string, data []byte) string {
	hasher := sha256.New()

	hashableData := []byte(oid + ":" + previousHash)
	hashableData = append(hashableData, data...)

	hasher.Write(hashableData)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}
