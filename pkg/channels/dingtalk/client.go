package dingtalk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"

	"github.com/sipeed/picoclaw/pkg/logger"
)

const (
	endpoint             = "https://api.dingtalk.com"
	accessToken          = "/v1.0/oauth2/accessToken"
	createCardAndDeliver = "/v1.0/card/instances/createAndDeliver"
	cardStreaming        = "/v1.0/card/streaming"
	privateChatMessages  = "/v1.0/robot/privateChatMessages/send"
	batchSendMessages    = "/v1.0/robot/oToMessages/batchSend"
	messageFilesDownload = "/v1.0/robot/messageFiles/download"
)

// MessageType represents the type of message to be sent to DingTalk. It can be one of the following:
// https://open.dingtalk.com/document/dingstart/types-of-messages-sent-by-robots#9c212e87342hn
type MessageType string

func (t MessageType) String() string {
	return string(t)
}

const (
	Markdown = MessageType("sampleMarkdown")
	Text     = MessageType("sampleText")
	Image    = MessageType("sampleImageMsg")
)

type Client struct {
	clientID     string
	clientSecret string
	token        string
	// card template id
	cardTemplateID string
	// card template message content default content
	cardTemplateContentKey string

	// robot code default use clientID
	robotCode string

	expires time.Time
	client  *http.Client
	mu      sync.Mutex // protects token and expires
}

// NewClient creates a new Client instance with the provided client ID, client secret, and optional configurations.
func NewClient(clientID, clientSecret string, opts ...ClientOption) *Client {
	client := &Client{
		clientID: clientID, clientSecret: clientSecret,
		client: &http.Client{Timeout: time.Second * 30},
	}
	for _, opt := range opts {
		opt(client)
	}
	if client.robotCode == "" {
		client.robotCode = client.clientID
	}
	if client.cardTemplateContentKey == "" {
		client.cardTemplateContentKey = "content"
	}
	return client
}

// GetToken retrieves the access token, refreshing it if it has expired.
func (c *Client) GetToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.expires) {
		return c.token, nil
	}
	data := map[string]any{
		"appKey":    c.clientID,
		"appSecret": c.clientSecret,
	}
	resp := &AccessTokenResponse{}
	if err := c.httpRequest(ctx, http.MethodPost, accessToken, data, resp); err != nil {
		return "", err
	}
	c.expires = time.Now().Add(time.Second * time.Duration(resp.Expires))
	c.token = resp.Token
	return c.token, nil
}

// BatchSendMessages sends a message to multiple users in a batch.
func (c *Client) BatchSendMessages(ctx context.Context, msgType MessageType, userIDs []string, content string) error {
	body := map[string]any{
		"robotCode": c.robotCode,
		"msgKey":    msgType,
		"userIds":   userIDs,
		"msgParam":  c.buildSendMessages(msgType, content),
	}
	return c.httpRequest(ctx, http.MethodPost, batchSendMessages, body, nil)
}

// CardStreaming updates the content of a card instance identified by cardInstanceID.
func (c *Client) CardStreaming(ctx context.Context, cardInstanceID, content string) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	body := map[string]any{
		"outTrackId": cardInstanceID,
		"guid":       id.String(),
		"key":        c.cardTemplateContentKey,
		"content":    content,
		"isFull":     true,
	}
	return c.httpRequest(ctx, http.MethodPut, cardStreaming, body, nil)
}

