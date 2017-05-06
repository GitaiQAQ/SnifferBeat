package beater

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gitaiqaq/serial"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/gitaiqaq/snifferbeat/config"
)

type Snifferbeat struct {
	done   chan struct{}
	frames chan string
	config config.Config
	client publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	logp.Info("Config: %v", config)

	bt := &Snifferbeat{
		done:   make(chan struct{}),
		frames: make(chan string, 1000),
		config: config,
	}

	return bt, nil
}

func (bt *Snifferbeat) Run(b *beat.Beat) error {
	logp.Info("SerialPool is running!")
	for _, serialConfig := range bt.config.SerialConfig {
		go SerialPool(&serialConfig, bt.frames)
	}
	fmt.Println("Snifferbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		len_of_frames := len(bt.frames)
		fmt.Printf("Sync %v item(s) at %v\n", len_of_frames, common.Time(time.Now()))
		for i := 0; i < len_of_frames; i++ {
			frame :=<- bt.frames

			tokens := strings.Split(frame, "|")
			version, err := strconv.Atoi(tokens[0])
			if err != nil {
				continue
			}
			frameType, err := strconv.Atoi(tokens[1])
			if err != nil {
				continue
			}
			frameSubType, err := strconv.Atoi(tokens[2])
			if err != nil {
				continue
			}
			rssi, err := strconv.Atoi(tokens[5])
			if err != nil {
				continue
			}
			channel, err := strconv.Atoi(tokens[6])
			if err != nil {
				channel = 0
			}
			bt.client.PublishEvent(common.MapStr{
				"type":         b.Name,
				"version":      version,
				"frameType":    frameType,
				"frameSubType": frameSubType,
				"@timestamp":   common.Time(time.Now()),
				"chipId":       tokens[3],
				"rssi":         rssi,
				"channel":      channel,
				"senderMAC":    tokens[7],
				"reciverMAC":   tokens[8],
				"ssid":         tokens[9],
			})
			logp.Info("Sent %v", frame)
		}
		fmt.Printf("Sync %v item(s) at %v SUCCESS!\n", len_of_frames, common.Time(time.Now()))
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
	if len(b) < 54 {
		return false
	}
	return is_Number(b[0]) && b[1] == 124
}

func SerialPool(config *serial.Config, frames chan string) error {
	serialPort, err := serial.OpenPort(config)
	if err != nil {
		fmt.Errorf("Could not open port '%v'.", config.Name)
		logp.Err("Could not open port '%v'.", config.Name, err.Error())
		return err
	}
	scanner := bufio.NewScanner(serialPort.File())
	for scanner.Scan() {
		if is_Frame(scanner.Bytes()) {
			frames <- scanner.Text()
		} else {
			logp.Err("Unknown format: %v", scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		logp.Err("%v", err)
		return err
	}
	return nil
}
