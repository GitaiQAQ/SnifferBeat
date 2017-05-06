// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
	"github.com/gitaiqaq/serial"
)

type Config struct {
	Period			time.Duration 		`config:"period"`
	SerialConfig 	[]serial.Config 	`config:"serials"`
}

var DefaultConfig = Config{
	Period: 3 * time.Second,
	SerialConfig: []serial.Config{
		serial.Config{
			Name: "/dev/ttyUSB0",
			Baud: 115200,
		},
	},
}
