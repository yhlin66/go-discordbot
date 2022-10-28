package controller

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/yhlin66/go-discordbot/api"
)

var tr = map[string]string{
	"Wx":   "天氣狀況",
	"MaxT": "最高溫度",
	"MinT": "最低溫度",
	"CI":   "舒適度",
	"PoP":  "降雨機率",
}

func WeatherCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	fmt.Println(m.Content)

	weather := api.WeatherApi(m.Content)

	embed := setWeatherEmbed(&weather)

	st, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(st)

}

func setWeatherEmbed(weather *api.Weather) *discordgo.MessageEmbed {
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