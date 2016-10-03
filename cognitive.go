package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

const EMOTION_API_URL string = "https://api.projectoxford.ai/emotion/v1.0/recognize"

// Emotion recognition types
type EmotionResponce struct {
	Face  FaceRectangle      `json:"faceRectangle"`
	Score map[string]float64 `json:"scores"`
}

type FaceRectangle struct {
	Height int `json:"height"`
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
}

type EmotionPair struct {
	emotion string
	value   float64
}

// type scores struct {
// 	Anger     float64 `json:"anger"` #d8333a
// 	Contempt  float64 `json:"contempt"` #898176
// 	Disgust   float64 `json:"disgust"` #708d4d
// 	Fear      float64 `json:"fear"` #8b61a3
// 	Happiness float64 `json:"happiness"` #f79f24
// 	Neutral   float64 `json:"neutral"` #f0f8ff
// 	Sadness   float64 `json:"sadness"` #2776b9
// 	Surprise  float64 `json:"surprise"` #edff21
// }

func GetEmotionByImageURL(imageURL string, accessToken string) ([]EmotionResponce, error) {
	imageUrlBytes := []byte(`{ "url": "` + imageURL + `" }`)
	emotionReq, err := http.NewRequest("POST", EMOTION_API_URL, bytes.NewBuffer(imageUrlBytes))
	if err != nil {
		return []EmotionResponce{}, err
	}

	emotionReq.Header.Add("Ocp-Apim-Subscription-Key", accessToken)
	emotionResp, err := http.DefaultClient.Do(emotionReq)
	if err != nil {
		return []EmotionResponce{}, err
	}

	message, err := ioutil.ReadAll(emotionResp.Body)
	if err != nil {
		return []EmotionResponce{}, err
	}
	emotionResp.Body.Close()

	// sometimes cognitive services returns empty string
	if "[]" != string(message) {
		var e []EmotionResponce
		err = json.Unmarshal(message, &e)
		if err != nil {
			return []EmotionResponce{}, err
		}
		return e, nil
	}

	return []EmotionResponce{}, errors.New("Empty responce")
}
