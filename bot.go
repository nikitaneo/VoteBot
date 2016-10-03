package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

type config struct {
	EmotionApiKey string `json:"emotion_api_key"`
	VkApiKey      string `json:"vk_api_key"`
	LoginDB       string `json:"login_db"`
	PasswordDB    string `json:"password_db"`
}

func main() {
	conf_data, err := ioutil.ReadFile("config.txt")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	var conf config
	err = json.Unmarshal(conf_data, &conf)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// open connection Microsoft Azure MySQL database
	db, err := sql.Open("mysql", conf.LoginDB+":"+conf.PasswordDB+"@tcp(eu-cdbr-azure-west-d.cloudapp.net:3306)/votebot_db")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// configure database
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(4)
	defer db.Close()

	// handle Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(db *sql.DB) {
		for _ = range c {
			os.Exit(0)
		}
	}(db)

	// objects for VK API
	var m Messages
	var p Photos

	// initialize long polling connection to VK API
	vkResp, err := InitLongPoll(conf.VkApiKey)
	if err != nil {
		log.Fatal(err)
	}

	var vkSessionAttrs LongPollAttr
	err = json.Unmarshal([]byte(vkResp), &vkSessionAttrs)

	// endless loop for VK events handling
	for {
		// get VK event
		event, err := vkSessionAttrs.GetEvent()
		if err != nil {
			log.Println(err)
			continue
		}
		vkSessionAttrs.Resp.Ts = event.Ts

		// handle VK group message
		message, err := GetVKGroupMessage(&event)
		if err != nil {
			log.Println(err)
			continue
		}

		// !important: mark message as read. there are bugs without this
		err = m.MarkAsRead(message.MessageId, conf.VkApiKey)
		if err != nil {
			log.Println(err)
			continue
		}

		if message.Text == "" {
			m.Send(message.FromId, "Ты не написал нам, кого оцениваешь! Приложи текстовое сообщение к своей фотографии.", nil, conf.VkApiKey)
			continue
		}

		// get image from message and emotion recognition
		vkMessage, err := vkSessionAttrs.GetImageByMessageId(message.MessageId, conf.VkApiKey)
		if err != nil {
			log.Println(err)
			continue
		}

		if len(vkMessage.Resps.Items[0].Attachments) == 0 {
			log.Println("Empty attachments")
			continue
		}

		emotions, err := GetEmotionByImageURL(vkMessage.Resps.Items[0].Attachments[0].Photo.Url, conf.EmotionApiKey)
		if err != nil {
			log.Println(err)
			continue
		}

		upServer, err := p.GetMessagesUploadServer(conf.VkApiKey)
		if err != nil {
			log.Println(err)
			continue
		}

		resp, err := http.Get(vkMessage.Resps.Items[0].Attachments[0].Photo.Url)
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()

		imageBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		// draw rectangle at picture
		inpBuffer := bytes.NewBuffer(imageBytes)
		if err != nil {
			log.Println(err)
			continue
		}

		rawImage, _ := jpeg.Decode(inpBuffer)

		cimg := image.NewRGBA(rawImage.Bounds())
		draw.Draw(cimg, rawImage.Bounds(), rawImage, image.Point{}, draw.Over)

		for i, val := range emotions {
			var max_k string
			for k, v := range val.Score {
				if max_k == "" {
					max_k = k
					continue
				}

				if v > val.Score[max_k] {
					max_k = k
				}
			}

			var c color.RGBA
			switch max_k {
			case "anger":
				c = color.RGBA{216, 51, 58, 255}
			case "contempt":
				c = color.RGBA{137, 129, 118, 255}
			case "disgust":
				c = color.RGBA{112, 141, 77, 255}
			case "fear":
				c = color.RGBA{139, 97, 169, 255}
			case "happiness":
				c = color.RGBA{247, 159, 36, 255}
			case "neutral":
				c = color.RGBA{240, 248, 255, 255}
			case "sadness":
				c = color.RGBA{39, 118, 185, 255}
			case "surprise":
				c = color.RGBA{237, 255, 33, 255}
			}
			DrawFaceRectangle(cimg, val.Face, c)
			DrawNumberOnImage(cimg, i+1, val.Face, c)
		}

		inpBuffer.Reset()
		jpeg.Encode(inpBuffer, cimg, nil)

		var new_anger, new_contempt, new_disgust, new_fear, new_happiness, new_neutral, new_sadness, new_surprise float64
		var emotions_count float64

		for _, val := range emotions {
			new_anger += val.Score["anger"]
			new_contempt += val.Score["contempt"]
			new_disgust += val.Score["disgust"]
			new_fear += val.Score["fear"]
			new_happiness += val.Score["happiness"]
			new_neutral += val.Score["neutral"]
			new_sadness += val.Score["sadness"]
			new_surprise += val.Score["surprise"]
			emotions_count++
		}

		select_rows, err := db.Query(`SELECT * FROM events WHERE name=?`, message.Text)
		if err != nil {
			log.Println(err)
			continue
		}

		if select_rows.Next() {
			var vote_numbers, anger, contempt, disgust, fear, happiness, neutral, sadness, surprise float64
			var name string
			var id int

			err = select_rows.Scan(&id, &name, &vote_numbers, &anger, &contempt, &disgust, &fear, &happiness, &neutral, &sadness, &surprise)
			if err != nil {
				log.Println(err)
				continue
			}

			// normalize emotions
			new_anger = (new_anger + anger) / (emotions_count + 1.0)
			new_contempt = (new_contempt + contempt) / (emotions_count + 1.0)
			new_disgust = (new_disgust + disgust) / (emotions_count + 1.0)
			new_fear = (new_fear + fear) / (emotions_count + 1.0)
			new_happiness = (new_happiness + happiness) / (emotions_count + 1.0)
			new_neutral = (new_neutral + neutral) / (emotions_count + 1.0)
			new_sadness = (new_sadness + sadness) / (emotions_count + 1.0)
			new_surprise = (new_surprise + surprise) / (emotions_count + 1.0)

			update_rows, err := db.Query(`UPDATE events SET vote_numbers=?, anger=?,
				contempt=?, disgust=?, fear=?, happiness=?, neutral=?, sadness=?,
				surprise=? 	WHERE name=?`, 1, new_anger, new_contempt, new_disgust,
				new_fear, new_happiness, new_neutral, new_sadness, surprise, name)
			if err != nil {
				log.Println(err)
				continue
			}

			err = update_rows.Err()
			if err != nil {
				log.Println(err)
				continue
			}

			update_rows.Close()
		} else {
			insert_rows, err := db.Query(`INSERT INTO events (name, vote_numbers, anger, contempt, disgust, fear, 
				happiness, neutral, sadness, surprise) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, message.Text, 1,
				new_anger/emotions_count, new_contempt/emotions_count, new_disgust/emotions_count,
				new_fear/emotions_count, new_happiness/emotions_count, new_neutral/emotions_count,
				new_sadness/emotions_count, new_surprise/emotions_count)
			if err != nil {
				log.Println(err)
				continue
			}

			err = insert_rows.Err()
			if err != nil {
				log.Println(err)
				continue
			}

			insert_rows.Close()
		}

		err = select_rows.Err()
		if err != nil {
			log.Println(err)
			continue
		}

		select_rows.Close()

		uploadedPhoto, err := p.SendPhotoToUploadServer(&upServer, inpBuffer)
		if err != nil {
			log.Println(err)
			continue
		}

		attachedPhoto, err := p.SaveMessagesPhoto(&uploadedPhoto, conf.VkApiKey)
		if err != nil {
			log.Println(err)
			continue
		}

		var s string
		for i, val := range emotions {
			s += strconv.Itoa(i+1) + ". " + fmt.Sprintln(val.Score)
		}

		// send message
		err = m.Send(message.FromId, s, &attachedPhoto, conf.VkApiKey)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
