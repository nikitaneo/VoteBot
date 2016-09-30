package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// Emotion recognition types
type EmotionResponce struct {
	Face  faceRectangle `json:"faceRectangle"`
	Score scores        `json:"scores"`
}

type faceRectangle struct {
	Height int `json:"height"`
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
}

type scores struct {
	Anger     float64 `json:"anger"`
	Contempt  float64 `json:"contempt"`
	Disgust   float64 `json:"disgust"`
	Fear      float64 `json:"fear"`
	Happiness float64 `json:"happiness"`
	Neutral   float64 `json:"neutral"`
	Sadness   float64 `json:"sadness"`
	Surprise  float64 `json:"surprise"`
}

func GetEmotionByImageURL(imageURL string) ([]EmotionResponce, error) {
	imageUrlBytes := []byte(`{ "url": "` + imageURL + `" }`)
	emotionReq, err := http.NewRequest("POST", EMOTION_API_URL, bytes.NewBuffer(imageUrlBytes))
	if err != nil {
		return []EmotionResponce{}, err
	}

	emotionReq.Header.Add("Ocp-Apim-Subscription-Key", EMOTION_KEY)
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
