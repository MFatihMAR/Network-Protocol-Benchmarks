package proxy

import "net"

type Proxy struct {
	LastError error

	FakeRecvCh chan []byte
	FakeSendCh chan []byte

	RealRecvCh chan []byte
	RealSendCh chan []byte

	fakeRecvBuf []byte
	realRecvBuf []byte

	conn *net.UDPConn
}

func NewProxy(fakePort, realPort uint16, bufferSize uint16) (*Proxy, error) {
	// todo: check `fakePort`, `realPort` and `bufferSize` arguments

	return nil, nil
}

func (p *Proxy) Close() {
}
