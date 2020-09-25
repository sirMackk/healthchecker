package healthchecker

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	influx_client "github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
)

type Emitter interface {
	Emit(name, checkType string, c *Result) chan struct{}
	Name() string
}

type FileSink struct {
	TargetFile *os.File
}

func NewFileSink(args map[string]string) (Emitter, error) {
	errPrefix := "Error creating FileSink - "
	path, ok := args["path"]
	if !ok {
		return nil, fmt.Errorf("%s path parameter missing", errPrefix)
	}
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("%s path must be absolute, got: %s", errPrefix, path)
	}

	pathDir := filepath.Dir(path)
	if _, err := os.Stat(pathDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s target directory does not exist: %s", errPrefix, pathDir)
	}

	targetFile, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return nil, fmt.Errorf("%s cannot open file: %s", errPrefix, pathDir)
	}
	return &FileSink{TargetFile: targetFile}, nil
}

func (f *FileSink) Emit(name, checkType string, c *Result) chan struct{} {
	ret := make(chan struct{})
	defer close(ret)
	fmt.Fprintf(f.TargetFile, "%s [%s]: %s %s\n", c.TimestampString(), name, c.Result, c.Duration.Round(time.Millisecond))
	return ret
}

func (f *FileSink) Name() string {
	return "FileSink"
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
			log.Debug("InfluxSink reached flush interval - flushing batch points")
			err := s.Client.Write(bp)
			if err != nil {
				log.Errorf("InfluxSink encountered problem while writing to db: %s", err)
			}
			bp = s.newBatchPoints()
		case point := <-s.pointBox:
			log.Debug("Received data point")
			bp.AddPoint(point)
			if len(bp.Points()) >= flushCount {
				log.Debug("InfluxSink reached flush count - flushing batch points")
				err := s.Client.Write(bp)
				if err != nil {
					log.Errorf("InfluxSink encountered problem while writing to db: %s", err)
				}
				bp = s.newBatchPoints()
			}
		}
	}
}

func (s *UDPInfluxSink) Emit(name, checkType string, c *Result) chan struct{} {
	ret := make(chan struct{})
	defer close(ret)
	tags := map[string]string{
		"name": name,
		"type": checkType,
	}
	fields := map[string]interface{}{
		"result":   c.Result,
		"duration": int64(c.Duration / time.Millisecond),
	}
	pt, _ := influx_client.NewPoint("healthcheck", tags, fields, c.Timestamp)
	s.pointBox <- pt
	return ret
}

func (s *UDPInfluxSink) Name() string {
	return "UDPInfluxSink"
}
