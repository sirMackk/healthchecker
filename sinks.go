package healthchecker

import (
	"fmt"
	"os"
	"strconv"
	"time"

	influx_client "github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
)

type Emitter interface {
	Emit(name, checkType string, c *CheckResult)
}

type ConsoleSink struct {
	TargetStream *os.File
}

func NewConsoleSink(args map[string]string) (Emitter, error) {
	useStdout, ok := args["useStdout"]
	if !ok {
		return nil, fmt.Errorf("Error creating ConsoleSink - useStdout option missing")
	}
	choice, err := strconv.ParseBool(useStdout)
	if err != nil {
		return nil, fmt.Errorf("Error parsing ConsoleSink 'useStdout' option")
	}

	if !choice {
		return &ConsoleSink{TargetStream: os.Stderr}, nil
	}
	return &ConsoleSink{TargetStream: os.Stdout}, nil
}

func (s *ConsoleSink) Emit(name, checkType string, c *CheckResult) {
	fmt.Fprintf(s.TargetStream, "%s [%s]: %s %s\n", c.TimestampString(), name, c.Result, c.Duration.Round(time.Millisecond))
}

type UDPInfluxSink struct {
	Client           influx_client.Client
	client_config    *influx_client.UDPConfig
	pointBox         chan *influx_client.Point
	collectorRunning bool
}

func NewUDPInfluxSink(args map[string]string) (Emitter, error) {
	addr, ok := args["addr"]
	if !ok {
		return nil, fmt.Errorf("Error creating UDPInfluxSink - addr option missing")
	}
	flushIntervalConf, ok := args["flushInterval"]
	if !ok {
		return nil, fmt.Errorf("Error creating UDPInfluxSink - flushInterval missing")
	}
	flushIntervalVal, err := strconv.Atoi(flushIntervalConf)
	if err != nil {
		return nil, fmt.Errorf("Error creating UDPInfluxSink - flushInterval not valid integer")
	}
	flushInterval := time.Duration(flushIntervalVal)
	flushCountConf, ok := args["flushCount"]
	if !ok {
		return nil, fmt.Errorf("Error creating UDPInfluxSink - flushCount missing")
	}
	flushCount, err := strconv.Atoi(flushCountConf)
	if err != nil {
		return nil, fmt.Errorf("Error creating UDPInfluxSink - flushCount not valid integer")
	}

	conf := influx_client.UDPConfig{Addr: addr}
	c, err := influx_client.NewUDPClient(conf)
	if err != nil {
		return nil, fmt.Errorf("Error while creating Influx UDP client: %s", err)
	}

	pointBox := make(chan *influx_client.Point, 1)
	sink := &UDPInfluxSink{
		Client:           c,
		client_config:    &conf,
		pointBox:         pointBox,
		collectorRunning: false,
	}
	sink.StartCollector(flushInterval, flushCount)
	return sink, nil
}

func (s *UDPInfluxSink) StartCollector(flushInterval time.Duration, flushCount int) {
	log.Debug("Starting InfluxDB collector")
	if !s.collectorRunning {
		s.collectorRunning = true
		go s.collectorRoutine(flushInterval, flushCount)
	} else {
		log.Error("Cannot start more than 1 collector goro for one UDPInfluxSink!")
	}
}

func (s *UDPInfluxSink) IsCollectorRunning() bool {
	return s.collectorRunning
}

func (s *UDPInfluxSink) newBatchPoints() influx_client.BatchPoints {
	bp, _ := influx_client.NewBatchPoints(influx_client.BatchPointsConfig{
		Precision: "ms",
	})
	return bp
}

func (s *UDPInfluxSink) collectorRoutine(flushInterval time.Duration, flushCount int) {
	defer s.Client.Close()
	bp := s.newBatchPoints()
	for {
		select {
		case <-time.Tick(flushInterval * time.Second):
			log.Debug("Reached flush interval, flushing batch points")
			err := s.Client.Write(bp)
			if err != nil {
				log.Debugf("Error while writing to db: %s", err)
			}
			bp = s.newBatchPoints()
		case point := <-s.pointBox:
			log.Debug("Received data point")
			bp.AddPoint(point)
			if len(bp.Points()) >= flushCount {
				log.Debug("Reached flush count, flushing batch points")
				err := s.Client.Write(bp)
				if err != nil {
					log.Debugf("Error while writing to db: %s", err)
				}
				bp = s.newBatchPoints()
			}
		}
	}
}

func (s *UDPInfluxSink) Emit(name, checkType string, c *CheckResult) {
	tags := map[string]string{
		"name": name,
		"type": checkType,
	}
	fields := map[string]interface{}{
		"result":   c.Result,
		"duration": c.Duration,
	}
	pt, _ := influx_client.NewPoint("healthcheck", tags, fields, c.Timestamp)
	s.pointBox <- pt
}
