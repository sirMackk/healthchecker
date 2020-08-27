package healthchecker

import (
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	log "github.com/sirupsen/logrus"
)

// TODO: refactor ICMPV4Check
// TODO: add constructor functions to link up with main checker

type ICMPChecker struct {
	Conn *icmp.PacketConn
}

func NewICMPChecker(timeout time.Duration) (*ICMPChecker, error) {
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


func (i *ICMPChecker) ICMPV4Check(targetIP *net.IPAddr) *CheckResult {
	timeStart := time.Now()
	echoReq := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID: 0,
			Seq: 1,
			Data: []byte("sirmackk/healthchecker"),
		},
	}
	encodedReq, _ := echoReq.Marshal(nil)
	_, err := i.Conn.WriteTo(encodedReq, targetIP)
	if err != nil {
		return &CheckResult{
			Timestamp: timeStart,
			Result: Failure,
			Duration: time.Since(timeStart),
		}
	}

	// TODO do I need to check the peer?
	buf := make([]byte, 2048)
	bytesRead, _, err := i.Conn.ReadFrom(buf)
	if err != nil {
		log.Debugf("ICMP check failed, couldn't read from conn: %v", err)
		return &CheckResult{
			Timestamp: timeStart,
			Result: Failure,
			Duration: time.Since(timeStart),
		}
	}

	resp, err := icmp.ParseMessage(58, buf[:bytesRead])
	if err != nil {
		log.Debugf("ICMP check failed, couldn't parse reply: %v - %v", err, buf[:bytesRead])
		return &CheckResult{
			Timestamp: timeStart,
			Result: Failure,
			Duration: time.Since(timeStart),
		}
	}

	switch resp.Type {
	case ipv4.ICMPTypeEchoReply:
		return &CheckResult{
			Timestamp: timeStart,
			Result: Success,
			Duration: time.Since(timeStart),
		}
	default:
		log.Debugf("ICMP check failed: %v - %v", targetIP, resp.Type)
		return &CheckResult{
			Timestamp: timeStart,
			Result: Failure,
			Duration: time.Since(timeStart),
		}
	}
}
