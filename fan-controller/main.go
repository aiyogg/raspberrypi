package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

var (
	pin = rpio.Pin(14)
)

func readTemp() (int64, error) {
	data, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		log.Println("err: ", err)
		return 0, err
	}

	temp, _ := strconv.ParseInt(string(data[:len(data)-1]), 10, 64)
	return temp, nil
}

func check(quit chan bool) {
	temp, err := readTemp()

	if err != nil {
		log.Println(err)
		quit <- true
	}
	if temp >= 54000 {
		pin.High()
	} else if temp <= 46000 {
		pin.Low()
	}
	log.Printf("Current temperature: %d", temp)
}

func main() {
	if err := rpio.Open(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	pin.Output()
	defer rpio.Close()

	quit := make(chan bool)

	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case <-ticker.C:
			check(quit)
		case <-quit:
			os.Exit(1)
		}
	}
}
