package util

import (
	"math/rand"
	"time"
)

func init() {
	timens := int64(time.Now().Nanosecond())
	rand.Seed(timens)
}

func Rand() uint32 {
	return rand.Uint32()
}
