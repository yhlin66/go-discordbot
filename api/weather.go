package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

type Weather struct {
	Success string  `json:"success,omitempty"`
	Records Records `json:"records,omitempty"`
}

type Parameter struct {
	ParameterName  string `json:"parameterName,omitempty"`
	ParameterValue string `json:"parameterValue,omitempty"`
	ParameterUnit  string `json:"parameterUnit,omitempty"`
}
type Time struct {
	StartTime string    `json:"startTime,omitempty"`
	EndTime   string    `json:"endTime,omitempty"`
	Parameter Parameter `json:"parameter,omitempty"`
}
type WeatherElement struct {
	ElementName string `json:"elementName,omitempty"`
	Time        []Time `json:"time,omitempty"`
}
type Location struct {
	LocationName   string           `json:"locationName,omitempty"`
	WeatherElement []WeatherElement `json:"weatherElement,omitempty"`
}
type Records struct {
	DatasetDescription string     `json:"datasetDescription,omitempty"`
	Location           []Location `json:"location,omitempty"`
}

func (p Parameter) SetContext() string {
	s := p.ParameterName
	if p.ParameterValue != "" {
		s += p.ParameterValue
	}
	switch p.ParameterUnit {
	case "C":
		s += "°C\n "
	case "百分比":
		s += "%\n "
	default:
		s += "\n "
	}
	return s
}

var (
	CWB_TOKEN = os.Getenv("CWB_TOKEN")
	weather Weather
) 


func WeatherApi(m *discordgo.MessageCreate) Weather {

	// If the message is "cityweather" reply with city weather information by api
	if strings.Contains(m.Content, "weather") {
		var locationName string
		var city string = "高雄市"
		//去掉 weather 
		r := []rune(strings.TrimSuffix(m.Content, "weather"))
		//有輸入城市名稱，預設高雄市
		if len(r) > 0 {
			// 台 > 臺
			tai := r[:1]
			if tai[0] == 21488 {
				tai[0] = 33274
			}
			city = string(r)
		}
		locationName = url.QueryEscape(city)

		// http request cwb api
		res, err := http.Get("https://opendata.cwb.gov.tw/api/v1/rest/datastore/F-C0032-001?Authorization="+ CWB_TOKEN +"&format=JSON&locationName=" + locationName)
		if err != nil {
			fmt.Println(err)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
		}

		// unmarchal json
		json.Unmarshal([]byte(string(body)), &weather)
		fmt.Println(weather)
	}
	return weather
}