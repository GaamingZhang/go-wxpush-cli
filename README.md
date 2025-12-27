# go-wxpush-cli

一个简单的微信模板消息推送命令行工具，基于 Go 语言开发。用于推送本地Jenkins流水线信息到订阅微信公众号的特定用户。

## 功能特性

- 支持微信模板消息推送
- 使用微信 `stable_token` 接口获取 Access Token，支持 Token 缓存
- 简洁的命令行参数设计

## 环境要求

- Go 1.21 或更高版本

## 安装方法

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/gaamingzhang/go-wxpush-cli.git
cd go-wxpush-cli

# 编译程序
go build -o wxpush main.go

## 使用方法

### 基本语法

```bash
wxpush -appID <AppID> -secret <Secret> -userID <UserID> -templateID <TemplateID> -title <消息标题> -content <消息内容>
```

### 参数说明

| 参数 | 必填 | 说明 |
|------|------|------|
| `-appID` | 是 | 微信公众号 AppID |
| `-secret` | 是 | 微信公众号 AppSecret |
| `-userID` | 是 | 接收消息用户的 UserID |
| `-templateID` | 是 | 微信模板消息的模板 ID |
| `-title` | 是 | 消息标题 |
| `-content` | 是 | 消息内容 |

### 使用示例

```bash
# 发送一条模板消息
wxpush \
  -appID "wx1234567890abcdef" \
  -secret "your_secret_here" \
  -userID "o1234567890abcdefghijklmnopqrstuv" \
  -templateID "template_id_here" \
  -title "Jenkins 流水线通知" \
  -content "您的 Jenkins 流水线已完成，构建状态为完成。"
```

## 微信公众号配置

在使用本工具之前，您需要：

1. 注册并登录 [微信公众平台](https://mp.weixin.qq.com/)
2. 获取 AppID 和 AppSecret（开发 -> 基本配置）
3. 创建并配置模板消息（功能 -> 模板消息）
4. 获取用户的 OpenID（可通过网页授权获取）

## 项目结构

```
go-wxpush-cli/
├── .gitignore      # Git 忽略文件配置
├── go.mod          # Go 模块定义
├── go.sum          # Go 模块依赖校验和
├── main.go         # 主程序入口
└── README.md       # 项目说明文档
```

## 代码结构说明

### 主要数据结构

- `RequestParams`: 存储命令行参数
- `AccessTokenResponse`: 微信 Access Token API 响应结构
- `TemplateMessageRequest`: 微信模板消息请求结构
- `WechatAPIResponse`: 微信 API 通用响应结构
- `TokenRequestParams`: 获取 Access Token 的请求参数

### 核心函数

- `getAccessToken(appID, secret)`: 获取微信 API 访问令牌
- `sendTemplateMessage(accessToken, params)`: 发送微信模板消息