package cosutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

// Client wraps Tencent COS official SDK operations used by upload service.
type Client struct {
	client        *cos.Client
	baseURL       string
	bucket        string
	secretID      string
	secretKey     string
	publicBaseURL string
}

// NewClient builds COS client from endpoint and bucket.
func NewClient(endpoint, publicBaseURL, bucket, secretID, secretKey string) (*Client, error) {
	base := strings.TrimRight(endpoint, "/")
	if base == "" || strings.TrimSpace(bucket) == "" {
		return nil, fmt.Errorf("invalid cos endpoint or bucket")
	}
	if publicBaseURL == "" {
		publicBaseURL = base
	}
	bucketURL, err := url.Parse(fmt.Sprintf("%s/%s", base, url.PathEscape(bucket)))
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport
	httpClient := &http.Client{Timeout: 30 * time.Second}
	if strings.TrimSpace(secretID) != "" && strings.TrimSpace(secretKey) != "" {
		httpClient.Transport = &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
			Transport: transport,
		}
	}
	return &Client{
		client:        cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, httpClient),
		baseURL:       base,
		bucket:        bucket,
		secretID:      secretID,
		secretKey:     secretKey,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}, nil
}

func (c *Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	key = normalizeObjectKey(key)
	if key == "" {
		return "", fmt.Errorf("invalid object key")
	}
	_, err := c.client.Object.Put(ctx, key, body, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		return "", err
	}
	return c.ObjectURL(key), nil
}

func (c *Client) PresignPutURL(key string, expiresIn int) string {
	key = normalizeObjectKey(key)
	if key == "" {
		return ""
	}
	if c.secretID == "" || c.secretKey == "" {
		return fmt.Sprintf("%s/%s/%s?expires_in=%d", c.baseURL, url.PathEscape(c.bucket), escapeObjectKey(key), expiresIn)
	}
	u, err := c.client.Object.GetPresignedURL(context.Background(), http.MethodPut, key, c.secretID, c.secretKey, time.Duration(expiresIn)*time.Second, nil)
	if err != nil {
		return ""
	}
	return u.String()
}

func (c *Client) PresignGetURL(key string, expiresIn int) string {
	key = normalizeObjectKey(key)
	if key == "" {
		return ""
	}
	if c.secretID == "" || c.secretKey == "" {
		return fmt.Sprintf("%s/%s/%s?expires_in=%d", c.baseURL, url.PathEscape(c.bucket), escapeObjectKey(key), expiresIn)
	}
	u, err := c.client.Object.GetPresignedURL(context.Background(), http.MethodGet, key, c.secretID, c.secretKey, time.Duration(expiresIn)*time.Second, nil)
	if err != nil {
		return ""
	}
	return u.String()
}

func (c *Client) ObjectURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", c.publicBaseURL, url.PathEscape(c.bucket), escapeObjectKey(key))
}

func (c *Client) IsStaticMediaObject(ctx context.Context, key string) (bool, error) {
	key = normalizeObjectKey(key)
	if key == "" {
		return false, fmt.Errorf("invalid object key")
	}
	resp, err := c.client.Object.Head(ctx, key, nil)
	if err != nil {
		return false, err
	}
	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/"), nil
}

func normalizeObjectKey(key string) string {
	raw := strings.TrimSpace(key)
	if raw == "" || strings.Contains(raw, "..") {
		return ""
	}
	clean := path.Clean("/" + raw)
	if clean == "/" {
		return ""
	}
	clean = strings.TrimPrefix(clean, "/")
	if strings.Contains(clean, "..") {
		return ""
	}
	return clean
}

func escapeObjectKey(key string) string {
	parts := strings.Split(key, "/")
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		escaped = append(escaped, url.PathEscape(part))
	}
	return strings.Join(escaped, "/")
}
