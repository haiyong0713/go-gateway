package newyear2021

import (
	"time"
)

var (
	currentUnix int64
)

func init() {
	currentUnix = time.Now().Unix()
	go updateCurrentUnix()
}

func updateCurrentUnix() {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			currentUnix = time.Now().Unix()
		}
	}
}
