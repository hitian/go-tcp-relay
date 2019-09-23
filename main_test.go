package main

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

func TestRemoteInfoListOrder(t *testing.T) {
	now := time.Now()
	list := []remoteInfo{
		remoteInfo{
			Addr:    "1",
			IsDown:  true,
			LastTry: now.Add(time.Minute * 2),
		},
		remoteInfo{
			Addr:         "2",
			IsDown:       false,
			DialDuration: time.Microsecond * 20,
			LastTry:      now.Add(time.Minute * 2),
		},
		remoteInfo{
			Addr:         "3",
			IsDown:       false,
			DialDuration: time.Microsecond * 10,
			LastTry:      now.Add(time.Minute * 2),
		},
	}

	orders := remoteInfoList(list)
	sort.Sort(orders)
	fmt.Println(orders)
	if orders[0].Addr != "3" {
		fmt.Println("order by Isdown and DialDuration failed")
		t.FailNow()
	}
}
