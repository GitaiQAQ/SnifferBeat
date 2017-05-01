package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/gitaiqaq/snifferbeat/beater"
)

func main() {
	err := beat.Run("snifferbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
