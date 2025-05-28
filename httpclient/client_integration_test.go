package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试用例数据结构
type testCase struct {
	name           string
	server         *httptest.Server
	setupClient    func() *HTTPClient
	expectedStatus int
	expectedBody   interface{}
	expectedError  bool
}

// TestHTTPClientGet 测试GET请求
func TestHTTPClientGet(t *testing.T) {
	// 测试用例
	tests := []testCase{
		{
			name: "成功的GET请求",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": "success"})
			})),
			setupClient: func() *HTTPClient {
				return NewHTTPClient(
					WithTimeout(5*time.Second),
					WithMaxRetries(1),
				)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "服务器错误带重试",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})),
			setupClient: func() *HTTPClient {
				return NewHTTPClient(
					WithTimeout(1*time.Second),
					WithMaxRetries(2),
					WithMaxWaitTime(2*time.Second),
				)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name: "带查询参数的GET请求",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "test_value", r.URL.Query().Get("test_param"))
				w.WriteHeader(http.StatusOK)
			})),
			setupClient: func() *HTTPClient {
				return NewHTTPClient()
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.server.Close()

			client := tt.setupClient()
			params := map[string]string{"test_param": "test_value"}
			headers := map[string]string{"X-Test": "test"}

			resp, err := client.Get(context.Background(), tt.server.URL, params, headers)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			defer resp.Body.Close()
		})
	}
}

// TestHTTPClientPost 测试POST请求
func TestHTTPClientPost(t *testing.T) {
	tests := []testCase{
		{
			name: "成功的POST请求",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, "test_value", body["test_key"])

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "success"})
			})),
			setupClient: func() *HTTPClient {
				return NewHTTPClient()
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "超时的POST请求",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			})),
			setupClient: func() *HTTPClient {
				return NewHTTPClient(
					WithTimeout(1*time.Second),
					WithMaxRetries(1),
				)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.server.Close()

			client := tt.setupClient()
			body := map[string]interface{}{"test_key": "test_value"}
			headers := map[string]string{"X-Test": "test"}

			resp, err := client.Post(context.Background(), tt.server.URL, body, headers)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			defer resp.Body.Close()
		})
	}
}

// TestHTTPClientRetry 测试重试机制
func TestHTTPClientRetry(t *testing.T) {
	// 记录请求次数
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(
		WithMaxRetries(3),
		WithMaxWaitTime(5*time.Second),
	)

	resp, err := client.Get(context.Background(), server.URL, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, requestCount) // 验证实际重试了2次
	defer resp.Body.Close()
}
