package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	VK_PUBLIC_KEY   string = "449e8239db0cd5d912e8018aa12026f4c71f2319e18d09426d342927137452129a17186fba57cd66f8d76"
	EMOTION_API_URL string = "https://api.projectoxford.ai/emotion/v1.0/recognize"
	EMOTION_KEY     string = "0da704a311a8409289385b42e50fe1d3"
)

type Changeable interface {
	Set(x, y int, c color.Color)
}

// Functions block
func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	var RECT_COLOR color.RGBA = color.RGBA{255, 215, 0, 255}
	var m Messages
	var p Photos

	var ignore bool = false

	vkResp, err := InitLongPoll(VK_PUBLIC_KEY)
	check(err)

	var vkSessionAttrs LongPollAttr
	err = json.Unmarshal([]byte(vkResp), &vkSessionAttrs)

	for {
		event, err := vkSessionAttrs.GetEvent()
		check(err)

		message, err := GetVKGroupMessage(&event)

		if err == nil {
			vkMessage, err := vkSessionAttrs.GetImageByMessageId(message.MessageId, VK_PUBLIC_KEY)
			check(err)

			if !ignore {
				emotions, err := GetEmotionByImageURL(vkMessage.Resps.Items[0].Attachments[0].Photo.Url)
				if err != nil {
					fmt.Println(err)
				}

				fmt.Println(emotions)

				upServer, err := p.GetMessagesUploadServer(VK_PUBLIC_KEY)
				if err != nil {
					fmt.Println(err)
				}

				//!-------------------------------------------------------------------
				// draw rectangle at picture
				resp, err := http.Get(vkMessage.Resps.Items[0].Attachments[0].Photo.Url)
				if err != nil {
					fmt.Println(err)
				}
				defer resp.Body.Close()

				imageBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
				}

				inpBuffer := bytes.NewBuffer(imageBytes)
				if err != nil {
					fmt.Println(err)
				}

				rawImage, _ := jpeg.Decode(inpBuffer)

				cimg := image.NewRGBA(rawImage.Bounds())
				draw.Draw(cimg, rawImage.Bounds(), rawImage, image.Point{}, draw.Over)

				// Now you have cimg which contains the original image and is changeable
				// (it has a Set() method)

				for _, val := range emotions {
					for x := val.Face.Left; x < val.Face.Left+val.Face.Width; x++ {
						cimg.Set(x, val.Face.Top, RECT_COLOR)
						cimg.Set(x, val.Face.Top-1, RECT_COLOR)
						cimg.Set(x, val.Face.Top-2, RECT_COLOR)
					}
					for x := val.Face.Left; x < val.Face.Left+val.Face.Width; x++ {
						cimg.Set(x, val.Face.Top+val.Face.Height, RECT_COLOR)
						cimg.Set(x, val.Face.Top+val.Face.Height-1, RECT_COLOR)
						cimg.Set(x, val.Face.Top+val.Face.Height-2, RECT_COLOR)
					}
					for y := val.Face.Top; y < val.Face.Top+val.Face.Height; y++ {
						cimg.Set(val.Face.Left, y, RECT_COLOR)
						cimg.Set(val.Face.Left+1, y, RECT_COLOR)
						cimg.Set(val.Face.Left+2, y, RECT_COLOR)
					}
					for y := val.Face.Top; y < val.Face.Top+val.Face.Height; y++ {
						cimg.Set(val.Face.Left+val.Face.Width, y, RECT_COLOR)
						cimg.Set(val.Face.Left+val.Face.Width-1, y, RECT_COLOR)
						cimg.Set(val.Face.Left+val.Face.Width-2, y, RECT_COLOR)
					}
				}

				// And when saving, save 'cimg' of course:
				inpBuffer.Reset()
				jpeg.Encode(inpBuffer, cimg, nil)

				// end of drawing
				//!-------------------------------------------------------------------

				uploadedPhoto, err := p.SendPhotoToUploadServer(&upServer, inpBuffer)
				if err != nil {
					fmt.Println(err)
				}

				attachedPhoto, err := p.SaveMessagesPhoto(&uploadedPhoto, VK_PUBLIC_KEY)
				if err != nil {
					fmt.Println(err)
				}

				err = m.Send(message.FromId, message.Text+"!", &attachedPhoto, VK_PUBLIC_KEY)
				if err != nil {
					fmt.Println(err)
				}

				_ = emotions
			}

			if ignore {
				ignore = false
			} else {
				ignore = true
			}
		}
		vkSessionAttrs.Resp.Ts = event.Ts
	}
}
