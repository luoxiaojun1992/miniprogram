package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client defines the wechat API client interface.
type Client interface {
	Code2Session(ctx context.Context, code string) (openID string, err error)
}

type client struct {
	appID     string
	appSecret string
	httpClient *http.Client
}

// NewClient creates a new wechat Client.
func NewClient(appID, appSecret string) Client {
	return &client{
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{},
	}
}

type code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func (c *client) Code2Session(ctx context.Context, code string) (string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.appID, c.appSecret, code,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("wechat api request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	var result code2SessionResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return result.OpenID, nil
}
