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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type QRCode struct {
	ID	 uint			`gorm:"primaryKey"`
	Content string		`gorm:"not null"`
}
type UserQRCode struct {
	ID        uint    `gorm:"primaryKey"`
	QRCodeID 	uint		`gorm:"not null"`
	UserID 		string	`gorm:"not null"`
}

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

	db := db_open()
	db.AutoMigrate(&QRCode{})
	// 登録されているQRコードの内容を取得
	var qrCodes []QRCode
	db.Find(&qrCodes)
	// 登録されているQRコードの内容と一致するか確認
	var qrCode QRCode
	db.Where("content = ?", result).First(&qrCode)
	if qrCode.ID == 0 {
		message := "登録されていないQRコードです...\nもう一度お試しください..."
		bot.ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do()
		return
	}
	db.AutoMigrate(&UserQRCode{})
	// ユーザーIDを取得
	userID := events[0].Source.UserID
	// すでに登録されているか確認
	var userQRCode UserQRCode
	db.Where("user_id = ? AND qr_code_id = ?", userID, qrCode.ID).First(&userQRCode)
	if userQRCode.ID != 0 {
		message := "すでに登録されているQRコードです..."
		bot.ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do()
		return
	}

	// ユーザーIDとQRコードIDを登録
	db.Create(&UserQRCode{QRCodeID: qrCode.ID, UserID: userID})

	// ユーザーが登録したQRコードの数がQRCodesテーブルのレコード数と一致するか確認
	var userQRCodes []UserQRCode
	db.Where("user_id = ?", userID).Find(&userQRCodes)
	if len(userQRCodes) == len(qrCodes) {
		message := "全てのQRコードを登録しました!"
		bot.ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do()
	}else {
		message := result + "のスタンプの登録が完了しました!"
		bot.ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do()
	}
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

func db_open() *gorm.DB {
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := "localhost"
	dbPort := "5432"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo", dbHost, dbUser, dbPass, dbName, dbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
	}

	return db
}
