// util_test.go
package util

import (
	"testing"
)

func TestRandomNumber(t *testing.T) {
	t.Log("1 random number :", Rand())
	t.Log("2 random number :", Rand())
}

func TestUUID(t *testing.T) {
	t.Log("1 UUID :", UUID())
	t.Log("2 UUID :", UUID())
}

func TestGetIntranetIp(t *testing.T) {
	t.Log("Local ip :", GetIntranetIP())
}

func TestChechEmail(t *testing.T) {
	t.Log("ChechEmail : ", ChechEmail("abc@adc.com.cn"))
	t.Log("ChechEmail : ", ChechEmail("abc@126.com"))
	t.Log("ChechEmail : ", ChechEmail("abc@126.com.cn"))
}
