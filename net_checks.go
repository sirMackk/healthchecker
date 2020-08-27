package healthchecker

import (
	"fmt"
	"os"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	log "github.com/sirupsen/logrus"
)

type ICMPPacketConn interface {
	Close() error
	IPv4PacketConn() *ipv4.PacketConn
	IPv6PacketConn() *ipv6.PacketConn
	LocalAddr() net.Addr
	ReadFrom(b []byte) (int, net.Addr, error)
	WriteTo(b []byte, dst net.Addr) (int, error)
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type ICMPChecker struct {
	Conn ICMPPacketConn
}

func NewICMPChecker(timeout time.Duration) (*ICMPChecker, error) {
	_, isSudo := os.LookupEnv("SUDO_COMMAND")
	if !isSudo {
		return nil, fmt.Errorf("If you want to use ICMPChecker, you must run as sudo")
	}
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(time.Second * time.Duration(timeout)))
	checker := ICMPChecker{
		Conn: conn,
	}
	return &checker, nil
}

func (i *ICMPChecker) sendICMPV4Echo(targetIP *net.IPAddr, ua []byte) (*icmp.Message, error) {
	echoReq := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   0,
			Seq:  1,
			Data: ua,
		},
	}
	encodedReq, err := echoReq.Marshal(nil)
	if err != nil {
		log.Debugf("ICMP check failed: couldn't marshal echo req: %s", err)
		return nil, err
	}
	_, err = i.Conn.WriteTo(encodedReq, targetIP)
	if err != nil {
		log.Debugf("ICMP check failed, couldn't write to conn: %s", err)
		return nil, err
	}

	buf := make([]byte, 256)
	bytesRead, _, err := i.Conn.ReadFrom(buf)
	if err != nil {
		log.Debugf("ICMP check failed, couldn't read from conn: %v", err)
		return nil, err
	}

	resp, err := icmp.ParseMessage(1, buf[:bytesRead])
	if err != nil {
		log.Debugf("ICMP check failed, couldn't parse reply: %v - %v", err, buf[:bytesRead])
		return nil, err
	}
	return resp, nil
}

func (i *ICMPChecker) ICMPV4Check(targetIP *net.IPAddr) *CheckResult {
	timeStart := time.Now()
	resp, err := i.sendICMPV4Echo(targetIP, []byte("sirmackk/healthchecker"))
	if err != nil || resp.Type != ipv4.ICMPTypeEchoReply {
		return &CheckResult{
			Timestamp: timeStart,
			Result:    Failure,
			Duration:  time.Since(timeStart),
		}
	}

	return &CheckResult{
		Timestamp: timeStart,
		Result:    Success,
		Duration:  time.Since(timeStart),
	}
}

func (i *ICMPChecker) NewICMPV4Check(args map[string]string) (func() *CheckResult, error) {
	IP, ok := args["targetIP"]
	if !ok {
		return nil, fmt.Errorf("ICMPV4Check missing 'targetIP' parameter")
	}
	targetIP, err := net.ResolveIPAddr("ip4", IP)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse %s into ipv4 address: %s", IP, err)
	}
	return func() *CheckResult {
		return i.ICMPV4Check(targetIP)
	}, nil
}
