package dingtalk

import (
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
)

// MessageParser parser dingtalk message
type videoMessageParser struct {
	client *Client
	data   *chatbot.BotCallbackDataModel
}

func newVideoMessageParser(client *Client, data *chatbot.BotCallbackDataModel) MessageParser {
	return &videoMessageParser{client, data}
}

func (video *videoMessageParser) Parser() (content string, media []string) {
	return video.data.Text.Content, nil
}
