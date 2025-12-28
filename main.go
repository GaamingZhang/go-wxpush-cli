package main

import (
	"encoding/json" // JSON编码和解码包，用于处理JSON格式的数据
	"flag"          // 命令行参数解析包，用于解析程序启动时的命令行参数
	"fmt"           // 格式化I/O包，用于字符串格式化和输出
	"io"            // 基础I/O接口包，提供输入输出的基本接口
	"net/http"      // HTTP客户端包，用于发送HTTP请求
	"os"            // 操作系统接口包，用于退出程序
	"strings"       // 字符串操作包，提供字符串处理函数
)

var Version = "dev"

// RequestParams 定义了请求参数的结构体
// 用于存储命令行参数
type RequestParams struct {
	AppID      string // 微信公众AppID
	Secret     string // 微信公众AppSecret
	UserID     string // 接收消息的用户的UserID
	TemplateID string // 微信模板消息的模板ID
	Title      string // 消息标题
	Content    string // 消息内容
}

// AccessTokenResponse 定义了微信AccessToken API的响应结构体
// 当调用微信API获取access_token时，微信服务器会返回此格式的JSON数据
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"` // 获取到的访问令牌，用于后续API调用
	ExpiresIn   int    `json:"expires_in"`   // 令牌过期时间，单位为秒
}

// TemplateDataItem 定义了模板消息数据的结构体
// 用于存储模板消息中的具体数据
type TemplateDataItem struct {
	Value string `json:"value"` // 消息文本
}

// https://developers.weixin.qq.com/doc/subscription/api/notify/subscribe/api_templatesubscribe.html
// TemplateMessageRequest 定义了发送微信模板消息的请求结构体
// 这是向微信服务器发送模板消息时所需的JSON数据格式
type TemplateMessageRequest struct {
	ToUser     string                      `json:"touser"`      // 接收消息用户的OpenID
	TemplateID string                      `json:"template_id"` // 模板消息的模板ID
	URL        string                      `json:"url"`         // 用户点击消息后跳转的URL
	Data       map[string]TemplateDataItem `json:"data"`        // 模板消息的具体数据，键值对形式
}

// WechatAPIResponse 定义了微信API的通用响应结构体
// 微信API调用成功或失败都会返回此格式的响应
type WechatAPIResponse struct {
	Errcode int    `json:"errcode"` // 错误码，0表示成功，其他值表示失败
	Errmsg  string `json:"errmsg"`  // 错误信息，当errcode不为0时提供详细错误描述
}

// TokenRequestParams 定义了获取微信AccessToken的请求参数结构体
// 这是调用微信stable_token API时需要的参数
type TokenRequestParams struct {
	AppID        string `json:"appid"`         // 微信公众号AppID
	Secret       string `json:"secret"`        // 微信公众号AppSecret
	GrantType    string `json:"grant_type"`    // 授权类型，固定为"client_credential"
	ForceRefresh bool   `json:"force_refresh"` // 是否强制刷新token，false表示使用缓存
}

