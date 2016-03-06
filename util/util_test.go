// util_test.go
package util

import (
	"testing"
)

func TestRandomNumber(t *testing.T) {
	t.Log("1 random number :", Rand())
	t.Log("2 random number :", Rand())
}

func TestGetUUID(t *testing.T) {
	t.Log("1 UUID :", GetUUID())
	t.Log("2 UUID :", GetUUID())
}

func TestGetIntranetIp(t *testing.T) {
	t.Log("Local ip :", GetIntranetIP())
}
