package dingtalk

import (
	"errors"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
)

// MessageParser parser dingtalk message
type MessageParser interface {
	Parser() (content string, media []string)
}

// NewMessageParser https://open.dingtalk.com/document/dingstart/dingstart-robot-receive-message
func NewMessageParser(client *Client, data *chatbot.BotCallbackDataModel) (MessageParser, error) {
	switch data.Msgtype {
	case "text":
		return newTextMessageParser(data), nil
	case "richText":
		return newRichTextMessageParser(client, data), nil
	case "picture":
		return newPictureMessageParser(client, data), nil
	case "audio":
		return newAudioMessageParser(client, data), nil
	case "video":
		return newVideoMessageParser(client, data), nil
	case "file":
		return newFileMessageParser(client, data), nil
	default:
		return nil, errors.New("unknown message type " + data.Msgtype)
	}
}
