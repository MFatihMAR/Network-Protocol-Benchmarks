package proxy

import "net"

type Proxy struct {
	NorthRecvCh chan []byte
	NorthSendCh chan []byte
	SouthRecvCh chan []byte
	SouthSendCh chan []byte

	Err error

	conn *net.UDPConn
	buf  []byte
	ok   bool
}

type Config struct {
	ListenPort uint16
	NorthPort  uint16
	SouthPort  uint16

	SockBufSize uint16
	ChanBufLen  uint16
}

func NewProxy(config *Config) (*Proxy, error) {
	// todo: check `config` arguments

	return nil, nil
}

func (p *Proxy) Close() {
}
