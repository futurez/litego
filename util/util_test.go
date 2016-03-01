// util_test.go
package util

import (
	"testing"
)

func TestRandomNumber(t *testing.T) {
	t.Log("first  random number :", Rand())
	t.Log("second random number :", Rand())
}

func TestGetUUID(t *testing.T) {
	t.Log("first  UUID :", GetUUID())
	t.Log("second UUID :", GetUUID())
}
