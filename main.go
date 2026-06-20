package main

import (
	"sync"
	"syscall"
	"unsafe"
)

const (
	idListenHost = 1001
	idListenPort = 1002
	idTargetHost = 1003
	idTargetPort = 1004
	idSocksHost  = 1005
	idSocksPort  = 1006
	idUsername   = 1007
	idPassword   = 1008
	idStart      = 1101
	idStop       = 1102
	idClose      = 1103
	idLanguage   = 1104
	idWebsite    = 1105
	appIconID    = 1
)

type appWindow struct {
	hwnd    HWND
	cfg     Config
	srv     ProxyServer
	running bool

	font     HFONT
	linkFont HFONT
	iconBig  HICON
	iconSm   HICON

	groupLocal  HWND
	groupTarget HWND
	groupSocks  HWND
	labelListen HWND
	labelTarget HWND
	labelSocks  HWND
	labelUser   HWND
	labelPass   HWND
	labelStatus HWND
	statusText  HWND
	copyRight   HWND
	website     HWND

	editListenHost HWND
	editListenPort HWND
	editTargetHost HWND
	editTargetPort HWND
	editSocksHost  HWND
	editSocksPort  HWND
	editUsername   HWND
	editPassword   HWND
	comboLanguage  HWND
	btnStart       HWND
	btnStop        HWND
	btnClose       HWND
}

var (
	app      *appWindow
	langMu   sync.RWMutex
	activeLn = "zh-CN"
)

func currentLanguage() string {
	langMu.RLock()
	defer langMu.RUnlock()
	return activeLn
}

func setCurrentLanguage(code string) {
	langMu.Lock()
	activeLn = normalizeLang(code)
	langMu.Unlock()
}

func main() {
	cfg := loadConfig()
	setCurrentLanguage(cfg.Language)
	app = &appWindow{cfg: cfg}
	app.srv.onError = func(err error) {
		messageBox(app.hwnd, tr(currentLanguage()).Error, err.Error(), MB_OK|MB_ICONERROR)
	}
	runGUI()
}

func runGUI() {
	instance := getModuleHandle()
	className := "WinProxyWindow"
	wndProc := syscall.NewCallback(windowProc)
	iconBig := loadIconSize(instance, appIconID, 32, 32)
	iconSm := loadIconSize(instance, appIconID, 16, 16)
	if iconBig == 0 {
		iconBig = loadIcon(instance, appIconID)
	}
	if iconSm == 0 {
		iconSm = iconBig
	}
	app.iconBig = iconBig
	app.iconSm = iconSm
	cls := WNDCLASSEX{
		Size:       uint32(unsafe.Sizeof(WNDCLASSEX{})),
		WndProc:    wndProc,
		Instance:   instance,
		Icon:       iconBig,
		IconSm:     iconSm,
		Cursor:     loadCursor(IDC_ARROW),
		Background: HBRUSH(COLOR_BTNFACE + 1),
		ClassName:  utf16Ptr(className),
	}
	if registerClass(&cls) == 0 {
		return
	}

	t := tr(currentLanguage())
	hwnd := createWindowEx(0, className, t.Title, WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_MINIMIZEBOX, CW_USEDEFAULT, CW_USEDEFAULT, 570, 480, 0, 0, instance, 0)
	if hwnd == 0 {
		return
	}
	app.hwnd = hwnd
	sendMessage(hwnd, WM_SETICON, ICON_BIG, uintptr(iconBig))
	sendMessage(hwnd, WM_SETICON, ICON_SMALL, uintptr(iconSm))
	procShowWindow.Call(uintptr(hwnd), SW_SHOW)
	procUpdateWindow.Call(uintptr(hwnd))

	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func windowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		app.hwnd = hwnd
		app.createControls()
		app.applyTexts()
		app.loadValues()
		app.setRunning(false)
		return 0
	case WM_COMMAND:
		app.handleCommand(loword(wParam), hiword(wParam))
		return 0
	case WM_CLOSE:
		app.close()
		return 0
	case WM_DESTROY:
		if app.font != 0 {
			deleteObject(HANDLE(app.font))
		}
		if app.linkFont != 0 {
			deleteObject(HANDLE(app.linkFont))
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	return defWindowProc(hwnd, msg, wParam, lParam)
}

func (a *appWindow) createControls() {
	a.font = createFont("Microsoft YaHei UI", -14, 400, false)
	a.linkFont = createFont("Microsoft YaHei UI", -14, 400, true)

	a.groupLocal = a.groupBox("", 8, 18, 540, 84)
	a.labelListen = a.label("", 22, 48, 170, 24)
	a.editListenHost = a.edit(192, 46, 240, 28, false)
	a.label(":", 438, 48, 12, 24)
	a.editListenPort = a.edit(455, 46, 64, 28, false)

	a.groupTarget = a.groupBox("", 8, 120, 540, 74)
	a.labelTarget = a.label("", 22, 150, 100, 24)
	a.editTargetHost = a.edit(128, 148, 304, 28, false)
	a.label(":", 438, 150, 12, 24)
	a.editTargetPort = a.edit(455, 148, 64, 28, false)

	a.groupSocks = a.groupBox("", 8, 216, 540, 104)
	a.labelSocks = a.label("", 22, 247, 100, 24)
	a.editSocksHost = a.edit(128, 245, 304, 28, false)
	a.label(":", 438, 247, 12, 24)
	a.editSocksPort = a.edit(455, 245, 64, 28, false)
	a.labelUser = a.label("", 22, 286, 70, 24)
	a.editUsername = a.edit(100, 284, 140, 28, false)
	a.labelPass = a.label("", 274, 286, 70, 24)
	a.editPassword = a.edit(344, 284, 175, 28, true)

	a.btnStart = a.button("", idStart, 8, 334, 88, 32)
	a.btnStop = a.button("", idStop, 104, 334, 88, 32)
	a.btnClose = a.button("", idClose, 200, 334, 88, 32)

	a.comboLanguage = a.combo(idLanguage, 326, 334, 110, 220)
	sendMessage(a.comboLanguage, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr("中文"))))
	sendMessage(a.comboLanguage, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr("English"))))

	a.labelStatus = a.label("", 8, 374, 54, 24)
	a.statusText = a.label("", 62, 374, 90, 24)
	a.copyRight = a.label("", 8, 406, 248, 24)
	a.website = a.linkLabel("", idWebsite, 242, 406, 130, 24)
	sendMessage(a.website, WM_SETFONT, uintptr(a.linkFont), 1)
}

