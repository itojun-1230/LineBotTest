package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file", err)
	}

	engine := gin.Default()
	engine.POST("/callback", postCallback)

	engine.Run(":" + os.Getenv("PORT"))
}

func postCallback(c *gin.Context) {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	events, berr := bot.ParseRequest(c.Request)
	if berr != nil {
		fmt.Println(berr)
		return
	}

	var replyToken string
	var message string

	replyToken = events[0].ReplyToken
	message = events[0].Message.(*linebot.TextMessage).Text

	_, err = bot.ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do()
	if err != nil {
		fmt.Println(err)
		return
	}
}
