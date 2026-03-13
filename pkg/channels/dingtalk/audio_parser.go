package dingtalk

import (
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
)

// MessageParser parser dingtalk message
type audioMessageParser struct {
	client *Client
	data   *chatbot.BotCallbackDataModel
}

func newAudioMessageParser(client *Client, data *chatbot.BotCallbackDataModel) MessageParser {
	return &audioMessageParser{client, data}
}

func (audio *audioMessageParser) Parser() (content string, media []string) {
	return audio.data.Text.Content, nil
}
