# WinProxy

WinProxy is a small Windows Remote Desktop SOCKS5 forwarding tool written in Go.

## Features

- Local TCP listener for Remote Desktop clients
- SOCKS5 CONNECT forwarding to a target RDP host
- Optional SOCKS5 username/password authentication
- Chinese and English Windows UI
- 32-bit Windows release build for broad compatibility
- Simple bilingual website pages with PHP language switching

## Build

```powershell
go test ./...
go build -buildvcs=false -ldflags "-H windowsgui" -o winproxy.exe .
```

Use `-buildvcs=false` because this recovered workspace may contain an incomplete `.git` directory.

For a one-click release build, run:

```powershell
.\build-release.cmd
```

It builds the 32-bit `winproxy.exe`, which runs on both 32-bit and 64-bit Windows.

The release script also generates the Windows icon and version resource files when needed.

## Configuration

The app reads and writes `winproxy.json` in the working directory.

- `ListenHost` / `ListenPort`: local address used by Remote Desktop, for example `127.0.0.1:757`
- `TargetHost` / `TargetPort`: target Remote Desktop server, usually port `3389`
- `SocksHost` / `SocksPort`: SOCKS5 proxy server
- `Username` / `Password`: optional SOCKS5 authentication
- `Language`: `zh-CN` or `en-US`

## Website

`index.php` serves `zh.html` or `en.html` based on the client's `Accept-Language` header.

## License

MIT
