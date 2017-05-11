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
	// fmt.Println(config.SerialConfig)

	bt := &Snifferbeat{
		done:   make(chan struct{}),
		frames: make(chan string, 10000),
		config: config,
	}

	return bt, nil
}

func ParseFrame(logType string, frame string) (error, string, string, common.MapStr) {
	tokens := strings.Split(frame, "|")
	version, err := strconv.Atoi(tokens[0])
	if err != nil {
		return err, "", "", nil
	}
	frameType, err := strconv.Atoi(tokens[1])
	if err != nil {
		return err, "", "", nil
	}
	frameSubType, err := strconv.Atoi(tokens[2])
	if err != nil {
		return err, "", "", nil
	}
	rssi, err := strconv.Atoi(tokens[5])
	if err != nil {
		return err, "", "", nil
	}
	channel, err := strconv.Atoi(tokens[6])
	if err != nil {
		channel = 0
	}
	event := common.MapStr{
		"type":         logType,
		"version":      version,
		"frameType":    frameType,
		"frameSubType": frameSubType,
		"@timestamp":   common.Time(time.Now()),
		"within":       rssi > -40,
		"chipId":       tokens[3],
		"rssi":         rssi,
		"channel":      channel,
		"senderMAC":    tokens[7],
		"reciverMAC":   tokens[8],
		"frame":        frame,
	}
	if len(tokens) > 8 {
		event.Put("ssid", tokens[9])
	}
	return nil, tokens[7], tokens[3], event
}

func (bt *Snifferbeat) Run(b *beat.Beat) error {
	logp.Info("SerialPool is running!")
	fmt.Printf("%d Serials connected!\n", len(bt.config.SerialConfig))
	for _, serialConfig := range bt.config.SerialConfig {
		go SerialPool(serialConfig, bt.frames)
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
		if len(bt.config.SerialConfig) < 2 {
			for i := 0; i < len_of_frames; i++ {
				frame := <-bt.frames
				err, _, _, event := ParseFrame(b.Name, frame)
				if err != nil {
					continue
				}
				bt.client.PublishEvent(event)
				logp.Info("Sent %v", frame)
			}
		} else {
			eventMap := map[string]map[string]common.MapStr{}
			for i := 0; i < len_of_frames; i++ {
				frame := <-bt.frames
				err, sendMAC, chipID, event := ParseFrame(b.Name, frame)
				if err != nil {
					continue
				}

				eventsBySender, ok := eventMap[sendMAC]
				if !ok {
					eventsBySender = make(map[string]common.MapStr)
					eventMap[sendMAC] = eventsBySender
				}
				eventsByChip, ok := eventsBySender[chipID]
				if !ok {
					eventsBySender[chipID] = event
				} else {
					rssi, err := event.GetValue("rssi")
					if err != nil {
						continue
					}
					if sumofrssi, err := eventsByChip.GetValue("rssi"); err != nil {
						continue
					} else {
						eventsByChip.Put("rssi", sumofrssi.(int)+rssi.(int))
					}
					if count, err := eventsByChip.GetValue("count"); err != nil {
						eventsByChip.Put("count", 2)
					} else {
						eventsByChip.Put("count", count.(int)+1)
					}
				}
			}

			for _, eventsBySender := range eventMap {
				chipId := ""
				event := common.MapStr{}
				for cid, eventsByChip := range eventsBySender {
					count, err := eventsByChip.GetValue("count")
					if err != nil {
						eventsByChip.Put("count", 1)
						count = 1
					} else {
						eventsByChip.Put("count", count.(int)+1)
					}
					if sumofrssi, err := eventsByChip.GetValue("rssi"); err != nil {
						continue
					} else {
						// fmt.Printf("Elm: %v, SUM: %d, Count: %v, S/C: %v\n", eventsByChip, sumofrssi, count, sumofrssi.(int)/count.(int))
						eventsByChip.Put("rssi", sumofrssi.(int)/count.(int))
						chipId = chipId + "." + cid
						event = eventsByChip
					}
				}
				if len(eventsBySender) > len(bt.config.SerialConfig)-1 {
					event.Put("chip_id", chipId)
					event.Put("within", true)
				} else {
					event.Put("chip_id", chipId)
					event.Put("within", false)
				}
				event.Put("number", len(eventsBySender))
				bt.client.PublishEvent(event)
			}
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

func SerialPool(config serial.Config, frames chan string) error {
	serialPort, err := serial.OpenPort(&config)
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
