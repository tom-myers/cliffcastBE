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

var d1, d2, d3 Final

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
func getInfo(url string) []byte {

	res, err := http.Get(url)
	check(err)
	info, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	check(err)

	return info

}

// unmarshal sorts through json to filter what is needed
func unmarshal(info []byte) []Res {
	// to struct
	var data R1
	err := json.Unmarshal(info, &data)
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

// quickPrint just displays to console (and now returns Final)
func quickPrint(f Final) Final {
	var r Final
	rmin, rmax := MinMax(f.RainChance)
	tmin, tmax := MinMax(f.Temp)
	wmin, wmax := MinMax(f.Wind)
	gmin, gmax := MinMax(f.Gust)
	hmin, hmax := MinMax(f.Humid)

	r.Day = f.Day
	r.RainTotal = f.RainTotal // math.Round(f.RainTotal) - rounds to whole?!?!?
	r.RainChance = append(r.RainChance, rmin, rmax)
	r.Temp = append(r.Temp, tmin, tmax)
	r.Wind = append(r.Wind, wmin, wmax)
	r.Gust = append(r.Gust, gmin, gmax)
	r.Humid = append(r.Humid, hmin, hmax)

	fmt.Println(f.Day)
	fmt.Printf("chance of rain between %d percent and %d percent\n", rmin, rmax)
	fmt.Printf("total rain = %.2f cm\n", f.RainTotal)
	fmt.Printf("temp between %dc and %dc\n", tmin, tmax)
	fmt.Printf("wind speed between %dmph and %dmph\n", wmin, wmax)
	fmt.Printf("gusts between %d and %d\n", gmin, gmax)
	fmt.Printf("humidity between %d and %d\n", hmin, hmax)
	return r
}

// forecast checks through forecast for next 7 days
func forecast(fCast []Res) (Final, Final, Final) {

	var tDay string
	//var d1, d2, d3 Final
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

	d1 = quickPrint(d1)
	d2 = quickPrint(d2)
	d3 = quickPrint(d3)
	return d1, d2, d3

}

// Call returns data
func Call(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, d1, d2, d3)
}

// Refresh resfreshes data
func Refresh(w http.ResponseWriter, r *http.Request) {
	Start()
	fmt.Fprint(w, "refreshing!")
}

// GetURL posts URL
func GetURL(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "expects to be passed a URL") // future proofing to allow for more locations to be added
}

// Start is used to get info and start passing info about
func Start() {

	url := "http://ws1.metcheck.com/ENGINE/v9_0/json.asp?lat=53.9&lon=-1.6&lid=67633&Fc=No"
	info := getInfo(url)
	unmarsh := unmarshal(info)
	forecast(unmarsh)

}

func main() {
	Start()
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/call", Call).Methods("GET")
	r.HandleFunc("/refresh", Refresh).Methods("GET")
	r.HandleFunc("/Url", GetURL).Methods("POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST"}), handlers.AllowedOrigins([]string{"*"}))(r)))

}
