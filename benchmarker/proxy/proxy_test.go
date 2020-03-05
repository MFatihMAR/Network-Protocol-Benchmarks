package proxy

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func assert(t *testing.T, cond bool, msg string) {
	if cond {
		t.Fatal(msg)
	}
}

func assertf(t *testing.T, cond bool, msg string, args ...interface{}) {
	assert(t, cond, fmt.Sprintf(msg, args...))
}

func TestNonProxy(t *testing.T) {
	northAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9696")
	assertf(t, err != nil, "cannot resolve north addr -> %s", err)
	southAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6969")
	assertf(t, err != nil, "cannot resolve south addr -> %s", err)

	northSock, err := net.DialUDP("udp", northAddr, southAddr)
	assertf(t, err != nil, "cannot create north socket -> %s", err)
	defer northSock.Close()
	southSock, err := net.DialUDP("udp", southAddr, northAddr)
	assertf(t, err != nil, "cannot create south socket -> %s", err)
	defer southSock.Close()

	northRecvCh := make(chan []byte, 1024)
	assert(t, northRecvCh == nil, "cannot create north receive channel")
	defer close(northRecvCh)
	southRecvCh := make(chan []byte, 1024)
	assert(t, southRecvCh == nil, "cannot create south receive channel")
	defer close(southRecvCh)

	run := true
	defer func() { run = false }()
	msgCount := 128

	recvFunc := func(sock *net.UDPConn, recvCh chan []byte) {
		buf := make([]byte, 1500)
		for run {
			len, err := sock.Read(buf)
			if !run {
				return
			}
			assertf(t, err != nil, "failed to read from socket -> %s", err)

			pkt := make([]byte, len)
			copy(pkt, buf)
			recvCh <- pkt
		}
	}
	go recvFunc(northSock, northRecvCh)
	go recvFunc(southSock, southRecvCh)

	sendFunc := func(sock *net.UDPConn, addr *net.UDPAddr, count int) {
		for idx := 0; run == true && idx < count; idx++ {
			b := []byte(fmt.Sprintf("hello from the other side -> %d", idx))
			s := len(b)

			w, err := sock.Write(b)
			if !run {
				return
			}
			assertf(t, err != nil, "failed to write to socket -> %s", err)
			assertf(t, w != s, "cannot send entire payload -> packet: %d / wrote: %d", s, w)
		}
	}
	go sendFunc(northSock, southAddr, msgCount)
	go sendFunc(southSock, northAddr, msgCount)

	startTime := time.Now()
	for idx := 0; idx < msgCount*2; {
		select {
		case northMsg := <-northRecvCh:
			nMsg := string(northMsg)
			assertf(t,
				strings.HasPrefix(nMsg, "hello from the other side ->") == false,
				"unexpected packet read from north channel -> %s", nMsg)
			idx++
		case southMsg := <-southRecvCh:
			sMsg := string(southMsg)
			assertf(t,
				strings.HasPrefix(sMsg, "hello from the other side ->") == false,
				"unexpected packet read from south channel -> %s", sMsg)
			idx++
		default:
			if time.Since(startTime) > time.Second*10 {
				assertf(t, true, "operation timed out - either north or south socket sent less packets than expected -> idx: %d", idx)
			} else {
				assert(t, t.Failed(), "one of the goroutines failed")
			}
		}
	}
}

func TestProxy(t *testing.T) {
	proxyAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2020")
	assertf(t, err != nil, "cannot resolve proxy addr -> %s", err)
	northAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9696")
	assertf(t, err != nil, "cannot resolve north addr -> %s", err)
	southAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6969")
	assertf(t, err != nil, "cannot resolve south addr -> %s", err)

	northSock, err := net.DialUDP("udp", northAddr, proxyAddr)
	assertf(t, err != nil, "cannot create north socket -> %s", err)
	defer northSock.Close()
	southSock, err := net.DialUDP("udp", southAddr, proxyAddr)
	assertf(t, err != nil, "cannot create south socket -> %s", err)
	defer southSock.Close()
	proxy, err := NewProxy(&Config{
		ProxyPort: uint16(proxyAddr.Port),
		NorthPort: uint16(northAddr.Port),
		SouthPort: uint16(southAddr.Port),

		SockBufSize: 1500,
		ChanBufLen:  0xC0DE,
	})
	assertf(t, err != nil, "cannot create proxy -> %s", err)
	assert(t, proxy == nil, "no error but proxy is nil")
	defer proxy.Close()

	northRecvCh := make(chan []byte, 1024)
	assert(t, northRecvCh == nil, "cannot create north receive channel")
	defer close(northRecvCh)
	southRecvCh := make(chan []byte, 1024)
	assert(t, southRecvCh == nil, "cannot create south receive channel")
	defer close(southRecvCh)

	run := true
	defer func() { run = false }()
	msgCount := 128

	recvFunc := func(sock *net.UDPConn, recvCh chan []byte) {
		buf := make([]byte, 1500)
		for run {
			len, err := sock.Read(buf)
			if !run {
				return
			}
			assertf(t, err != nil, "failed to read from socket -> %s", err)

			pkt := make([]byte, len)
			copy(pkt, buf)
			recvCh <- pkt
		}
	}
	go recvFunc(northSock, northRecvCh)
	go recvFunc(southSock, southRecvCh)

	sendFunc := func(sock *net.UDPConn, addr *net.UDPAddr, count int) {
		for idx := 0; run == true && idx < count; idx++ {
			b := []byte(fmt.Sprintf("hello from the other side -> %d", idx))
			s := len(b)

			w, err := sock.Write(b)
			if !run {
				return
			}
			assertf(t, err != nil, "failed to write to socket -> %s", err)
			assertf(t, w != s, "cannot send entire payload -> packet: %d / wrote: %d", s, w)
		}
	}
	go sendFunc(northSock, southAddr, msgCount)
	go sendFunc(southSock, northAddr, msgCount)

	go func() {
		for run {
			select {
			case northPkt := <-proxy.NorthRecvCh:
				proxy.SouthSendCh <- northPkt
			case southPkt := <-proxy.SouthRecvCh:
				proxy.NorthSendCh <- southPkt
			}
		}
	}()

	startTime := time.Now()
	for idx := 0; idx < msgCount*2; {
		select {
		case northMsg := <-northRecvCh:
			nMsg := string(northMsg)
			assertf(t,
				strings.HasPrefix(nMsg, "hello from the other side ->") == false,
				"unexpected packet read from north channel -> %s", nMsg)
			idx++
		case southMsg := <-southRecvCh:
			sMsg := string(southMsg)
			assertf(t,
				strings.HasPrefix(sMsg, "hello from the other side ->") == false,
				"unexpected packet read from south channel -> %s", sMsg)
			idx++
		default:
			if time.Since(startTime) > time.Second*10 {
				assertf(t, true, "operation timed out but task is not completed -> idx: %d", idx)
			} else {
				assert(t, t.Failed(), "one of the goroutines failed")
			}
		}
	}
}
