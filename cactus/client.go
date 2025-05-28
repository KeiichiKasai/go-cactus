package cactus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"

	"go-cactus/httpclient"
	"go-cactus/model"

	"github.com/google/uuid"
)

// Client 定义与Cactus API交互的接口
type Client interface {
	// CheckAddress 检验地址是否合法
	CheckAddress(ctx context.Context, req *model.CheckAddressReq) (*model.CheckAddressResp, error)
	// CreateOrder 创建提币订单
	CreateOrder(ctx context.Context, req *model.CreateOrderReq) (*model.CreateOrderResp, error)
	// TxDetail 查询钱包记录明细
	TxDetail(ctx context.Context, req *model.TxDetailReq) (*model.TxDetailResp, error)
	// TxSummary 查询钱包交易记录概要
	TxSummary(ctx context.Context, req *model.TxSummaryReq) (*model.TxSummaryResp, error)
	// GetAddressList 获取该钱包所有地址
	GetAddressList(ctx context.Context, req *model.GetAddressesReq) (*model.GetAddressesResp, error)

	// GetPublicIP 获取当前的公共 IP 地址（在白名单内的IP才可以访问Cactus）
	GetPublicIP(ctx context.Context) (string, error)
}

// ClientImpl 实现了Client接口
type ClientImpl struct {
	baseURL    string                 //第三方api所在URL
	privateKey *ecdsa.PrivateKey      //私钥
	client     *httpclient.HTTPClient //客户端
}

// NewClient 创建一个新的Cactus客户端
func NewClient() Client {
	return &ClientImpl{
		baseURL:    model.URL_PRE,
		privateKey: InitPrivateKey(),
		client: httpclient.NewHTTPClient(
			httpclient.WithTimeout(30*time.Second),
			httpclient.WithMaxRetries(3),
			httpclient.WithInsecureSkipVerify(true),
		),
	}
}

// Sign 进行签名
func (c *ClientImpl) Sign(content string) string {
	hashed := sha256.Sum256([]byte(content))
	r, s, _ := ecdsa.Sign(rand.Reader, c.privateKey, hashed[:])

	// 将r和s转换为ASN.1 DER格式
	type ecdsaSignature struct {
		R, S *big.Int
	}
	sigAsn1, _ := asn1.Marshal(ecdsaSignature{r, s})

	return base64.StdEncoding.EncodeToString(sigAsn1)
}

// buildRequest 构造请求
func (c *ClientImpl) buildRequest(ctx context.Context, method, uri string, body []byte) ([]byte, error) {
	//0.生成唯一标识
	date := time.Now().UTC().Format(model.TimeFormat)
	nonce := uuid.New().String()

	//1.构造签名体并进行签名
	signContent, err := buildContentToSign(method, uri, date, nonce, body)
	if err != nil {
		return nil, err
	}
	sign := c.Sign(signContent)

	//2.构造Authorization
	auth := buildAuthorization(sign)

	//3.生成一个请求头
	req, err := http.NewRequest(method, c.baseURL+uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	headers := req.Header

	//4.把相应信息放入请求头
	headers.Set("x-api-key", model.API_KEY)
	headers.Set("x-api-nonce", nonce)
	headers.Set("Accept", "application/json")
	headers.Set("Date", date)
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", auth)
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		headers.Set("Content-SHA256", getContentSha256(body))
	}

	//5.发送请求
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return respBody, nil
}

// CheckAddress 检验地址是否合法
func (c *ClientImpl) CheckAddress(ctx context.Context, req *model.CheckAddressReq) (*model.CheckAddressResp, error) {
	uri := "/custody/v1/api/addresses/type/check"
	body, err := json.Marshal(*req)
	if err != nil {
		return nil, errors.New("json marshal fail")
	}

	resp, err := c.buildRequest(ctx, http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	var result model.CheckAddressResp
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.New("json unmarshal fail")
	}

	return &result, nil
}

// CreateOrder 创建提币订单
func (c *ClientImpl) CreateOrder(ctx context.Context, req *model.CreateOrderReq) (*model.CreateOrderResp, error) {
	uri := fmt.Sprintf("/custody/v1/api/projects/%s/order/create", model.Bid)
	body, err := json.Marshal(*req)
	if err != nil {
		return nil, errors.New("json marshal fail")
	}
	resp, err := c.buildRequest(ctx, http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}
	var result model.CreateOrderResp
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.New("json unmarshal fail")
	}
	return &result, nil
}

// TxDetail 查询钱包记录明细
func (c *ClientImpl) TxDetail(ctx context.Context, req *model.TxDetailReq) (*model.TxDetailResp, error) {
	uri := fmt.Sprintf("/custody/v1/api/projects/%s/wallets/%s/tx-details?tx_types=WITHDRAW,DEPOSIT&id=%d", req.BID, req.WalletCode, req.ID) //按需调整参数
	resp, err := c.buildRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	var result model.TxDetailResp
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.New("json unmarshal fail")
	}
	return &result, nil
}

// TxSummary 查询钱包交易记录概要
func (c *ClientImpl) TxSummary(ctx context.Context, req *model.TxSummaryReq) (*model.TxSummaryResp, error) {
	uri := fmt.Sprintf("/custody/v1/api/projects/%s/wallets/%s/tx-summaries", model.Bid, model.ETHWallet)
	body, err := json.Marshal(*req)
	if err != nil {
		return nil, errors.New("json marshal fail")
	}
	resp, err := c.buildRequest(ctx, http.MethodGet, uri, body)
	if err != nil {
		return nil, err
	}
	var result model.TxSummaryResp
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.New("json unmarshal fail")
	}
	return &result, nil
}

// GetAddressList 获取该钱包所有地址
func (c *ClientImpl) GetAddressList(ctx context.Context, req *model.GetAddressesReq) (*model.GetAddressesResp, error) {
	uri := fmt.Sprintf("/custody/v1/api/projects/%s/wallets/%s/addresses", model.Bid, model.ETHWallet)
	body, err := json.Marshal(*req)
	if err != nil {
		return nil, errors.New("json marshal fail")
	}
	resp, err := c.buildRequest(ctx, http.MethodGet, uri, body)
	if err != nil {
		return nil, err
	}
	var result model.GetAddressesResp
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.New("json unmarshal fail")
	}

	return &result, nil
}

// GetPublicIP 获取当前的公共 IP 地址（在白名单内的IP才可以访问Cactus）
func (c *ClientImpl) GetPublicIP(ctx context.Context) (string, error) {
	// 构造请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ipconfig.io", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请求
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	return string(body), nil
}
