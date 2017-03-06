package myset

import (
	"testing"
)

func TestSet(t *testing.T){
	s := New()
	s.Add(1)
	s.Add(1)
	s.Add(0)
	s.Add(2)
	s.Add(4)
	s.Add(3)
	s.Clear()
	if s.IsEmpty() {
		t.Log("0 item")
	}
	s.Add(1)
	s.Add(2)
	s.Add(3)
	if s.Has(2) {
		t.Log("2 does exist")
	}
	s.Remove(2)
	s.Remove(3)
	t.Log("无序的切片", s.List())
	t.Log("有序的切片", s.SortList())
}
