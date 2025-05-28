package cactus

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"go-cactus/model"
	"io/ioutil"
	"net/http"
	"net/url"
	pkcs "software.sslmate.com/src/go-pkcs12"
	"sort"
	"strings"
)

// PEMToECDSA pem转成*ecdsa.PrivateKey类型
func PEMToECDSA(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("PEM解码失败：无效的PEM结构")
	}
	// 处理标准ECDSA私钥
	if block.Type == "EC PRIVATE KEY" {
		return x509.ParseECPrivateKey(block.Bytes)
	}
	// 处理PKCS#8封装私钥
	if block.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		ecdsaKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("PKCS#8内容非ECDSA私钥")
		}
		return ecdsaKey, nil
	}

	return nil, fmt.Errorf("不支持的PEM类型: %s", block.Type)
}

// InitPrivateKey 加载私钥
func InitPrivateKey() *ecdsa.PrivateKey {
	// 读取 PKCS12 文件
	pfxData, err := ioutil.ReadFile(model.SIGN_PIRVATE_PATH) // 替换为你的 PKCS12 文件路径
	if err != nil {
		fmt.Printf("无法读取 PKCS12 文件: %v\n", err)
		return nil
	}
	// 解析 PKCS12 文件
	privateKey, _, _, err := pkcs.DecodeChain(pfxData, model.KEY_PASS)
	if err != nil {
		fmt.Printf("无法解析 PKCS12 文件: %v\n", err)
		return nil
	}
	// 将私钥转换为 PEM 格式
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		fmt.Printf("无法转换私钥为 PKCS8 格式: %v\n", err)
		return nil
	}
	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	pemData := pem.EncodeToMemory(pemBlock)
	// 将PEM格式转化*ecdsa.PrivateKey类型
	key, _ := PEMToECDSA(pemData)
	return key
}

// buildAuthorization 构造Authorization : api+ " " + AKId + ":" + Sign
func buildAuthorization(sign string) string {
	return fmt.Sprintf("api %s:%s", model.AK_ID, sign)
}

// buildContentToSign 构造签名体(uri可携带参数)
// 签名体例子：
// "GET\n" +
// "application/json\n" +
// "\n" +
// "application/json\n" +
// "Tue, 03 Mar 2020 12:26:57 GMT\n" +
// "x-api-key:X5SGmgTAoYaVw1t7oD2p82pHgf0eNNVw3wxYGgM2\n" +
// "x-api-nonce:36dbe33ed529455cb0638eef0f5f59e3\n" +
// "/custody/v1/api/wallets?{b_id=[4a3e2fb40faa4b9d94480559ac01e8de], coin_names=[BTC,LTC], hide_no_coin_wallet=[false], total_market_order=[0]}"
func buildContentToSign(method, uri, date, nonce string, body []byte) (string, error) {
	//先格式化URI
	formatURI, err := formatURIParameters(uri)
	if err != nil {
		return "", err
	}
	//签名体
	var ret string

	if method == http.MethodGet {
		ret = fmt.Sprintf("%s\napplication/json\n\napplication/json\n%s\nx-api-key:%s\nx-api-nonce:%s\n%s",
			method, date, model.API_KEY, nonce, formatURI)
	} else {
		var contentSHA string
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
			contentSHA = getContentSha256(body)
		} else {
			contentSHA = ""
		}
		ret = fmt.Sprintf("%s\napplication/json\n%s\napplication/json\n%s\nx-api-key:%s\nx-api-nonce:%s\n%s",
			method, contentSHA, date, model.API_KEY, nonce, formatURI)
	}

	return ret, nil
}

// formatURIParameters 格式化URI（可携带params参数），其中多个paramName按照字典顺序排序:
// 举例子：/custody/v1/api/wallets?{b_id=[4a3e2fb40faa4b9d94480559ac01e8de], coin_names=[BTC,LTC], hide_no_coin_wallet=[false], total_market_order=[0]}
func formatURIParameters(uriStr string) (string, error) {
	u, err := url.Parse(uriStr)
	if err != nil {
		return "", err
	}

	query := u.Query()
	if len(query) == 0 {
		return u.Path, nil // 无参数直接返回路径
	}

	// 提取并排序参数
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	params := make([]string, 0, len(keys))
	for _, k := range keys {
		values := query[k]
		param := fmt.Sprintf("%s=[%s]", k, strings.Join(values, ","))
		params = append(params, param)
	}

	// 构造格式化后的URI
	return fmt.Sprintf("%s?{%s}", u.Path, strings.Join(params, ", ")), nil
}

// getContentSha256 构造Content-SHA256:
// Content-SHA256:当请求方式为POST、PUT或PATCH时
// 添加Content-SHA256请求头对应的值，Content-SHA256=Base64SHA256(body)，body为请求体
func getContentSha256(body []byte) string {
	hasher := sha256.New()
	hasher.Write(body)
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
