package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// InputURL is URL passed to endpoint
type InputURL struct {
	I string `json:"url"`
}

// R1 is top level struct
type R1 struct {
	D1 D1 `json:"metcheckData"`
}

// D1 is level 2 struct
type D1 struct {
	Location F1 `json:"forecastLocation"`
}

// F1 is level 3
type F1 struct {
	Forecast []Res `json:"forecast"`
}

// Final is used to calc the final values to display
type Final struct {
	Day        string
	RainTotal  float64
	RainChance []int64
	Temp       []int64
	Wind       []int64
	Gust       []int64
	Humid      []int64
}

// Res represents the things we actually want from the json response
type Res struct {
	Temp   string `json:"temperature"`
	Chance string `json:"chanceofrain"`
	Rain   string `json:"rain"`
	Wind   string `json:"windgustspeed"`
	Humid  string `json:"humidity"`
	Utc    string `json:"utcTime"`
	DayN   string `json:"weekday"`
	WindS  string `json:"windspeed"`
}

// check is a basic error checker
func check(e error) {
	if e != nil {
		fmt.Println("an error occured... panic!!!")
		log.Fatal(e)
	}
}

// getInfo calls endpoint and returns []byte
func getInfo(url string) []Res {

	res, err := http.Get(url)
	check(err)
	info, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	check(err)

	// to struct
	var data R1
	err = json.Unmarshal(info, &data)
	check(err)
	fCast := data.D1.Location.Forecast

	return fCast

}

// MinMax returns the min and max values frm int64 slice
func MinMax(array []int64) (int64, int64) {
	var max = array[0]
	var min = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

// Format just displays to console (and now returns Final)
func Format(f Final) Final {
	var r Final
	rmin, rmax := MinMax(f.RainChance)
	tmin, tmax := MinMax(f.Temp)
	wmin, wmax := MinMax(f.Wind)
	gmin, gmax := MinMax(f.Gust)
	hmin, hmax := MinMax(f.Humid)

	r.Day = f.Day
	r.RainTotal = f.RainTotal
	r.RainChance = append(r.RainChance, rmin, rmax)
	r.Temp = append(r.Temp, tmin, tmax)
	r.Wind = append(r.Wind, wmin, wmax)
	r.Gust = append(r.Gust, gmin, gmax)
	r.Humid = append(r.Humid, hmin, hmax)

	return r
}

// forecast checks through forecast for next 7 days
func forecast(fCast []Res) (Final, Final, Final) {

	var tDay string
	var d1, d2, d3 Final
	loc, _ := time.LoadLocation("UTC")
	layout := "2006-01-02T15:04:05"
	now := time.Now().In(loc).Truncate(24 * time.Hour)

	for _, t := range fCast {
		// attempt to parse utc to time
		tt, err := time.Parse(layout, t.Utc)
		check(err)
		diff := now.Sub(tt.Truncate(24 * time.Hour))

		xTemp, err := strconv.ParseInt(t.Temp, 10, 10)
		check(err)
		xRChance, err := strconv.ParseInt(t.Chance, 10, 10)
		check(err)
		xRTotal, err := strconv.ParseFloat(t.Rain, 10)
		check(err)
		xWind, err := strconv.ParseInt(t.Wind, 10, 10)
		check(err)
		xGust, err := strconv.ParseInt(t.Wind, 10, 10)
		check(err)
		xHumid, err := strconv.ParseInt(t.Humid, 10, 10)
		check(err)

		switch {
		case diff.Hours() == 0: // today

			tDay = "today"
			d1.Day = tDay
			d1.RainChance = append(d1.RainChance, xRChance)
			d1.RainTotal = d1.RainTotal + xRTotal
			d1.Wind = append(d1.Wind, xWind)
			d1.Gust = append(d1.Gust, xGust)
			d1.Temp = append(d1.Temp, xTemp)
			d1.Humid = append(d1.Humid, xHumid)

		case diff.Hours() == -24: // tomorrow

			tDay = "tomorrow"
			d2.Day = tDay
			d2.RainChance = append(d2.RainChance, xRChance)
			d2.RainTotal = d2.RainTotal + xRTotal
			d2.Wind = append(d2.Wind, xWind)
			d2.Gust = append(d2.Gust, xGust)
			d2.Temp = append(d2.Temp, xTemp)
			d2.Humid = append(d2.Humid, xHumid)

		case diff.Hours() == -48: //day after

			tDay = t.DayN
			d3.Day = tDay
			d3.RainChance = append(d3.RainChance, xRChance)
			d3.RainTotal = d3.RainTotal + xRTotal
			d3.Wind = append(d3.Wind, xWind)
			d3.Gust = append(d3.Gust, xGust)
			d3.Temp = append(d3.Temp, xTemp)
			d3.Humid = append(d3.Humid, xHumid)

		default:
			// do nothing
		}

	}

	r1 := Format(d1)
	r2 := Format(d2)
	r3 := Format(d3)
	return r1, r2, r3

}

// Call returns data
func Call(w http.ResponseWriter, r *http.Request) {
	var input InputURL
	_ = json.NewDecoder(r.Body).Decode(&input)
	//json.NewEncoder(w).Encode(input)
	info := getInfo(input.I)
	d1, d2, d3 := forecast(info)
	fmt.Fprint(w, d1, d2, d3)
}

func main() {
	// start endpoint
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/call", Call).Methods("GET", "POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST"}), handlers.AllowedOrigins([]string{"*"}))(r)))

}
