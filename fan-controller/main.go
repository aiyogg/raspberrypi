package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/stianeikeland/go-rpio/v4"
)

var (
	pin      = rpio.Pin(14)
	globalDB *gorm.DB
)

type Temperature struct {
	ID          uint       `json:"-" gorm:"primary_key"`
	Temperature float32    `json:"temperature" gorm:"type:numeric"`
	CreatedAt   *time.Time `json:"time"`
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
	} else if temp <= 50500 {
		pin.Low()
	}
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
func getTempHandle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	start, _ := strconv.ParseInt(query.Get("s"), 10, 64)
	end, _ := strconv.ParseInt(query.Get("e"), 10, 64)

	now := time.Now()
	var startTime, endTime time.Time
	if start == 0 || start > now.Unix() {
		startTime = now.AddDate(0, 0, -1)
	} else {
		startTime = time.Unix(start, 0)
	}
	if end == 0 || end < start || end > now.Unix() {
		endTime = now
	} else {
		endTime = time.Unix(start, 0)
	}
	type Rsp struct {
		Temperatures []Temperature `json:"temperatures"`
		Code         int           `json:"code"`
		Msg          string        `json:"msg"`
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

	var (
		temperatures []Temperature
		rsp          Rsp
	)
	if err := globalDB.Where("created_at BETWEEN ? AND ?", startTime, endTime).Find(&temperatures).Error; err != nil {
		log.Println("DB query error", err)
		rsp = Rsp{Temperatures: temperatures, Code: -1, Msg: "db error"}
	} else {
		rsp = Rsp{Temperatures: temperatures, Code: 0, Msg: ""}
	}
	if err := json.NewEncoder(w).Encode(rsp); err != nil {
		log.Println("response error", err)
	}
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
	globalDB = db
	defer db.Close()
	db.DropTableIfExists(Temperature{})
	db.CreateTable(Temperature{})

	// timer
	quit := make(chan bool)
	ticker := time.NewTicker(time.Minute * 5)

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
