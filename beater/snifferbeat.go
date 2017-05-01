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
	frames   chan Frame
	config config.Config
	client publisher.Client
}

type Frame struct {
	version 		int
	frameType 		int
	frameSubType 	int
	time 		 	common.Time
	chipId 			string
	rssi			string
	channel			int
	reciverMAC				string
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
		frames:  make(chan Frame, 30),
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
		uniqeMap := map[string]Frame{}
		for frame := range bt.frames {
			uniqeMap[frame.senderMAC] = frame
    	}
    	for _,frame := range uniqeMap {
    		fmt.Println("frame", frame.senderMAC)
	    	event := common.MapStr{
				"type"			: b.Name,
				"version"		: frame.version,
				"frameType"		: frame.frameType,
				"frameSubType"	: frame.frameSubType,
				"@timestamp"	: frame.time,
				"chipId"		: frame.chipId,
				"rssi"			: frame.rssi,
				"channel"		: frame.channel,
				"reciverMAC"	: frame.reciverMAC,
				"senderMAC"		: frame.senderMAC,
				"ssid"			: frame.ssid,
			}
			bt.client.PublishEvent(event)
			logp.Info("Event sent")
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
	return is_Number(b[0]) && b[1] == 124
}

func SerialPool(config *serial.Config, frames chan Frame) error {
	serialPort, err := serial.OpenPort(config)
    if err != nil {
		fmt.Println("Err:", err)
        return err
    }
    scanner := bufio.NewScanner(serialPort.File())
    for scanner.Scan() {
    	if(is_Frame(scanner.Bytes())){
	    	tokens := strings.Split(scanner.Text(), "|")
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
	    	channel, err	:= strconv.Atoi(tokens[6])
	    	if err != nil {
	    		channel = 0
	    	}
	    	frames <- Frame{
	    		version:		version,
				frameType:		frameType,
				frameSubType:	frameSubType,
				time:			common.Time(time.Now()),
				chipId:			tokens[3],
				rssi:			tokens[5],
				channel:		channel,
				reciverMAC:		tokens[7],
				senderMAC:		tokens[8],
				ssid:			tokens[9],
	    	}
    	} else {
    		fmt.Println(scanner.Text())
    	}
    }

    if err := scanner.Err(); err != nil {
       	logp.Err("Err:", err)
        return err
    }
    return nil
}