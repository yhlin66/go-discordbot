package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

var DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")
var CWB_TOKEN = os.Getenv("CWB_TOKEN")

var tr = map[string]string{
	"Wx":   "天氣狀況",
	"MaxT": "最高溫度",
	"MinT": "最低溫度",
	"CI":   "舒適度",
	"PoP":  "降雨機率",
}

func DiscordBot() {

	dg, err := discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		fmt.Println("error creating discord session ", err)
		return
	}

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection ", err)
		return
	}
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	fmt.Println(m.Content)

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

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

		//new weather struct
		var weather Weather
		// unmarchal json
		json.Unmarshal([]byte(string(body)), &weather)
		fmt.Println(weather)
		embed := setWeatherEmbed(&weather)

		st, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(st)
	}

}

func setWeatherEmbed(weather *Weather) *discordgo.MessageEmbed {
	//Create a double dimensionate map to storage Embed's fieldContent
	var fieldContent = make(map[int]map[string]string)

	//New Embed
	embed := &discordgo.MessageEmbed{}
	//Set Embed's image
	embed.Image = &discordgo.MessageEmbedImage{URL: "https://image.shutterstock.com/image-vector/sunny-weather-icon-600w-1023412543.jpg"}
	//Set Embed's sidebar color
	embed.Color = 0x00ff00

	// for loop nested
	for _, p := range weather.Records.Location {
		//Set Embed's Title
		embed.Title = p.LocationName + "天氣資訊 by api"

		//Create a map to save the message per/day
		var pstring = make(map[int]string)
		for _, w := range p.WeatherElement {
			for i := range w.Time {
				fieldContent[i] = map[string]string{}
				//Combine StartTime and EndTime as the fieldContent.Name
				fieldContent[i]["time"] = w.Time[i].StartTime + " ~ " + w.Time[i].EndTime
				//Combine elementName and time.Parameter as the fieldContent.value
				pstring[i] += tr[w.ElementName] + ": " + w.Time[i].Parameter.SetContext()
				fieldContent[i]["value"] = pstring[i]
			}
		}
		//append Name and Value to embed.fields then set embed.field
		for i := 0; i < len(fieldContent); i = i + 1 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: fieldContent[i]["time"], Value: fieldContent[i]["value"]})
		}

	}
	return embed
}