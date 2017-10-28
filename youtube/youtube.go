package youtube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var pageToken = ""
var startupTime = time.Now()

type Client struct {
	ChannelID         string
	ChatID            string
	ApiKey            string
	accessToken       string
	accessTokenExpiry time.Time
	RefreshToken      string
	ClientID          string
	ClientSecret      string
	MessageHandlers   []func(msg Message)
}

func (yt *Client) AddMessageHandler(callback func(msg Message)) {
	yt.MessageHandlers = append(yt.MessageHandlers, callback)
}
func (yt *Client) Start() {
	log.Println("Registered", len(yt.MessageHandlers), "message handlers.")
	yt.refreshAccessToken()

	for {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/liveChat/messages?liveChatId=%s&part=id,snippet,authorDetails&key=%s", yt.ChatID, yt.ApiKey)
		if pageToken != "" {
			url += "&pageToken=" + pageToken
		}
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Fatal("Error: ", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var data messageResponse

		if err := json.Unmarshal(body, &data); err != nil {
			panic(err)
		}

		pageToken = data.NextPageToken

		for _, message := range data.Items {
			published, err := time.Parse(time.RFC3339, message.Snippet.PublishedAt)
			if err != nil {
				log.Fatal(err)
			}
			if !published.Before(startupTime) && message.AuthorDetails.ChannelId != yt.ChannelID && message.Snippet.Type == "textMessageEvent" {
				message.yt = yt

				for _, handler := range yt.MessageHandlers {
					handler(message)
				}
			}
		}
		time.Sleep(time.Duration(data.PollingIntervalMillis) * time.Millisecond)
	}
}

func (client *Client) SendMessage(chatID, content string) (Message, error) {
	data := []byte(`{
	"snippet": {
		"type": "textMessageEvent",
		"liveChatId": "` + chatID + `",
		"textMessageDetails": {
			"messageText": "` + content + `"
		}
	}
}`)

	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/liveChat/messages?part=snippet&key=%s", client.ApiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+client.getAccessToken())
	req.Header.Set("Content-Type", "application/json")

	hClient := &http.Client{}
	resp, err := hClient.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var msg Message

	if resp.StatusCode != 200 {
		log.Println("Failed to send message")
		fmt.Println("response Body:", string(body))
		return msg, errors.New(string(body))
	}

	if err := json.Unmarshal(body, &msg); err != nil {
		panic(err)
	}
	msg.yt = client
	return msg, nil
}

func (client *Client) DeleteMessage(msgID string) bool {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/liveChat/messages?id=%s&key=%s", msgID, client.ApiKey)

	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+client.getAccessToken())

	hClient := &http.Client{}
	resp, err := hClient.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		log.Println("Failed to delete message", msgID)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
		return false
	}
	return true

}

type accessTokenResponse struct {
	Access_token string
	expires_in   int
}

func (client *Client) getAccessToken() string {
	if client.accessTokenExpiry.Before(time.Now()) {
		client.refreshAccessToken()
	}
	return client.accessToken
}

func (client *Client) refreshAccessToken() {
	apiUrl := "https://www.googleapis.com"
	resource := "/oauth2/v4/token"
	data := url.Values{}
	data.Set("client_id", client.ClientID)
	data.Add("client_secret", client.ClientSecret)
	data.Add("refresh_token", client.RefreshToken)
	data.Add("grant_type", "refresh_token")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	hClient := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode())) // <-- URL-encoded payload
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := hClient.Do(r)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Println(string(body))
		log.Fatal("Error: ", resp.StatusCode)
	}

	var responseData accessTokenResponse

	if err := json.Unmarshal(body, &responseData); err != nil {
		panic(err)
	}
	client.accessToken = responseData.Access_token
	client.accessTokenExpiry = time.Now().Add(time.Duration(responseData.expires_in) * time.Second)
}
