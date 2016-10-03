package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	apiURL string = "https://api.vk.com/method/"
)

type Photos int
type Messages int

type AttachPhoto struct {
	Responses []attachPhotoResponse `json:"response"`
}

type attachPhotoResponse struct {
	Id       int    `json:"id"`
	AlbumId  int    `json:"album_id"`
	OwnerId  int    `json:"owner_id"`
	UserId   int    `json:"user_id"`
	Photo75  string `json:"photo_75"`
	Photo130 string `json:"photo_130"`
	Photo604 string `json:"photo_604"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Text     string `json:"text"`
	Date     int64  `json:"Date"`
}

type UploadedPhoto struct {
	Server int    `json:"server"`
	Photo  string `json:"photo"`
	Hash   string `json:"hash"`
}

type MessagesUploadServer struct {
	Response messagesUploadServerResponse `json:"response"`
}

type messagesUploadServerResponse struct {
	UploadURL string `json:"upload_url"`
	AlbumId   int64  `json:"album_id"`
	GroupId   int64  `json:"group_id"`
}

// LongPoll server attributes
type LongPollAttr struct {
	Resp longPollResp `json:"response"`
}

type longPollResp struct {
	Key    string `json:"key"`
	Server string `json:"server"`
	Ts     int64  `json:"ts"`
}

type jsonBody struct {
	Failed  int64           `json:"failed"`
	Ts      int64           `json:"ts"`
	Updates [][]interface{} `json:"updates"`
}

// VK group  messages types
type VKGroupMessage struct {
	MessageId   int
	Flags       int
	FromId      int
	Timestamp   int
	Subject     string
	Text        string
	Attachments map[string]interface{}
}

// VK private messages types
type VKMessage struct {
	Resps VKMessageResponse `json:"response"`
}

type VKMessageResponse struct {
	Items []VKMessageItems `json:"items"`
}

type VKMessageItems struct {
	Id          int                   `json:"id"`
	Attachments []VKMessageAttachment `json:"attachments"`
}

type VKMessageAttachment struct {
	Type  string         `json:"type"`
	Photo VKMessagePhoto `json:"photo"`
}

type VKMessagePhoto struct {
	Url string `json:"photo_604"`
}

func InitLongPoll(tocken string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET",
		"https://api.vk.com/method/messages.getLongPollServer?v=5.41&access_token="+
			tocken+"&use_ssl=0&need_ptc=0", nil)

	vkResp, err := client.Do(req)
	if err != nil {
		log.Println("InitLongPoll function error: ", err)
		return "", err
	}
	defer vkResp.Body.Close()

	vkMess, err := ioutil.ReadAll(vkResp.Body)
	if err != nil {
		log.Println("InitLongPoll function error: ", err)
		return "", err
	}

	log.Println("InitLongPoll function success: ", string(vkMess))
	return string(vkMess), nil
}

func (sessionAttrs *LongPollAttr) GetEvent() (jsonBody, error) {
	resp, err := http.Get("https://" + sessionAttrs.Resp.Server + "?act=a_check&key=" +
		sessionAttrs.Resp.Key + "&ts=" + fmt.Sprintf("%d", sessionAttrs.Resp.Ts) + "&wait=25&mode=2 ")
	if err != nil {
		log.Println("LongPollAttr.GetEvent function error: ", err)
		return jsonBody{}, err
	}
	defer resp.Body.Close()

	message, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("LongPollAttr.GetEvent function error: ", err)
		return jsonBody{}, err
	}

	var body jsonBody
	err = json.Unmarshal(message, &body)
	if err != nil {
		log.Println("LongPollAttr.GetEvent function error: ", err)
		return jsonBody{}, err
	}

	return body, nil
}

func GetVKGroupMessage(message *jsonBody) (VKGroupMessage, error) {
	// timeout elapsed
	if len(message.Updates) == 0 {
		return VKGroupMessage{}, errors.New("Timeout elapsed")
	}

	messageType, ok := message.Updates[0][0].(float64)

	if !ok {
		return VKGroupMessage{}, errors.New("Convertion error")
	}

	if 4 == messageType {
		vkMessage := VKGroupMessage{
			MessageId:   int(message.Updates[0][1].(float64)),
			Flags:       int(message.Updates[0][2].(float64)),
			FromId:      int(message.Updates[0][3].(float64)),
			Timestamp:   int(message.Updates[0][4].(float64)),
			Subject:     message.Updates[0][5].(string),
			Text:        message.Updates[0][6].(string),
			Attachments: message.Updates[0][7].(map[string]interface{}),
		}
		return vkMessage, nil
	} else {
		return VKGroupMessage{}, errors.New("Not message event")
	}
}

func (sessionAttrs *LongPollAttr) GetImageByMessageId(MessageId int, accessToken string) (VKMessage, error) {
	resp, err := http.Get(apiURL + "messages.getById?v=5.41&access_token=" +
		accessToken + "&message_ids=" + strconv.Itoa(MessageId))
	if err != nil {
		return VKMessage{}, err
	}
	defer resp.Body.Close()

	messageInfo, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return VKMessage{}, err
	}

	var vkMessageResp VKMessage
	err = json.Unmarshal(messageInfo, &vkMessageResp)
	if err != nil {
		return VKMessage{}, err
	}

	return vkMessageResp, nil
}

func (photo Photos) GetMessagesUploadServer(accessToken string) (MessagesUploadServer, error) {
	resp, err := http.Get(apiURL + "photos.getMessagesUploadServer?v=5.56&access_token=" + accessToken)
	if err != nil {
		return MessagesUploadServer{}, err
	}

	jsonStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return MessagesUploadServer{}, err
	}

	var upServer MessagesUploadServer
	err = json.Unmarshal(jsonStr, &upServer)
	if err != nil {
		return MessagesUploadServer{}, err
	}

	return upServer, nil
}

func (photo Photos) SendPhotoToUploadServer(upServer *MessagesUploadServer, img io.Reader) (UploadedPhoto, error) {
	var buffer bytes.Buffer
	multWriter := multipart.NewWriter(&buffer)

	fileWriter, err := multWriter.CreateFormFile("photo", "photo.jpg")
	if err != nil {
		return UploadedPhoto{}, err
	}

	if _, err = io.Copy(fileWriter, img); err != nil {
		return UploadedPhoto{}, err
	}
	multWriter.Close()

	resp, err := http.DefaultClient.Post(upServer.Response.UploadURL, multWriter.FormDataContentType(), &buffer)
	if err != nil {
		return UploadedPhoto{}, err
	}
	defer resp.Body.Close()

	jsonMessage, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return UploadedPhoto{}, nil
	}

	var p UploadedPhoto
	err = json.Unmarshal(jsonMessage, &p)
	if err != nil {
		return UploadedPhoto{}, err
	}

	return p, nil
}

func (p Photos) SaveMessagesPhoto(photo *UploadedPhoto, accessToken string) (AttachPhoto, error) {
	resp, err := http.Get(apiURL + "photos.saveMessagesPhoto?v=5.56&access_token=" +
		accessToken + "&photo=" + photo.Photo + "&server=" + strconv.Itoa(photo.Server) +
		"&hash=" + photo.Hash)

	if err != nil {
		return AttachPhoto{}, err
	}
	defer resp.Body.Close()

	jsonMessage, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return AttachPhoto{}, err
	}

	var savedPhoto AttachPhoto
	err = json.Unmarshal(jsonMessage, &savedPhoto)
	if err != nil {
		return AttachPhoto{}, err
	}

	return savedPhoto, nil
}

func (m Messages) Send(userId int, mess string, attachedPhoto *AttachPhoto, accessToken string) error {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	attachment := ""
	if attachedPhoto != nil {
		attachment = "&attachment=photo" + strconv.Itoa(attachedPhoto.Responses[0].OwnerId) +
			"_" + strconv.Itoa(attachedPhoto.Responses[0].Id)
	}
	resp, err := http.DefaultClient.Get(apiURL + "messages.send?v=5.56&access_token=" +
		accessToken + "&user_id=" + strconv.Itoa(userId) + "&message=" + url.QueryEscape(mess) +
		"&random_id=" + strconv.Itoa(random.Int()) + attachment)
	if err != nil {
		log.Println("Messages.Send function error: ", err)
		return err
	}
	defer resp.Body.Close()

	message, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Messages.Send function error: ", err)
		return err
	}

	log.Println("Messages.Send function success: ", string(message))
	return nil
}

func (m Messages) MarkAsRead(message_id int, accessToken string) error {
	resp, err := http.DefaultClient.Get(apiURL + "messages.markAsRead?v=5.56&access_token=" +
		accessToken + "&message_ids=" + strconv.Itoa(message_id))
	if err != nil {
		log.Println("Messages.MarkAsRead function error: ", err)
		return err
	}
	defer resp.Body.Close()

	message, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Messages.MarkAsRead function error: ", err)
		return err
	}

	log.Println("Messages.MarkAsRead function success: ", string(message))
	return nil
}
