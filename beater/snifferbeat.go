package beater

import (
	"fmt"
	"time"
	"bufio"

	"github.com/gitaiqaq/serial"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/gitaiqaq/snifferbeat/config"
)

type Snifferbeat struct {
	done   chan struct{}
	lines   chan string
	config config.Config
	client publisher.Client
}

type WIFIFrame struct {
	varsion 		int
	frameType 		int
	frameSubType 	int
	chipId 			string
	rssi			string
	channel			int
	receiverMAC		string
	senderMAC		string
	ssid			string
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Snifferbeat{
		done:   make(chan struct{}),
		lines:  make(chan string, 1000),
		config: config,
	}

	return bt, nil
}

func (bt *Snifferbeat) Run(b *beat.Beat) error {
	logp.Info("SerialPool is running!")
	go SerialPool(&bt.config.SerialConfig, bt.lines)
	logp.Info("snifferbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		
		for line := range bt.lines {
			event := common.MapStr{
				"@timestamp": common.Time(time.Now()),
				"type":       b.Name,
				"msg": 		  line,
			}
			bt.client.PublishEvent(event)
			logp.Info("Event sent")
    	}
	}
}

func (bt *Snifferbeat) Stop() {
	bt.client.Close()
	close(bt.done)
	close(bt.lines)
}

func is_Number(b byte) bool {
	return b >= 48 && b <= 57
}

func is_Line(b []byte) bool {
	return is_Number(b[0]) && b[1] == 124
}

func SerialPool(config *serial.Config, lines chan string) error {
	serialPort, err := serial.OpenPort(config)
    if err != nil {
		fmt.Println("Err:", err)
        return err
    }
    scanner := bufio.NewScanner(serialPort.File())
    for scanner.Scan() {
    	if (is_Line(scanner.Bytes())) {
    		logp.Info(scanner.Text())
    		lines <- scanner.Text()
    	}else{
    		fmt.Println(scanner.Text())
    	}
    }

    if err := scanner.Err(); err != nil {
       	logp.Err("Err:", err)
        return err
    }
    return nil
}