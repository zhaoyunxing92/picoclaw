package dingtalk

import (
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
)

// MessageParser parser dingtalk message
type pictureMessageParser struct {
	client *Client
	data   *chatbot.BotCallbackDataModel
}

func newPictureMessageParser(client *Client, data *chatbot.BotCallbackDataModel) MessageParser {
	return &pictureMessageParser{client, data}
}

func (picture *pictureMessageParser) Parser() (content string, media []string) {
	return picture.data.Text.Content, nil
}
