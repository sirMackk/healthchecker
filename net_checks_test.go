package healthchecker

import (
	"net"
	"testing"
	"time"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type fakePacketConn struct {
	InputBuf  []byte
	OutputBuf []byte
	Addr      net.Addr
}

func (f fakePacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n = copy(p, f.OutputBuf)
	return n, f.Addr, nil
}

func (f fakePacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n = copy(f.InputBuf, p)
	return n, nil
}

func (f fakePacketConn) LocalAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1")
	return addr
}

func (f fakePacketConn) Close() error                       { return nil }
func (f fakePacketConn) SetDeadline(t time.Time) error      { return nil }
func (f fakePacketConn) SetReadDeadline(t time.Time) error  { return nil }
func (f fakePacketConn) SetWriteDeadline(t time.Time) error { return nil }
func (f fakePacketConn) IPv4PacketConn() *ipv4.PacketConn   { return nil }
func (f fakePacketConn) IPv6PacketConn() *ipv6.PacketConn   { return nil }

func NewFakePacketConn() ICMPPacketConn {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1")
	return &fakePacketConn{
		InputBuf:  make([]byte, 32),
		OutputBuf: nil,
		Addr:      addr,
	}
}

func TestICMPV4Check(t *testing.T) {
	OutputMsgs := []struct {
		result ResultCode
		outbuf []byte
	}{
		{Success, []byte{
			0x0, 0x0, 0xb7, 0xd2, 0x3, 0xe9, 0x0, 0x1, 0x70, 0x69, 0x6e, 0x67, 0x65, 0x72}},
		{Failure, []byte{
			0xFF, 0x65, 0x72}},
		{Failure, []byte{
			0x8, 0x0, 0xaf, 0xd2, 0x3, 0xe9, 0x0, 0x1, 0x70, 0x69, 0x6e, 0x67, 0x65, 0x72}},
	}

	for _, tt := range OutputMsgs {
		t.Run(string(tt.result), func(t *testing.T) {
			fakePktConn := NewFakePacketConn()
			fakePktConn.(*fakePacketConn).OutputBuf = tt.outbuf
			checker := &ICMPChecker{
				Conn: fakePktConn,
			}

			checkFunc, _ := checker.NewICMPV4Check(map[string]string{"targetIP": "localhost"})
			res := <-checkFunc()

			if res.Result != tt.result {
				t.Errorf(
					"Got: %#v, Wanted: %#v\nInputBuf: %#v",
					res.Result,
					tt.result,
					fakePktConn.(*fakePacketConn).InputBuf,
				)
			}
		})
	}
}
