package main

import (
	"github.com/tortlewortle/bot/youtube"
	"log"
	"time"
)

func main() {
	yt := youtube.Client{
		ChannelID:    "-- Channel ID of bot account --",
		ChatID:       "-- LiveChat ID --",
		ApiKey:       "-- Google Client Secret --",
		RefreshToken: "-- Refresh token --",
		ClientID:     "-- Google Client Id --",
		ClientSecret: "-- Google Client Secret --",
	}

	yt.AddMessageHandler(func(msg youtube.Message) {
		if msg.Snippet.TextMessageDetails.MessageText == "!hello" {
			msg, err := msg.Reply("Hi :)")
			if err != nil {
				log.Println("Could not send message")
			}
			msg.Delete(500 * time.Millisecond)
		} else if msg.Snippet.TextMessageDetails.MessageText == "!delete" {
			msg.Delete(500 * time.Millisecond)
		}
	})

	yt.Start()
}
