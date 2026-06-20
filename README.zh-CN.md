# WinProxy

[English README](README.md)

WinProxy 是一个使用 Go 编写的 Windows 远程桌面 SOCKS5 转发工具。

## 功能

- 为远程桌面客户端提供本地 TCP 监听入口
- 通过 SOCKS5 CONNECT 转发到目标 RDP 主机
- 支持可选的 SOCKS5 用户名和密码认证
- Windows 界面支持中文和英文
- 发布版使用 32 位构建，兼容 32 位和 64 位 Windows
- 提供中英文网站页面，并通过 PHP 根据客户端语言自动切换

## 编译

```powershell
go test ./...
go build -buildvcs=false -ldflags "-H windowsgui" -o winproxy.exe .
```

由于这个重建工作区可能包含不完整的 `.git` 目录，构建时使用 `-buildvcs=false` 可以避免 Go 的 VCS 信息探测失败。

一键发布构建：

```powershell
.\build-release.cmd
```

脚本会生成 32 位 `winproxy.exe`，可运行在 32 位和 64 位 Windows 上。

发布脚本还会在需要时生成 Windows 图标和版本资源文件。

## 配置

程序会读取并写入当前目录下的 `winproxy.json`。

- `ListenHost` / `ListenPort`：远程桌面客户端连接的本地地址，例如 `127.0.0.1:757`
- `TargetHost` / `TargetPort`：目标远程桌面服务器，通常端口为 `3389`
- `SocksHost` / `SocksPort`：SOCKS5 代理服务器
- `Username` / `Password`：可选的 SOCKS5 认证信息
- `Language`：`zh-CN` 或 `en-US`

## 网站

`index.php` 会根据客户端的 `Accept-Language` 请求头输出 `zh.html` 或 `en.html`。

## 许可证

MIT
