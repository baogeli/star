package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient 封装 HTTP 请求客户端
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 创建新的 HTTP 客户端
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get 发送 GET 请求
func (hc *HTTPClient) Get(url string, headers map[string]string, cookies map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 设置 cookies
	for key, value := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  key,
			Value: value,
		})
	}

	// 发送请求
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ParseJSON 解析 JSON 响应
func ParseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}