// CardCreateAndDeliver creates a card instance and delivers it to the user or group.
// It returns the outTrackID of the created card instance, which can be used for subsequent updates via CardStreaming.
// If the chatbot is in a group conversation, the card will be delivered to the group; otherwise, it will be delivered to the user.
// <a href="https://open.dingtalk.com/document/development/create-and-deliver-cards">Card Delivery API Documentation</a>
func (c *Client) CardCreateAndDeliver(ctx context.Context, chatbot *chatbot.BotCallbackDataModel) (string, error) {
	if c.cardTemplateID == "" {
		return "", errors.New("cardTemplateId is required")
	}
	var (
		group                   = chatbot.ConversationType == "2"
		openSpaceID             = "dtv1.card//IM_ROBOT." + chatbot.SenderStaffId
		imRobotOpenDeliverModel = map[string]any{
			"spaceType": "IM_ROBOT",
		}
		imGroupOpenSpaceModel = map[string]any{
			"supportForward": true,
		}
		imRobotOpenSpaceModel = map[string]any{
			"supportForward": true,
		}
		imGroupOpenDeliverModel = map[string]any{
			"robotCode": c.robotCode,
		}
	)

	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	if group {
		openSpaceID = "dtv1.card//IM_GROUP." + chatbot.ConversationId
		imRobotOpenDeliverModel = map[string]any{}
		imGroupOpenDeliverModel["robotCode"] = c.robotCode
	}

	body := map[string]any{
		"cardTemplateId":          c.cardTemplateID,
		"outTrackId":              id.String(),
		"cardData":                map[string]any{},
		"openSpaceId":             openSpaceID,
		"userIdType":              1,
		"imGroupOpenDeliverModel": imGroupOpenDeliverModel,
		"imGroupOpenSpaceModel":   imGroupOpenSpaceModel,
		"imRobotOpenSpaceModel":   imRobotOpenSpaceModel,
		"imRobotOpenDeliverModel": imRobotOpenDeliverModel,
	}
	resp := &CardCreateAndDeliverResponse{}
	return resp.Result.OutTrackID, c.httpRequest(ctx, http.MethodPost, createCardAndDeliver, body, resp)
}

// PrivateChatMessages sends a message to a user in a private chat.
func (c *Client) PrivateChatMessages(
	ctx context.Context,
	msgType MessageType,
	openConversationID, content string,
) error {
	body := map[string]any{
		"msgKey":             msgType,
		"msgParam":           c.buildSendMessages(msgType, content),
		"openConversationId": openConversationID,
		"robotCode":          c.robotCode,
	}
	return c.httpRequest(ctx, http.MethodPost, privateChatMessages, body, nil)
}

//MessageFilesDownload download message files
// https://open.dingtalk.com/document/development/download-the-file-content-of-the-robot-receiving-message
func (c *Client) MessageFilesDownload(downloadCode string) (string, error) {
	body := map[string]any{
		"downloadCode": downloadCode,
		"robotCode":    c.robotCode,
	}
	resp := &MessageFilesDownloadResponse{}
	return resp.DownloadUrl, c.httpRequest(context.Background(), http.MethodPost, messageFilesDownload, body, resp)
}

func (c *Client) buildSendMessages(msgType MessageType, content string) string {
	data := map[string]any{}
	switch msgType {
	case Markdown:
		data = map[string]any{
			"title": "PicoClaw",
			"text":  content,
		}
	case Text:
		data = map[string]any{
			"content": content,
		}
	}
	msg, _ := json.Marshal(data)
	return string(msg)
}

func (c *Client) httpRequest(ctx context.Context, method, path string, body any, resp Unmarshalled) error {
	var (
		err   error
		token string
		data  []byte
		hc    = c.client
		req   *http.Request
		res   *http.Response
	)
	if path == accessToken {
		token = ""
	} else {
		token, err = c.GetToken(ctx)
		if err != nil {
			return err
		}
	}
	url := endpoint + path
	if body != nil {
		data, _ = json.Marshal(body)
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(data))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Acs-Dingtalk-Access-Token", token)
	}
	if res, err = hc.Do(req); err != nil {
		logger.ErrorCF("dingtalk", "dingtalk http request error", map[string]any{
			"url":    url,
			"method": method,
		})
		return err
	}
	if resp == nil {
		return nil
	}
	defer res.Body.Close()
	if data, err = io.ReadAll(res.Body); err != nil {
		logger.ErrorCF("dingtalk", "dingtalk http response read error", map[string]any{
			"url":    url,
			"status": res.Status,
			"body":   string(data),
		})
		return err
	}
	if err = json.Unmarshal(data, resp); err != nil {
		logger.ErrorCF("dingtalk", "dingtalk http response Unmarshal error", map[string]any{
			"url":    url,
			"status": res.Status,
			"body":   string(data),
		})
		return err
	}
	return resp.CheckError()
}
