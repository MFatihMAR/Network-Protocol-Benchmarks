package proxy

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type Proxy struct {
	NorthRecvCh chan []byte
	NorthSendCh chan []byte
	SouthRecvCh chan []byte
	SouthSendCh chan []byte

	Err error

	northAddr *net.UDPAddr
	southAddr *net.UDPAddr

	buf  []byte
	conn *net.UDPConn

	closeOnce sync.Once
	closed    bool
}

type Config struct {
	ProxyPort uint16
	NorthPort uint16
	SouthPort uint16

	SockBufSize uint16
	ChanBufLen  uint16
}

func NewProxy(cfg *Config) (*Proxy, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if cfg.ProxyPort == 0 {
		return nil, errors.New("proxy port is required")
	}
	if cfg.NorthPort == 0 {
		return nil, errors.New("north port is required")
	}
	if cfg.SouthPort == 0 {
		return nil, errors.New("south port is required")
	}
	if cfg.ProxyPort == cfg.NorthPort ||
		cfg.ProxyPort == cfg.SouthPort ||
		cfg.NorthPort == cfg.SouthPort {
		return nil, errors.New("all ports must be different")
	}
	if cfg.SockBufSize < 68 {
		return nil, errors.New("socket buffer cannot be smaller than 68 (minimum MTU size)")
	}
	if cfg.ChanBufLen < 64 {
		return nil, errors.New("channels should buffer at least 64 packets")
	}

	proxyAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", cfg.ProxyPort))
	if err != nil {
		return nil, err
	}
	northAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", cfg.NorthPort))
	if err != nil {
		return nil, err
	}
	southAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", cfg.SouthPort))
	if err != nil {
		return nil, err
	}

	ok := false
	p := Proxy{
		NorthRecvCh: make(chan []byte, cfg.ChanBufLen),
		NorthSendCh: make(chan []byte, cfg.ChanBufLen),
		SouthRecvCh: make(chan []byte, cfg.ChanBufLen),
		SouthSendCh: make(chan []byte, cfg.ChanBufLen),

		northAddr: northAddr,
		southAddr: southAddr,

		buf:  make([]byte, cfg.SockBufSize),
		conn: nil,

		closed: false,
	}
	defer func() {
		if !ok {
			p.Close()
		}
	}()

	p.conn, err = net.ListenUDP("udp", proxyAddr)
	if err != nil {
		return nil, err
	}

	go p.recvRoutine()
	go p.sendRoutine()

	ok = true
	return &p, nil
}

func (p *Proxy) recvRoutine() {
	for !p.closed {
		len, addr, err := p.conn.ReadFromUDP(p.buf)
		if p.closed {
			return
		}
		if err != nil {
			p.Err = err
			p.Close()
			return
		}

		pkt := make([]byte, len)
		copy(pkt, p.buf)

		if addr.Port == p.northAddr.Port {
			p.NorthRecvCh <- pkt
		} else if addr.Port == p.southAddr.Port {
			p.SouthRecvCh <- pkt
		} else {
			// unexpected data from arbitrary port is ignored
		}
	}
}

func (p *Proxy) sendRoutine() {
	for !p.closed {
		select {
		case nPkt, nOk := <-p.NorthSendCh:
			if p.closed {
				return
			}
			if !nOk {
				p.Err = errors.New("north send channel is closed")
				p.Close()
				return
			} else {
				_, err := p.conn.WriteToUDP(nPkt, p.northAddr)
				if err != nil {
					p.Err = err
					p.Close()
					return
				}
			}
		case sPkt, sOk := <-p.SouthSendCh:
			if p.closed {
				return
			}
			if !sOk {
				p.Err = errors.New("south send channel is closed")
				p.Close()
				return
			} else {
				_, err := p.conn.WriteToUDP(sPkt, p.southAddr)
				if err != nil {
					p.Err = err
					p.Close()
					return
				}
			}
		}
	}
}

func (p *Proxy) Close() {
	p.closeOnce.Do(func() {
		p.closed = true

		close(p.NorthRecvCh)
		close(p.NorthSendCh)
		close(p.SouthRecvCh)
		close(p.SouthSendCh)

		p.conn.Close()
	})
}
