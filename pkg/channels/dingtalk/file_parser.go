package dingtalk

import (
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
)

// MessageParser parser dingtalk message
type fileMessageParser struct {
	client *Client
	data   *chatbot.BotCallbackDataModel
}

func newFileMessageParser(client *Client, data *chatbot.BotCallbackDataModel) MessageParser {
	return &fileMessageParser{client, data}
}

func (file *fileMessageParser) Parser() (content string, media []string) {
	return file.data.Text.Content, nil
}