func main() {
	fmt.Printf("wxpush 版本: %s\n", Version)

	// 定义命令行参数
	var appID string      // 微信公众AppID
	var secret string     // 微信公众AppSecret
	var userID string     // 接收消息的用户的UserID
	var templateID string // 微信模板消息的模板ID
	var title string      // 消息标题
	var content string    // 消息内容

	flag.StringVar(&appID, "appID", "", "微信公众AppID")
	flag.StringVar(&secret, "secret", "", "微信公众AppSecret")
	flag.StringVar(&userID, "userID", "", "接收消息的用户的UserID")
	flag.StringVar(&templateID, "templateID", "", "微信模板消息的模板ID")
	flag.StringVar(&title, "title", "", "消息标题")
	flag.StringVar(&content, "content", "", "消息内容")

	// 解析命令行参数
	flag.Parse()

	// 验证必填参数
	if appID == "" || secret == "" || userID == "" || templateID == "" || title == "" || content == "" {
		fmt.Println("Error: 缺少参数")
		fmt.Println("Usage:")
		fmt.Println("./wxpush -appID <AppID> -secret <Secret> -userID <UserID> \\")
		fmt.Println("         -templateID <TemplateID> -title <消息标题> -content <消息内容>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 构建请求参数
	params := RequestParams{
		AppID:      appID,
		Secret:     secret,
		UserID:     userID,
		TemplateID: templateID,
		Title:      title,
		Content:    content,
	}

	fmt.Println("开始发送微信消息...")
	fmt.Printf("用户: %s\n", userID)
	fmt.Printf("模板ID: %s\n", templateID)
	fmt.Printf("标题: %s\n", title)
	fmt.Printf("内容: %s\n", content)

	// 获取微信AccessToken
	fmt.Println("正在获取微信AccessToken...")
	token, err := getAccessToken(appID, secret)
	if err != nil {
		fmt.Printf("获取AccessToken失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("AccessToken获取成功")

	// 发送微信模板消息
	fmt.Println("正在发送模板消息...")
	resp, err := sendTemplateMessage(token, params)
	if err != nil {
		fmt.Printf("发送消息失败: %v\n", err)
		os.Exit(1)
	}

	// 检查发送结果
	if resp.Errcode == 0 {
		fmt.Println("消息发送成功！")
		fmt.Printf("响应码: %d\n", resp.Errcode)
		fmt.Printf("响应信息: %s\n", resp.Errmsg)
	} else {
		fmt.Println("消息发送失败")
		fmt.Printf("错误码: %d\n", resp.Errcode)
		fmt.Printf("错误信息: %s\n", resp.Errmsg)
		os.Exit(1)
	}
}

// getAccessToken 获取微信API的访问令牌
// 微信API调用需要使用access_token作为身份验证凭证
// 使用stable_token接口获取token，该接口支持token缓存
// 参数：
//   - appid: 微信公众号的AppID
//   - secret: 微信公众号的AppSecret
//
// 返回：
//   - string: 获取到的access_token
//   - error: 如果获取失败，返回错误信息
func getAccessToken(appID string, secret string) (string, error) {
	// 构建请求参数
	requestParams := TokenRequestParams{
		AppID:        appID,               // 微信公众号的AppID
		Secret:       secret,              // 微信公众号的AppSecret
		GrantType:    "client_credential", // 固定的授权类型
		ForceRefresh: false,               // 不强制刷新，使用微信缓存的token
	}

	// 将结构体转换为JSON格式
	// json.Marshal将Go结构体编码为JSON字节数组
	jsonData, err := json.Marshal(requestParams)
	if err != nil {
		return "", err
	}

	// 创建HTTP客户端
	client := &http.Client{}

	// 发送POST请求到微信API获取token
	// URL: https://api.weixin.qq.com/cgi-bin/stable_token
	// Content-Type: application/json
	// Body: JSON格式的请求参数
	resp, err := client.Post("https://api.weixin.qq.com/cgi-bin/stable_token", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}
	// 确保在函数返回前关闭响应体
	defer resp.Body.Close()

	// 读取响应体内容
	// io.ReadAll读取所有数据直到EOF
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析JSON响应
	// json.Unmarshal将JSON数据解析到AccessTokenResponse结构体
	var tokenResp AccessTokenResponse
	err = json.Unmarshal(body, &tokenResp)

	if err != nil {
		return "", err
	}

	// 返回获取到的access_token
	return tokenResp.AccessToken, nil
}

// sendTemplateMessage 发送微信模板消息
// 这是核心功能函数，负责构建模板消息请求并调用微信API发送
// 模板消息会推送到用户的微信，用户点击后可以跳转到详情页面
// 参数：
//   - accessToken: 微信API访问令牌
//   - params: 包含消息标题、内容、用户ID等信息的参数结构体
//
// 返回：
//   - WechatAPIResponse: 微信API的响应，包含错误码和错误信息
//   - error: 如果发送失败，返回错误信息
func sendTemplateMessage(accessToken string, params RequestParams) (WechatAPIResponse, error) {
	// 构建请求URL
	// 微信模板消息发送接口：https://api.weixin.qq.com/cgi-bin/message/template/send
	// 需要在URL中附带access_token参数
	apiUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken)

	// 构建请求数据
	// TemplateMessageRequest结构体定义了微信模板消息的格式
	requestData := TemplateMessageRequest{
		ToUser:     params.UserID,
		TemplateID: params.TemplateID,
		URL:        "",
		Data: map[string]TemplateDataItem{
			"title": {
				Value: params.Title,
			},
			"content": {
				Value: params.Content,
			},
		},
	}

	// 将结构体转换为JSON格式
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return WechatAPIResponse{}, err
	}

	fmt.Printf("发送的请求数据: %s\n", string(jsonData))

	// 创建HTTP客户端
	client := &http.Client{}

	// 发送POST请求到微信API发送模板消息
	resp, err := client.Post(apiUrl, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return WechatAPIResponse{}, err
	}
	// 确保在函数返回前关闭响应体
	defer resp.Body.Close()

	// 读取响应体内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WechatAPIResponse{}, err
	}

	fmt.Printf("微信API响应: %s\n", string(body))

	// 解析JSON响应
	// 微信API会返回errcode和errmsg
	// errcode为0表示成功，其他值表示失败
	var apiResp WechatAPIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return WechatAPIResponse{}, err
	}

	// 返回微信API的响应
	return apiResp, nil
}
