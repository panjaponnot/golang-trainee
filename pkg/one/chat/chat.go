package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sale_ranking/pkg/requests"

	"github.com/labstack/echo"
)

const ProductionEndpoint = "https://chat-api.one.th/message/api/v1"

type Chat struct {
	BotId       string
	Token       string
	TokenType   string
	ApiEndpoint string
}

type ChatFriend struct {
	OneEmail    string `json:"one_email"`
	UserId      string `json:"user_id"`
	AccountId   string `json:"one_id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

type ChatProfile struct {
	Email          string `json:"email"`
	Nickname       string `json:"nickname"`
	AccountId      string `json:"one_id"`
	ProfilePicture string `json:"profilepicture"`
}

func NewChatBot(botId string, token string, tokenType string) Chat {
	return Chat{
		BotId:       botId,
		Token:       token,
		TokenType:   tokenType,
		ApiEndpoint: ProductionEndpoint,
	}
}

func (c *Chat) send(method string, url string, body []byte) (requests.Response, error) {
	headers := map[string]string{
		echo.HeaderContentType:   "application/json",
		echo.HeaderAuthorization: fmt.Sprintf("%s %s", c.TokenType, c.Token),
	}
	r, err := requests.Request(method, url, headers, bytes.NewBuffer(body), 0)
	if err != nil {
		return r, err
	}
	return r, nil
}

func (c *Chat) url(path string) string {
	return fmt.Sprintf("%s%s", c.ApiEndpoint, path)
}

func (c *Chat) PushTextMessage(to string, msg string, customNotify *string) error {
	pushMessage := struct {
		To           string `json:"to"`
		BotId        string `json:"bot_id"`
		Type         string `json:"type"`
		Message      string `json:"message"`
		CustomNotify string `json:"custom_notification,omitempty"`
	}{
		To:      to,
		BotId:   c.BotId,
		Type:    "text",
		Message: msg,
	}
	fmt.Println("==", c.BotId)
	if customNotify != nil {
		pushMessage.CustomNotify = *customNotify
	}
	body, _ := json.Marshal(&pushMessage)
	_, err := c.send(http.MethodPost, c.url("/push_message"), body)
	return err
}

func (c *Chat) GetChatProfile(oneChatToken string) (ChatProfile, error) {
	var chatProfile ChatProfile
	msg := struct {
		BotId        string `json:"bot_id"`
		OneChatToken string `json:"source"`
	}{
		BotId:        c.BotId,
		OneChatToken: oneChatToken,
	}
	body, _ := json.Marshal(&msg)
	r, err := c.send(http.MethodPost, "https://chat-api.one.th/manage/api/v1/getprofile", body)
	if err != nil {
		return chatProfile, err
	}
	chatProfileResult := struct {
		Data   ChatProfile `json:"data"`
		Status string      `json:"status"`
	}{}
	if err := json.Unmarshal(r.Body, &chatProfileResult); err != nil {
		return chatProfile, err
	}
	return chatProfileResult.Data, nil
}

func (c *Chat) PushWebView(to string, label string, path string, img string, title string, detail string, customNotify *string) error {
	type Choice struct {
		Label string `json:"label"`
		Type  string `json:"type"`
		Url   string `json:"url"`
		Size  string `json:"size"`
	}
	type Elements struct {
		Image   string   `json:"image"`
		Title   string   `json:"title"`
		Detail  string   `json:"detail"`
		Choices []Choice `json:"choice"`
	}
	pushMessage := struct {
		To           string     `json:"to"`
		BotId        string     `json:"bot_id"`
		Type         string     `json:"type"`
		CustomNotify string     `json:"custom_notification,omitempty"`
		Elements     []Elements `json:"elements"`
	}{
		To:    to,
		BotId: c.BotId,
		Type:  "template",
		Elements: []Elements{
			{
				Image:  img,
				Title:  title,
				Detail: detail,
				Choices: []Choice{
					{
						Label: label,
						Type:  "webview",
						Url:   path,
						Size:  "full",
					},
				},
			},
		},
	}

	if customNotify != nil {
		pushMessage.CustomNotify = *customNotify
	}
	body, _ := json.Marshal(&pushMessage)
	r, err := c.send(http.MethodPost, c.url("/push_message"), body)
	if err != nil {
		return err
	}
	if r.Code != 200 {
		return errors.New(fmt.Sprintf("server return error with http code %d : %s", r.Code, string(r.Body)))
	}
	return nil
}

func (c *Chat) PushLinkTemplate(to string, label string, path string, img string, title string, detail string, customNotify *string) error {
	type Choice struct {
		Label        string `json:"label"`
		Type         string `json:"type"`
		Url          string `json:"url"`
		OneChatToken string `json:"onechat_token"`
	}
	type Elements struct {
		Image   string   `json:"image"`
		Title   string   `json:"title"`
		Detail  string   `json:"detail"`
		Choices []Choice `json:"choice"`
	}
	pushMessage := struct {
		To           string     `json:"to"`
		BotId        string     `json:"bot_id"`
		Type         string     `json:"type"`
		CustomNotify string     `json:"custom_notification,omitempty"`
		Elements     []Elements `json:"elements"`
	}{
		To:    to,
		BotId: c.BotId,
		Type:  "template",
		Elements: []Elements{
			{
				Image:  img,
				Title:  title,
				Detail: detail,
				Choices: []Choice{
					{
						Label:        label,
						Type:         "link",
						Url:          path,
						OneChatToken: "true",
					},
				},
			},
		},
	}

	if customNotify != nil {
		pushMessage.CustomNotify = *customNotify
	}
	body, _ := json.Marshal(&pushMessage)
	r, err := c.send(http.MethodPost, c.url("/push_message"), body)
	if err != nil {
		return err
	}
	if r.Code != 200 {
		return errors.New(fmt.Sprintf("server return error with http code %d : %s", r.Code, string(r.Body)))
	}
	return nil
}