func (a *appWindow) groupBox(text string, x, y, w, h int32) HWND {
	hwnd := createWindowEx(0, "BUTTON", text, WS_CHILD|WS_VISIBLE|BS_GROUPBOX, x, y, w, h, a.hwnd, 0, getModuleHandle(), 0)
	a.setFont(hwnd)
	return hwnd
}

func (a *appWindow) label(text string, x, y, w, h int32) HWND {
	hwnd := createWindowEx(0, "STATIC", text, WS_CHILD|WS_VISIBLE, x, y, w, h, a.hwnd, 0, getModuleHandle(), 0)
	a.setFont(hwnd)
	return hwnd
}

func (a *appWindow) linkLabel(text string, id uintptr, x, y, w, h int32) HWND {
	hwnd := createWindowEx(0, "STATIC", text, WS_CHILD|WS_VISIBLE|SS_NOTIFY, x, y, w, h, a.hwnd, HMENU(id), getModuleHandle(), 0)
	a.setFont(hwnd)
	return hwnd
}

func (a *appWindow) edit(x, y, w, h int32, password bool) HWND {
	style := uint32(WS_CHILD | WS_VISIBLE | WS_BORDER | WS_TABSTOP | ES_AUTOHSCROLL)
	if password {
		style |= ES_PASSWORD
	}
	hwnd := createWindowEx(0, "EDIT", "", style, x, y, w, h, a.hwnd, 0, getModuleHandle(), 0)
	a.setFont(hwnd)
	return hwnd
}

