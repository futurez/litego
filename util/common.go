package util

import (
	"crypto/md5"
	"encoding/hex"
	"os"

	"github.com/zhoufuture/golite/logger"
)

func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func CheckError(err error) {
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
