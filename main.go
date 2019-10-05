package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	listen  = flag.String("listen", "127.0.0.1:8081", "local listen port")
	targets = flag.String("targets", "", "forward target list: ip:port,domian:port")

	manager = &remoteManager{}
)

func main() {
	flag.Parse()
	if *listen == "" {
		log.Fatal("listen address cant not empty")
	}
	if *targets == "" {
		log.Fatal("forward target list can not empty")
	}
	log.Fatal(startServer(*listen, *targets))
}

func startServer(listen, targets string) error {
	manager.Init(targets)

	ln, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("listen on %s failed: %s", listen, err)
	}

	log.Println("listen on ", listen)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("accept connect from %s failed: %s", ln.Addr(), err)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("accept from ", conn.RemoteAddr())

	var r net.Conn
	for {
		target, err := manager.Pick()
		if err != nil {
			log.Println("[ERROR]pick target failed: ", err)
			return
		}
		log.Println("target >>", target)

		startTime := time.Now()
		r, err = net.DialTimeout("tcp", target.Addr, time.Second*3)
		if err != nil {
			log.Printf("dial remote %s failed: %s\n", target.Addr, err)
			manager.SetDown(target.ID)
			continue
		}
		dialDuration := time.Now().Sub(startTime)
		log.Println("set dialDuration ", target.Addr, " to ", dialDuration)
		manager.UpdateDialDuration(target.ID, dialDuration)

		log.Println("use ", target.Addr)
		break
	}

	desc := fmt.Sprintf("%s -> %s", conn.RemoteAddr(), r.RemoteAddr())

	defer r.Close()
	log.Println(desc, "connected")
	r.(*net.TCPConn).SetKeepAlive(true)
	r.(*net.TCPConn).SetNoDelay(true)

	inChan := make(chan int, 1)
	outChan := make(chan int, 1)

	go relayConn(conn, r, outChan)
	go relayConn(r, conn, inChan)

	select {
	case <-inChan:
		log.Println(desc, "in closed")
	case <-outChan:
		log.Println(desc, "out closed")
	}
	log.Println(desc, "disconnected ")
}

func relayConn(in, out net.Conn, ch chan int) {
	defer func() {
		ch <- 1
	}()
	_, err := io.Copy(out, in)
	if err != nil {
		return
	}
}

type remoteManager struct {
	list map[int]remoteInfo
	sync.Mutex
}

func (m *remoteManager) Init(remotes string) {
	m.Lock()
	defer m.Unlock()
	m.list = make(map[int]remoteInfo)
	for i, target := range strings.Split(remotes, ",") {
		m.list[i] = remoteInfo{
			ID:           i,
			Addr:         target,
			DialDuration: 0,
			IsDown:       false,
		}
	}
}

func (m *remoteManager) Pick() (remoteInfo, error) {
	m.Lock()
	defer m.Unlock()

	list := make([]remoteInfo, 0)

	retryTimeLimit := time.Now().Add(-time.Minute)

	for _, item := range m.list {
		if item.IsDown {
			if item.LastTry.After(retryTimeLimit) {
				continue
			}
			//reset
			item.IsDown = false
			item.DialDuration = 0
			m.list[item.ID] = item
		}
		list = append(list, item)
	}

	if len(list) < 1 {
		return remoteInfo{}, errors.New("All remote check failed")
	}

	ordered := remoteInfoList(list)
	sort.Sort(ordered)
	return ordered[0], nil
}

func (m *remoteManager) SetDown(ID int) {
	m.Lock()
	defer m.Unlock()
	target, _ := m.list[ID]
	target.IsDown = true
	target.LastTry = time.Now()
	m.list[ID] = target
}

func (m *remoteManager) UpdateDialDuration(ID int, duration time.Duration) {
	m.Lock()
	defer m.Unlock()
	target, _ := m.list[ID]
	target.IsDown = false
	target.LastTry = time.Now()
	target.DialDuration = duration
	m.list[ID] = target
}

type remoteInfo struct {
	ID           int
	Addr         string
	DialDuration time.Duration
	IsDown       bool
	LastTry      time.Time
}

type remoteInfoList []remoteInfo

func (r remoteInfoList) Len() int {
	return len(r)
}

func (r remoteInfoList) Less(i, j int) bool {
	if r[i].IsDown == r[j].IsDown {
		return r[i].DialDuration < r[j].DialDuration
	}
	return !r[i].IsDown
}
func (r remoteInfoList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
