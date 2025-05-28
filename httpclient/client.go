package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// HTTPClient 封装了带有重试功能的HTTP客户端
type HTTPClient struct {
	client      *http.Client
	maxRetries  int
	maxWaitTime time.Duration
	headers     map[string]string // 默认请求头
}

// Option 定义HTTP客户端的可选配置
type Option func(*HTTPClient)

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *HTTPClient) {
		c.client.Timeout = timeout
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(c *HTTPClient) {
		c.maxRetries = maxRetries
	}
}

// WithMaxWaitTime 设置最大等待时间
func WithMaxWaitTime(maxWaitTime time.Duration) Option {
	return func(c *HTTPClient) {
		c.maxWaitTime = maxWaitTime
	}
}

// WithDefaultHeaders 设置默认请求头
func WithDefaultHeaders(headers map[string]string) Option {
	return func(c *HTTPClient) {
		c.headers = headers
	}
}

// WithInsecureSkipVerify 设置是否跳过TLS证书验证
// 注意：仅在开发环境使用，生产环境应该使用正确的证书
func WithInsecureSkipVerify(skip bool) Option {
	return func(c *HTTPClient) {
		if c.client.Transport == nil {
			c.client.Transport = &http.Transport{}
		}
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig == nil {
				transport.TLSClientConfig = &tls.Config{}
			}
			transport.TLSClientConfig.InsecureSkipVerify = skip
		}
	}
}

// NewHTTPClient 创建一个新的HTTP客户端实例
func NewHTTPClient(opts ...Option) *HTTPClient {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // 自动从环境变量 http_proxy 和 https_proxy 中读取代理配置
	}
	client := &HTTPClient{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		maxRetries:  3,
		maxWaitTime: 1 * time.Minute,
		headers:     make(map[string]string),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// setHeaders 设置请求头
func (c *HTTPClient) setHeaders(req *http.Request, headers map[string]string) {
	// 设置默认请求头
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	// 设置自定义请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// closeBody 安全地关闭响应体
func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

// Do 执行HTTP请求，支持重试
func (c *HTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// 创建重试策略
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = c.maxWaitTime

	// 执行请求，支持重试
	operation := func() error {
		if resp != nil {
			closeBody(resp) // 关闭之前的响应体
		}

		resp, err = c.client.Do(req)
		if err != nil {
			return err
		}

		// 如果响应状态码大于等于500，标记为需要重试
		if resp.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("server error: status code %d", resp.StatusCode)
		}

		return nil
	}

	err = backoff.Retry(operation, backoff.WithMaxRetries(expBackoff, uint64(c.maxRetries)))
	if err != nil {
		closeBody(resp) // 如果发生错误，确保关闭响应体
		return nil, err
	}

	return resp, nil
}

// DoJSON 执行HTTP请求并解析JSON响应
func (c *HTTPClient) DoJSON(ctx context.Context, req *http.Request, v interface{}) error {
	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	return json.NewDecoder(resp.Body).Decode(v)
}

// Get 发送GET请求，支持URL参数和请求头
func (c *HTTPClient) Get(ctx context.Context, urlStr string, params map[string]string, headers map[string]string) (*http.Response, error) {
	// 构建URL with query parameters
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// 添加查询参数
	query := baseURL.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	baseURL.RawQuery = query.Encode()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	c.setHeaders(req, headers)

	return c.Do(ctx, req)
}

// GetJSON 发送GET请求并解析JSON响应
func (c *HTTPClient) GetJSON(ctx context.Context, urlStr string, params map[string]string, headers map[string]string, v interface{}) error {
	resp, err := c.Get(ctx, urlStr, params, headers)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	return json.NewDecoder(resp.Body).Decode(v)
}

// Post 发送POST请求，支持JSON body和请求头
func (c *HTTPClient) Post(ctx context.Context, urlStr string, body interface{}, headers map[string]string) (*http.Response, error) {
	// 将body转换为JSON
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置Content-Type为application/json
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}

	// 设置请求头
	c.setHeaders(req, headers)

	return c.Do(ctx, req)
}

// PostJSON 发送POST请求并解析JSON响应
func (c *HTTPClient) PostJSON(ctx context.Context, urlStr string, body interface{}, headers map[string]string, v interface{}) error {
	resp, err := c.Post(ctx, urlStr, body, headers)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	return json.NewDecoder(resp.Body).Decode(v)
}
