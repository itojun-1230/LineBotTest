package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"io"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/liyue201/goqr"
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
	bot, _ := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)

	events, _ := bot.ParseRequest(c.Request)

	content, _ := bot.GetMessageContent(events[0].Message.(*linebot.ImageMessage).ID).Do()
	defer content.Content.Close()

	img, _ := convertReadCloserToImage(content.Content)
	result, _ := getQRCodeContent(img)

	var replyToken = events[0].ReplyToken

	if result == "" {
		message := "QRコードが見つかりませんでした...\nもう一度お試しください..."
		bot.ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do()
	}

	bot.ReplyMessage(replyToken, linebot.NewTextMessage(result)).Do()

}

func getQRCodeContent(img image.Image) (string, error) {
	// QRコードのデコード
	codes, err := goqr.Recognize(img)
	if err != nil {
		return "", err
	}

	// QRコードの内容を出力
	for _, code := range codes {
		return string(code.Payload), nil
	}
	return "", nil
}

func convertReadCloserToImage(reader io.ReadCloser) (image.Image, error) {
	defer reader.Close() // 関数終了時にクローズ

	// バイナリデータをすべて読み込み
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// バッファを使用してデコード
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}
