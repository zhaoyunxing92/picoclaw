package dingtalk

import (
	"fmt"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
)

// MessageParser parser dingtalk message
type richTextMessageParser struct {
	client *Client
	data   *chatbot.BotCallbackDataModel
}

func newRichTextMessageParser(client *Client, data *chatbot.BotCallbackDataModel) MessageParser {
	return &richTextMessageParser{client, data}
}

func (rich *richTextMessageParser) Parser() (content string, media []string) {
	richText := rich.data.Content.(map[string]interface{})["richText"].([]interface{})
	media = make([]string, 0)
	for _, text := range richText {
		data := text.(map[string]interface{})
		if dc, ok := data["downloadCode"]; ok {
			if download, err := rich.client.MessageFilesDownload(dc.(string)); err != nil {
				continue
			} else {
				media = append(media, download)
			}
		} else if msgRaw, ok := data["text"]; ok {
			msg := msgRaw.(string)
			if msg != "" && msg != "\n" {
				content += fmt.Sprintf("%s\n", msg)
			}
		}
	}
	//fmt.Println(string(marshal))
	return
}
