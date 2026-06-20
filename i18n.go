package main

type texts struct {
	Title            string
	LocalGroup       string
	ListenAddress    string
	TargetGroup      string
	TargetAddress    string
	SocksGroup       string
	SocksAddress     string
	Username         string
	Password         string
	Status           string
	Start            string
	Stop             string
	Close            string
	Copyright        string
	Website          string
	Started          string
	Stopped          string
	StartFailed      string
	StopFailed       string
	ConfigSaved      string
	Error            string
	Info             string
	NeedListen       string
	NeedTarget       string
	NeedSocks        string
	ListenFailed     string
	SocksFailed      string
	TargetFailed     string
	SocksRejected    string
	AuthRejected     string
	UnsupportedAuth  string
	LanguageTip      string
	OpenWebsiteError string
}

var translations = map[string]texts{
	"zh-CN": {
		Title:            "WinProxy - 远程桌面 SOCKS5 代理",
		LocalGroup:       "本地监听",
		ListenAddress:    "远程桌面连接到此地址:",
		TargetGroup:      "远程桌面目标",
		TargetAddress:    "目标地址:",
		SocksGroup:       "SOCKS5 代理",
		SocksAddress:     "代理地址:",
		Username:         "用户名:",
		Password:         "密码:",
		Status:           "状态:",
		Start:            "启动",
		Stop:             "停止",
		Close:            "关闭",
		Copyright:        "Copyright (c) 2026 WinProxy. 官网:",
		Website:          "winproxy.org",
		Started:          "已启动",
		Stopped:          "未启动",
		StartFailed:      "启动失败",
		StopFailed:       "停止失败",
		ConfigSaved:      "配置已保存",
		Error:            "错误",
		Info:             "提示",
		NeedListen:       "请填写本地监听地址和端口。",
		NeedTarget:       "请填写远程桌面目标地址和端口。",
		NeedSocks:        "请填写 SOCKS5 代理地址和端口。",
		ListenFailed:     "监听失败",
		SocksFailed:      "连接 SOCKS5 代理失败",
		TargetFailed:     "连接远程桌面目标失败",
		SocksRejected:    "SOCKS5 代理拒绝连接",
		AuthRejected:     "SOCKS5 用户名或密码认证失败",
		UnsupportedAuth:  "SOCKS5 代理不支持当前认证方式",
		LanguageTip:      "语言已切换，界面已更新。",
		OpenWebsiteError: "无法打开网站。",
	},
	"en-US": {
		Title:            "WinProxy - Remote Desktop SOCKS5 Proxy",
		LocalGroup:       "Local Listener",
		ListenAddress:    "Remote Desktop connects to:",
		TargetGroup:      "Remote Desktop Target",
		TargetAddress:    "Target address:",
		SocksGroup:       "SOCKS5 Proxy",
		SocksAddress:     "Proxy address:",
		Username:         "Username:",
		Password:         "Password:",
		Status:           "Status:",
		Start:            "Start",
		Stop:             "Stop",
		Close:            "Close",
		Copyright:        "Copyright (c) 2026 WinProxy. Website:",
		Website:          "winproxy.org",
		Started:          "Running",
		Stopped:          "Stopped",
		StartFailed:      "Start failed",
		StopFailed:       "Stop failed",
		ConfigSaved:      "Configuration saved",
		Error:            "Error",
		Info:             "Info",
		NeedListen:       "Please enter the local listen address and port.",
		NeedTarget:       "Please enter the Remote Desktop target address and port.",
		NeedSocks:        "Please enter the SOCKS5 proxy address and port.",
		ListenFailed:     "Failed to listen",
		SocksFailed:      "Failed to connect to the SOCKS5 proxy",
		TargetFailed:     "Failed to connect to the Remote Desktop target",
		SocksRejected:    "The SOCKS5 proxy rejected the connection",
		AuthRejected:     "SOCKS5 username or password authentication failed",
		UnsupportedAuth:  "The SOCKS5 proxy does not support this authentication method",
		LanguageTip:      "Language changed. The interface has been updated.",
		OpenWebsiteError: "Unable to open the website.",
	},
}

func langName(code string) string {
	switch code {
	case "en-US":
		return "English"
	default:
		return "中文"
	}
}

func normalizeLang(code string) string {
	if code == "en-US" {
		return code
	}
	return "zh-CN"
}

func tr(code string) texts {
	code = normalizeLang(code)
	return translations[code]
}
