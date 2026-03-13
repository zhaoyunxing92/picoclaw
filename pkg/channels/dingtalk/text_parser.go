package dingtalk

import (
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
)

// MessageParser parser dingtalk message
type textMessageParser struct {
	data *chatbot.BotCallbackDataModel
}

func newTextMessageParser(data *chatbot.BotCallbackDataModel) MessageParser {
	return &textMessageParser{data: data}
}

func (text *textMessageParser) Parser() (content string, media []string) {
	return text.data.Text.Content, nil
}
