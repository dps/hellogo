package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var updateInterval = flag.Int("update_interval", 30, "update interval in mins.")
var city = flag.String("city", "san+francisco", "city")
var accessToken = flag.String("spark_access_token", "", "Access token for your spark core - required.")
var deviceId = flag.String("spark_device_id", "48ff6c065067555029251287", "Spark Core device ID.")

type weatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin   float64 `json:"temp"`
		Pressure int     `json:"pressure"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
}

func query(city string) (weatherData, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city)
	if err != nil {
		return weatherData{}, err
	}

	defer resp.Body.Close()

	var d weatherData

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return weatherData{}, err
	}

	return d, nil
}

func update(now time.Time) {
	data, _ := query(*city)
	fmt.Printf("%v %d mbar\n", now, data.Main.Pressure)
	mbar := strconv.Itoa(data.Main.Pressure)

	uri := "https://api.spark.io/v1/devices/" + *deviceId + "/servo"
	resp, err := http.PostForm(uri, url.Values{"access_token": {*accessToken}, "args": {mbar}})
	if err != nil {
		resp.Body.Close()
	}
}

func main() {
	flag.Parse()
	if *accessToken == "" {
		println("You must provide a valid spark_access_token.")
		return
	}
	update(time.Now())
	c := time.Tick(time.Duration(*updateInterval) * time.Minute)
	for now := range c {
		update(now)
	}
}
