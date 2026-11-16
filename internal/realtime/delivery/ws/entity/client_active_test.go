package entity

import (
	"testing"
	"time"
)

func TestTouchOrdersClientsByRecency(t *testing.T) {
	a, b := &Client{}, &Client{}
	a.Touch()
	time.Sleep(2 * time.Millisecond)
	b.Touch()
	if !(b.LastActiveUnixNano() > a.LastActiveUnixNano()) {
		t.Fatal("the later Touch must have the larger timestamp")
	}
}
