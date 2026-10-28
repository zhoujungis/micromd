package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

//go:embed web/index.html
var indexHTML []byte

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	shell32 = windows.NewLazySystemDLL("shell32.dll")
)

type initialDoc struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func main() {
	enableDPI()

	base := appDataDir()
	appDir := filepath.Join(base, "app")
	page := filepath.Join(appDir, "index.html")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return
	}
	if current, err := os.ReadFile(page); err != nil || string(current) != string(indexHTML) {
		if err := os.WriteFile(page, indexHTML, 0o644); err != nil {
			return
		}
	}

	docPath := ""
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		if abs, err := filepath.Abs(os.Args[1]); err == nil {
			if info, err := os.Stat(abs); err == nil && !info.IsDir() {
				docPath = abs
			}
		}
	}

	width, height := windowSize()
	view := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		DataPath:  filepath.Join(base, "webview"),
		WindowOptions: webview2.WindowOptions{
			Title:  "MicroMD",
			Width:  width,
			Height: height,
			IconId: 1,
			Center: true,
		},
	})
	if view == nil {
		openExternal(page)
		return
	}
	defer view.Destroy()

	_ = view.Bind("micromdLoad", func() (string, error) {
		if docPath == "" {
			return "", nil
		}
		content, err := os.ReadFile(docPath)
		if err != nil {
			return "", err
		}
		payload, err := json.Marshal(initialDoc{Name: filepath.Base(docPath), Content: string(content)})
		if err != nil {
			return "", err
		}
		return string(payload), nil
	})

	_ = view.Bind("micromdSave", func(content string) (string, error) {
		if docPath == "" {
			return "", errors.New("no file")
		}
		if err := os.WriteFile(docPath, []byte(content), 0o644); err != nil {
			return "", err
		}
		return "ok", nil
	})

	_ = view.Bind("micromdOpenExternal", func(target string) (string, error) {
		if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "mailto:") {
			return "", nil
		}
		return "ok", openExternal(target)
	})

	view.Navigate(fileURL(page))
	view.Run()
}

func windowSize() (uint, uint) {
	scale := dpiScale()
	width := 1280 * scale
	height := 860 * scale
	metrics := user32.NewProc("GetSystemMetrics")
	if sw, _, _ := metrics.Call(0); sw > 0 && width > float64(sw)*0.92 {
		width = float64(sw) * 0.92
	}
	if sh, _, _ := metrics.Call(1); sh > 0 && height > float64(sh)*0.88 {
		height = float64(sh) * 0.88
	}
	return uint(math.Round(width)), uint(math.Round(height))
}

func dpiScale() float64 {
	if proc := user32.NewProc("GetDpiForSystem"); proc.Find() == nil {
		if dpi, _, _ := proc.Call(); dpi > 0 {
			return float64(dpi) / 96
		}
	}
	return 1
}

func appDataDir() string {
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		return filepath.Join(dir, "MicroMD")
	}
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "MicroMD")
	}
	return "MicroMD"
}

func fileURL(path string) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p}).String()
}

func openExternal(target string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	ret, _, _ := shell32.NewProc("ShellExecuteW").Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		0,
		0,
		uintptr(1),
	)
	if ret <= 32 {
		return errors.New("ShellExecute failed")
	}
	return nil
}

func enableDPI() {
	if proc := user32.NewProc("SetProcessDpiAwarenessContext"); proc.Find() == nil {
		proc.Call(^uintptr(3))
		return
	}
	if shcore := windows.NewLazySystemDLL("shcore.dll"); shcore.Load() == nil {
		if proc := shcore.NewProc("SetProcessDpiAwareness"); proc.Find() == nil {
			proc.Call(2)
		}
	}
}
