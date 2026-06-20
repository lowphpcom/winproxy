package main

import (
	"syscall"
	"unsafe"
)

type (
	HANDLE    uintptr
	HWND      HANDLE
	HINSTANCE HANDLE
	HICON     HANDLE
	HCURSOR   HANDLE
	HBRUSH    HANDLE
	HMENU     HANDLE
	HFONT     HANDLE
)

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   HINSTANCE
	Icon       HICON
	Cursor     HCURSOR
	Background HBRUSH
	MenuName   *uint16
	ClassName  *uint16
	IconSm     HICON
}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW  = user32.NewProc("RegisterClassExW")
	procCreateWindowExW   = user32.NewProc("CreateWindowExW")
	procDefWindowProcW    = user32.NewProc("DefWindowProcW")
	procDestroyWindow     = user32.NewProc("DestroyWindow")
	procDispatchMessageW  = user32.NewProc("DispatchMessageW")
	procGetMessageW       = user32.NewProc("GetMessageW")
	procLoadCursorW       = user32.NewProc("LoadCursorW")
	procMessageBoxW       = user32.NewProc("MessageBoxW")
	procPostQuitMessage   = user32.NewProc("PostQuitMessage")
	procSendMessageW      = user32.NewProc("SendMessageW")
	procSetWindowTextW    = user32.NewProc("SetWindowTextW")
	procShowWindow        = user32.NewProc("ShowWindow")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procUpdateWindow      = user32.NewProc("UpdateWindow")
	procGetWindowTextW    = user32.NewProc("GetWindowTextW")
	procGetWindowTextLenW = user32.NewProc("GetWindowTextLengthW")
	procEnableWindow      = user32.NewProc("EnableWindow")
	procLoadIconW         = user32.NewProc("LoadIconW")
	procLoadImageW        = user32.NewProc("LoadImageW")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")

	procCreateFontW  = gdi32.NewProc("CreateFontW")
	procDeleteObject = gdi32.NewProc("DeleteObject")

	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

const (
	WS_OVERLAPPED       = 0x00000000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_MINIMIZEBOX      = 0x00020000
	WS_CHILD            = 0x40000000
	WS_VISIBLE          = 0x10000000
	WS_TABSTOP          = 0x00010000
	WS_GROUP            = 0x00020000
	WS_BORDER           = 0x00800000
	WS_OVERLAPPEDWINDOW = 0x00CF0000

	ES_AUTOHSCROLL = 0x0080
	ES_PASSWORD    = 0x0020

	BS_PUSHBUTTON = 0x00000000
	BS_GROUPBOX   = 0x00000007

	SS_NOTIFY = 0x00000100

	CBS_DROPDOWNLIST = 0x0003
	CBS_HASSTRINGS   = 0x0200

	CW_USEDEFAULT = -2147483648
	SW_SHOW       = 5

	WM_CREATE   = 0x0001
	WM_DESTROY  = 0x0002
	WM_COMMAND  = 0x0111
	WM_CLOSE    = 0x0010
	WM_SETFONT  = 0x0030
	WM_SETTEXT  = 0x000C
	WM_SETICON  = 0x0080
	BM_SETSTATE = 0x00F3

	CB_ADDSTRING       = 0x0143
	CB_SETCURSEL       = 0x014E
	CB_GETCURSEL       = 0x0147
	CBN_SELCHANGE      = 1
	BN_CLICKED         = 0
	COLOR_BTNFACE      = 15
	MB_OK              = 0x00000000
	MB_ICONERROR       = 0x00000010
	MB_ICONINFORMATION = 0x00000040
	IDC_ARROW          = 32512
	IMAGE_ICON         = 1
	LR_DEFAULTCOLOR    = 0x00000000
	ICON_SMALL         = 0
	ICON_BIG           = 1
)

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func loword(v uintptr) uint16 {
	return uint16(v & 0xffff)
}

func hiword(v uintptr) uint16 {
	return uint16((v >> 16) & 0xffff)
}

func getModuleHandle() HINSTANCE {
	r, _, _ := procGetModuleHandleW.Call(0)
	return HINSTANCE(r)
}

func loadCursor(id uintptr) HCURSOR {
	r, _, _ := procLoadCursorW.Call(0, id)
	return HCURSOR(r)
}

func loadIcon(instance HINSTANCE, id uintptr) HICON {
	r, _, _ := procLoadIconW.Call(uintptr(instance), id)
	return HICON(r)
}

func loadIconSize(instance HINSTANCE, id uintptr, width, height int32) HICON {
	r, _, _ := procLoadImageW.Call(
		uintptr(instance),
		id,
		IMAGE_ICON,
		uintptr(width),
		uintptr(height),
		LR_DEFAULTCOLOR,
	)
	return HICON(r)
}

func registerClass(cls *WNDCLASSEX) uintptr {
	r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(cls)))
	return r
}

func createWindowEx(exStyle uint32, className, windowName string, style uint32, x, y, w, h int32, parent HWND, menu HMENU, instance HINSTANCE, param uintptr) HWND {
	r, _, _ := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(windowName))),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		uintptr(parent), uintptr(menu), uintptr(instance), param,
	)
	return HWND(r)
}

func defWindowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func setWindowText(hwnd HWND, text string) {
	procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func sendMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func enableWindow(hwnd HWND, enabled bool) {
	v := uintptr(0)
	if enabled {
		v = 1
	}
	procEnableWindow.Call(uintptr(hwnd), v)
}

func getWindowText(hwnd HWND) string {
	n, _, _ := procGetWindowTextLenW.Call(uintptr(hwnd))
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

func messageBox(hwnd HWND, title, text string, flags uintptr) {
	procMessageBoxW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		flags,
	)
}

func createFont(name string, height int32, weight int32, underline bool) HFONT {
	u := uintptr(0)
	if underline {
		u = 1
	}
	r, _, _ := procCreateFontW.Call(
		uintptr(height), 0, 0, 0, uintptr(weight), 0, u, 0,
		1, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(utf16Ptr(name))),
	)
	return HFONT(r)
}

func deleteObject(h HANDLE) {
	if h != 0 {
		procDeleteObject.Call(uintptr(h))
	}
}

func shellExecute(hwnd HWND, verb, file string) uintptr {
	r, _, _ := procShellExecuteW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(utf16Ptr(verb))),
		uintptr(unsafe.Pointer(utf16Ptr(file))),
		0,
		0,
		SW_SHOW,
	)
	return r
}