func (a *appWindow) button(text string, id uintptr, x, y, w, h int32) HWND {
	hwnd := createWindowEx(0, "BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, x, y, w, h, a.hwnd, HMENU(id), getModuleHandle(), 0)
	a.setFont(hwnd)
	return hwnd
}

func (a *appWindow) combo(id uintptr, x, y, w, h int32) HWND {
	hwnd := createWindowEx(0, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST|CBS_HASSTRINGS, x, y, w, h, a.hwnd, HMENU(id), getModuleHandle(), 0)
	a.setFont(hwnd)
	return hwnd
}

func (a *appWindow) setFont(hwnd HWND) {
	if a.font != 0 {
		sendMessage(hwnd, WM_SETFONT, uintptr(a.font), 1)
	}
}

func (a *appWindow) applyTexts() {
	t := tr(currentLanguage())
	setWindowText(a.hwnd, t.Title)
	setWindowText(a.groupLocal, t.LocalGroup)
	setWindowText(a.groupTarget, t.TargetGroup)
	setWindowText(a.groupSocks, t.SocksGroup)
	setWindowText(a.labelListen, t.ListenAddress)
	setWindowText(a.labelTarget, t.TargetAddress)
	setWindowText(a.labelSocks, t.SocksAddress)
	setWindowText(a.labelUser, t.Username)
	setWindowText(a.labelPass, t.Password)
	setWindowText(a.labelStatus, t.Status)
	setWindowText(a.btnStart, t.Start)
	setWindowText(a.btnStop, t.Stop)
	setWindowText(a.btnClose, t.Close)
	setWindowText(a.copyRight, t.Copyright)
	setWindowText(a.website, t.Website)
	a.applyStatusText()
}

func (a *appWindow) loadValues() {
	cfg := a.cfg
	setWindowText(a.editListenHost, cfg.ListenHost)
	setWindowText(a.editListenPort, cfg.ListenPort)
	setWindowText(a.editTargetHost, cfg.TargetHost)
	setWindowText(a.editTargetPort, cfg.TargetPort)
	setWindowText(a.editSocksHost, cfg.SocksHost)
	setWindowText(a.editSocksPort, cfg.SocksPort)
	setWindowText(a.editUsername, cfg.Username)
	setWindowText(a.editPassword, cfg.Password)
	if normalizeLang(cfg.Language) == "en-US" {
		sendMessage(a.comboLanguage, CB_SETCURSEL, 1, 0)
	} else {
		sendMessage(a.comboLanguage, CB_SETCURSEL, 0, 0)
	}
}

func (a *appWindow) readConfig() Config {
	lang := "zh-CN"
	if sendMessage(a.comboLanguage, CB_GETCURSEL, 0, 0) == 1 {
		lang = "en-US"
	}
	return Config{
		ListenHost: getWindowText(a.editListenHost),
		ListenPort: getWindowText(a.editListenPort),
		TargetHost: getWindowText(a.editTargetHost),
		TargetPort: getWindowText(a.editTargetPort),
		SocksHost:  getWindowText(a.editSocksHost),
		SocksPort:  getWindowText(a.editSocksPort),
		Username:   getWindowText(a.editUsername),
		Password:   getWindowText(a.editPassword),
		Language:   lang,
	}
}

func (a *appWindow) saveValues() {
	a.cfg = a.readConfig()
	_ = saveConfig(a.cfg)
}

func (a *appWindow) handleCommand(id uint16, code uint16) {
	switch id {
	case idStart:
		if code == BN_CLICKED {
			a.startProxy()
		}
	case idStop:
		if code == BN_CLICKED {
			a.stopProxy()
		}
	case idClose:
		if code == BN_CLICKED {
			a.close()
		}
	case idLanguage:
		if code == CBN_SELCHANGE {
			a.changeLanguage()
		}
	case idWebsite:
		if code == BN_CLICKED {
			if shellExecute(a.hwnd, "open", "https://winproxy.org") <= 32 {
				messageBox(a.hwnd, tr(currentLanguage()).Error, tr(currentLanguage()).OpenWebsiteError, MB_OK|MB_ICONERROR)
			}
		}
	}
}

func (a *appWindow) close() {
	_ = a.srv.Stop()
	a.saveValues()
	procDestroyWindow.Call(uintptr(a.hwnd))
}

func (a *appWindow) startProxy() {
	cfg := a.readConfig()
	setCurrentLanguage(cfg.Language)
	if err := saveConfig(cfg); err != nil {
		messageBox(a.hwnd, tr(currentLanguage()).Error, err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	err := a.srv.Start(ProxyConfig{
		ListenHost: cfg.ListenHost,
		ListenPort: cfg.ListenPort,
		TargetHost: cfg.TargetHost,
		TargetPort: cfg.TargetPort,
		SocksHost:  cfg.SocksHost,
		SocksPort:  cfg.SocksPort,
		Username:   cfg.Username,
		Password:   cfg.Password,
	})
	if err != nil {
		messageBox(a.hwnd, tr(currentLanguage()).StartFailed, err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	a.cfg = cfg
	a.setRunning(true)
}

func (a *appWindow) stopProxy() {
	if err := a.srv.Stop(); err != nil {
		messageBox(a.hwnd, tr(currentLanguage()).StopFailed, err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	a.setRunning(false)
}

func (a *appWindow) changeLanguage() {
	a.cfg = a.readConfig()
	setCurrentLanguage(a.cfg.Language)
	_ = saveConfig(a.cfg)
	a.applyTexts()
}

func (a *appWindow) setRunning(running bool) {
	a.running = running
	enableWindow(a.btnStart, !running)
	enableWindow(a.btnStop, running)
	a.applyStatusText()
}

func (a *appWindow) applyStatusText() {
	if a.running {
		setWindowText(a.statusText, tr(currentLanguage()).Started)
	} else {
		setWindowText(a.statusText, tr(currentLanguage()).Stopped)
	}
}
