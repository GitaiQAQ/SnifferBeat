package beater

import (
	"fmt"
	"time"
	"bufio"
	"strings"
	"strconv"
	"github.com/gitaiqaq/serial"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/gitaiqaq/snifferbeat/config"
)

type Snifferbeat struct {
	done   chan struct{}
	frames   chan string
	config config.Config
	client publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Snifferbeat{
		done:   make(chan struct{}),
		frames:  make(chan string),
		config: config,
	}

	return bt, nil
}

func (bt *Snifferbeat) Run(b *beat.Beat) error {
	logp.Info("SerialPool is running!")
	go SerialPool(&bt.config.SerialConfig, bt.frames)
	logp.Info("snifferbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
    	for frame := range bt.frames {
    		tokens := strings.Split(frame, "|")
	    	version, err	:= strconv.Atoi(tokens[0])
	    	if err != nil {
	    		continue
	    	}
	    	frameType, err	:= strconv.Atoi(tokens[1])
	    	if err != nil {
	    		continue
	    	}
	    	frameSubType, err	:= strconv.Atoi(tokens[2])
	    	if err != nil {
	    		continue
	    	}
	    	rssi, err	:= strconv.Atoi(tokens[5])
	    	if err != nil {
	    		continue
	    	}
	    	channel, err	:= strconv.Atoi(tokens[6])
	    	if err != nil {
	    		channel = 0
	    	}
	    	event := common.MapStr{
				"type"			: b.Name,
				"version"		: version,
				"frameType"		: frameType,
				"frameSubType"	: frameSubType,
				"@timestamp"	: common.Time(time.Now()),
				"chipId"		: tokens[3],
				"rssi"			: rssi,
				"channel"		: channel,
				"senderMAC"		: tokens[7],
				"reciverMAC"	: tokens[8],
				"ssid"			: tokens[9],
			}
			bt.client.PublishEvent(event)
			logp.Info("Sent data %v", frame)
    	}
	}
}

func (bt *Snifferbeat) Stop() {
	bt.client.Close()
	close(bt.done)
	close(bt.frames)
}

func is_Number(b byte) bool {
	return b >= 48 && b <= 57
}

func is_Frame(b []byte) bool {
	if (len(b) < 54) {
		return false
	}
	return is_Number(b[0]) && b[1] == 124
}

func SerialPool(config *serial.Config, frames chan string) error {
	serialPort, err := serial.OpenPort(config)
    if err != nil {
		logp.Err("Could not open port '%v'.", config.Name, err)
        return err
    }
    scanner := bufio.NewScanner(serialPort.File())
    for scanner.Scan() {
    	if(is_Frame(scanner.Bytes())){
	    	frames <- scanner.Text()
    	} else {
    		logp.Err("Unknown format: ", scanner.Text())
    	}
    }

    if err := scanner.Err(); err != nil {
		logp.Err("%v", err)
        return err
    }
    return nil
}