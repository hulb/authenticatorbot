package main

import (
	"crypto/tls"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	tb "gopkg.in/tucnak/telebot.v2"
)

var token string

func init() {
	token = os.Getenv("BOTTOKEN")
	if token == "" {
		token = ""
		log.Fatalln("bot token required.exit")
	}
}

func main() {
	// DEBUG set http proxy
	proxy, _ := url.Parse("http://192.168.1.1:1080")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		Client: client,
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		if m.Photo != nil {

			f, err := b.GetFile(&m.Photo.File)
			defer f.Close()

			if err != nil {
				log.Println(err)
				b.Send(m.Sender, err)
				return
			}

			img, _, err := image.Decode(f)
			if err != nil {
				log.Println(err)
				b.Send(m.Sender, "invalid pic")
				return
			}

			// prepare BinaryBitmap
			bmp, err := gozxing.NewBinaryBitmapFromImage(img)
			if err != nil {
				log.Println(err)
				b.Send(m.Sender, "invalid qrcode pic")
				return
			}

			// decode image
			qrReader := qrcode.NewQRCodeReader()
			result, err := qrReader.Decode(bmp, nil)
			if err != nil {
				log.Println(err)
				b.Send(m.Sender, "qrcode pic decode fails")
				return
			}

			authURL, err := url.Parse(result.GetText())
			if err != nil {
				log.Println(err)
				b.Send(m.Sender, "not a valid otpauth qrcode pic")
				return
			}

			splitArr := strings.Split(authURL.Path[1:], ":")
			if len(splitArr) != 2 {
				b.Send(m.Sender, "not a valid otpauth qrcode pic")
				return
			}

			account, err := ParseAccount(authURL.String())
			if err != nil {
				log.Println(err)
				b.Send(m.Sender, fmt.Sprintf("url %s parse fail", authURL.String()))
				return
			}

			if err := account.Save(); err != nil {
				log.Println(err)
				b.Send(m.Sender, fmt.Sprintf("save account fail, err: %v", err))
			}
			b.Send(m.Sender, string(account.JSON()))
			// lable := splitArr[0]
			// username := splitArr[1]
			// // b.Send(m.Sender, result.GetText())

			// db, err := bolt.Open("my.db", 0600, nil)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// defer db.Close()

		}
	})

	b.Start()
}
