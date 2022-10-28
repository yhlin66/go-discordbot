package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/yhlin66/go-discordbot/controller"
)

var DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")

func main() {

		dg, err := discordgo.New("Bot " + DISCORD_TOKEN)
		if err != nil {
			fmt.Println("error creating discord session ", err)
			return
		}
	
		dg.AddHandler(controller.WeatherCreate)
	
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