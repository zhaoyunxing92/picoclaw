package dingtalk

import (
	"errors"

	"github.com/sipeed/picoclaw/pkg/logger"
)

type Unmarshalled interface {
	CheckError() error

	GetMessage() string
}

type Response struct {
	Code      string `json:"code"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"requestid,omitempty"`
}

func (resp *Response) CheckError() error {
	if resp.Code != "" && resp.Message != "" {
		logger.ErrorCF("dingtalk", "dingtalk call error", map[string]any{
			"msg":       resp.Message,
			"code":      resp.Code,
			"requestId": resp.RequestID,
		})
		return errors.New(resp.Message)
	}
	return nil
}

func (resp *Response) GetMessage() string {
	return resp.Message
}

type AccessTokenResponse struct {
	Response
	Expires int64  `json:"expireIn"`
	Token   string `json:"accessToken"`
}

type CardCreateAndDeliverResponse struct {
	Response
	Result struct {
		DeliverResults []struct {
			SpaceID   string `json:"spaceId"`
			SpaceType string `json:"spaceType"`
			Success   bool   `json:"success"`
			CarrierID string `json:"carrierId"`
			ErrorMsg  string `json:"errorMsg"`
		} `json:"deliverResults"`
		OutTrackID string `json:"outTrackId"`
	} `json:"result"`
	Success bool `json:"success"`
}

func (resp *CardCreateAndDeliverResponse) CheckError() error {
	if (resp.Code != "" && resp.Message != "") || !resp.Success {
		logger.ErrorCF("dingtalk", "dingtalk call error", map[string]any{
			"error":     resp.Message,
			"code":      resp.Code,
			"requestId": resp.RequestID,
		})
		return errors.New(resp.Message)
	}
	return nil
}

type MessageFilesDownloadResponse struct {
	Response
	DownloadUrl string `json:"downloadUrl"`
}
