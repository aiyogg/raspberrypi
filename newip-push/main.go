package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func init() {
	getConfig()
}

func getConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}

func sendMail(target, subject, content string) {
	server := viper.GetString("mail.smtp")
	port := viper.GetInt("mail.smtp-port")
	user := viper.GetString("mail.user")
	pwd := viper.GetString("mail.password")
	d := gomail.NewDialer(server, port, user, pwd)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", user, "robot")
	m.SetHeader("To", target)
	m.SetAddressHeader("Cc", user, "admin")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	if err := d.DialAndSend(m); err != nil {
		log.Println("Email Error:", err)
		return
	}
}

func getIP(res chan string) {
	resp, err := http.Get(viper.GetString("api"))
	if err != nil {
		log.Fatalf("Request api failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	res <- string(body)
}

type dnsReqBody struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Content string `json:"content"`
}

func setDNSRecords(ip string) (res string) {
	url := viper.GetString("cloudflare.url")
	record := viper.GetString("cloudflare.record")
	zoneId := viper.GetString("cloudflare.zone_id")
	id := viper.GetString("cloudflare.id")
	email := viper.GetString("cloudflare.email")
	apiKey := viper.GetString("cloudflare.api_key")
	body, _ := json.Marshal(dnsReqBody{"A", record, strings.TrimSpace(ip)})

	req, err := http.NewRequest("PUT", fmt.Sprintf(url, zoneId, id), bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %#v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	resBody, _ := ioutil.ReadAll(resp.Body)
	return string(resBody)
}

func main() {
	c := cron.New()

	ipres := make(chan string, 1)

	c.AddFunc("0 */30 * * * *", func() {
		go getIP(ipres)
		ip := <-ipres
		log.Printf("IP address: %s", ip)
		oldip, err := ioutil.ReadFile("ip.txt")
		if err != nil {
			log.Fatalf("Read file failed: %v \r\n", err)
			return
		}
		if strings.Compare(string(oldip), ip) != 0 {
			sendMail("i@chenteng.me", "新的公网ip为："+ip, "旧的ip为: "+string(oldip)+"\r\n新的ip为："+ip)
			ioutil.WriteFile("ip.txt", []byte(ip), 0777)
			log.Println("Old IP: " + string(oldip) + "New IP：" + ip)

			recordRes := setDNSRecords(ip)
			log.Printf("setDNSRecords result: %s \r\n", recordRes)
		} else {
			log.Println("IP is not modified：", string(oldip))
		}
	})

	c.Run()
}
