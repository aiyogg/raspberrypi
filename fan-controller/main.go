package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

var (
	pin = rpio.Pin(14)
)

type Temperature struct {
	ID          uint       `json:"id" gorm:"primary_key"`
	Temperature float32    `json:"temperature" gorm:"type:numeric"`
	CreatedAt   *time.Time `json:"time" `
}

func readTemp() (int64, error) {
	data, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		log.Println("err: ", err)
		return 0, err
	}

	temp, _ := strconv.ParseInt(string(data[:len(data)-1]), 10, 64)
	log.Printf("Current temperature: %d", temp)
	return temp, nil
}

func check(temp int64) {
	if temp >= 58000 {
		pin.High()
	} else if temp <= 48000 {
		pin.Low()
	}
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
func getTempHandle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	startTime := query.Get("s")
	endTime := query.Get("e")
	w.Write([]byte(startTime + endTime))
}

func main() {
	// pin open
	if err := rpio.Open(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	pin.Output()
	defer rpio.Close()

	// db connect
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=chuck password=chuck@2020 sslmode=disable dbname=pgdb")
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	db.DropTableIfExists(Temperature{})
	db.CreateTable(Temperature{})

	// timer
	quit := make(chan bool)
	ticker := time.NewTicker(time.Minute * 2)

	go func() {
		log.Println("goroutine...")
		for {
			select {
			case <-ticker.C:
				if temp, err := readTemp(); err != nil {
					quit <- true
				} else {
					check(temp)
					db.Create(&Temperature{Temperature: float32(temp) / 1000})
				}
			case <-quit:
				os.Exit(1)
			}
		}
	}()

	// http server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	http.HandleFunc("/get", getTempHandle)
	http.HandleFunc("/", indexHandle)
	http.ListenAndServe(":10001", nil)
	quit <- true
}
