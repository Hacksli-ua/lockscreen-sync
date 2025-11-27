package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	user32                   = windows.NewLazySystemDLL("user32.dll")
	procSystemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

const (
	SPI_GETDESKWALLPAPER = 0x0073
	MAX_PATH             = 260
)

// Іконка для трею (16x16 ICO в байтах - проста синя іконка)
var iconData = []byte{
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x00, 0x00, 0x01, 0x00, 0x20, 0x00, 0x68, 0x04,
	0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

type App struct {
	lastWallpaper string
	stopChan      chan struct{}
	syncEnabled   bool
}

func NewApp() *App {
	return &App{
		stopChan:    make(chan struct{}),
		syncEnabled: true,
	}
}

// GetCurrentWallpaper отримує шлях до поточних шпалер робочого столу
func (a *App) GetCurrentWallpaper() (string, error) {
	// Спочатку спробуємо через реєстр
	key, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\Desktop`, registry.QUERY_VALUE)
	if err == nil {
		defer key.Close()
		wallpaper, _, err := key.GetStringValue("Wallpaper")
		if err == nil && wallpaper != "" && fileExists(wallpaper) {
			return wallpaper, nil
		}
	}

	// Якщо не вдалося, спробуємо через Windows API
	buffer := make([]uint16, MAX_PATH)
	ret, _, err := procSystemParametersInfo.Call(
		SPI_GETDESKWALLPAPER,
		MAX_PATH,
		uintptr(unsafe.Pointer(&buffer[0])),
		0,
	)
	if ret == 0 {
		return "", fmt.Errorf("SystemParametersInfo failed: %v", err)
	}

	wallpaperPath := windows.UTF16ToString(buffer)
	if wallpaperPath == "" {
		// Спробуємо TranscodedWallpaper
		appData := os.Getenv("APPDATA")
		transcodedPath := filepath.Join(appData, "Microsoft", "Windows", "Themes", "TranscodedWallpaper")
		if fileExists(transcodedPath) {
			return transcodedPath, nil
		}
		return "", fmt.Errorf("wallpaper path is empty")
	}

	return wallpaperPath, nil
}

// SetLockScreenWallpaper встановлює шпалери для екрану блокування
func (a *App) SetLockScreenWallpaper(imagePath string) error {
	if !fileExists(imagePath) {
		return fmt.Errorf("image file not found: %s", imagePath)
	}

	// Копіюємо зображення в публічну папку
	publicPictures := filepath.Join(os.Getenv("PUBLIC"), "Pictures")
	if err := os.MkdirAll(publicPictures, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	destPath := filepath.Join(publicPictures, "LockScreenWallpaper.jpg")

	// Копіюємо файл
	if err := copyFile(imagePath, destPath); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// 1. Встановлюємо через Policies\Personalization
	policyKey, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Policies\Microsoft\Windows\Personalization`,
		registry.ALL_ACCESS,
	)
	if err != nil {
		return fmt.Errorf("failed to open Personalization policy key (потрібні права адміністратора): %v", err)
	}
	defer policyKey.Close()

	// LockScreenImage
	if err := policyKey.SetStringValue("LockScreenImage", destPath); err != nil {
		return fmt.Errorf("failed to set LockScreenImage: %v", err)
	}

	// NoChangingLockScreen - заборонити зміну користувачами (опціонально)
	// policyKey.SetDWordValue("NoChangingLockScreen", 1)

	// 2. Встановлюємо через PersonalizationCSP (для Windows 10/11)
	cspKey, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\PersonalizationCSP`,
		registry.ALL_ACCESS,
	)
	if err == nil {
		defer cspKey.Close()
		cspKey.SetDWordValue("LockScreenImageStatus", 1)
		cspKey.SetStringValue("LockScreenImagePath", destPath)
		cspKey.SetStringValue("LockScreenImageUrl", destPath)
	}

	return nil
}

// SyncWallpaper синхронізує шпалери робочого столу з екраном блокування
func (a *App) SyncWallpaper() error {
	wallpaper, err := a.GetCurrentWallpaper()
	if err != nil {
		return fmt.Errorf("failed to get wallpaper: %v", err)
	}

	if err := a.SetLockScreenWallpaper(wallpaper); err != nil {
		return fmt.Errorf("failed to set lock screen: %v", err)
	}

	a.lastWallpaper = wallpaper
	return nil
}

// StartWatching починає моніторинг змін шпалер
func (a *App) StartWatching() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			if !a.syncEnabled {
				continue
			}

			wallpaper, err := a.GetCurrentWallpaper()
			if err != nil {
				continue
			}

			if wallpaper != a.lastWallpaper {
				if err := a.SetLockScreenWallpaper(wallpaper); err == nil {
					a.lastWallpaper = wallpaper
				}
			}
		}
	}
}

// StopWatching зупиняє моніторинг
func (a *App) StopWatching() {
	close(a.stopChan)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

var app *App

func main() {
	app = NewApp()

	// Початкова синхронізація
	app.SyncWallpaper()

	// Запускаємо системний трей
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("LockScreen Sync")
	systray.SetTooltip("Синхронізація шпалер з екраном блокування")

	// Створюємо тимчасовий файл іконки
	iconPath := filepath.Join(os.TempDir(), "lockscreen_icon.ico")
	if err := os.WriteFile(iconPath, generateIcon(), 0644); err == nil {
		if iconBytes, err := os.ReadFile(iconPath); err == nil {
			systray.SetIcon(iconBytes)
		}
	}

	mSync := systray.AddMenuItem("🔄 Синхронізувати зараз", "Синхронізувати шпалери")
	systray.AddSeparator()
	mToggle := systray.AddMenuItem("✅ Автосинхронізація увімкнена", "Увімкнути/вимкнути автосинхронізацію")
	systray.AddSeparator()
	mStatus := systray.AddMenuItem("", "Поточний статус")
	mStatus.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("❌ Вийти", "Закрити програму")

	// Оновлюємо статус
	updateStatus := func() {
		wallpaper, err := app.GetCurrentWallpaper()
		if err != nil {
			mStatus.SetTitle("Статус: помилка")
		} else {
			mStatus.SetTitle(fmt.Sprintf("📁 %s", filepath.Base(wallpaper)))
		}
	}
	updateStatus()

	// Запускаємо моніторинг в окремій горутині
	go app.StartWatching()

	// Обробка подій меню
	go func() {
		for {
			select {
			case <-mSync.ClickedCh:
				app.SyncWallpaper()
				updateStatus()

			case <-mToggle.ClickedCh:
				app.syncEnabled = !app.syncEnabled
				if app.syncEnabled {
					mToggle.SetTitle("✅ Автосинхронізація увімкнена")
				} else {
					mToggle.SetTitle("⬜ Автосинхронізація вимкнена")
				}

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	app.StopWatching()
}

// generateIcon створює просту іконку
func generateIcon() []byte {
	// Створюємо просту 16x16 ICO іконку (синій квадрат)
	width := 16
	height := 16

	// ICO header
	ico := []byte{
		0x00, 0x00, // Reserved
		0x01, 0x00, // Type (1 = ICO)
		0x01, 0x00, // Number of images
	}

	// ICO directory entry
	bmpSize := 40 + (width * height * 4) + (width * height / 8)
	ico = append(ico, []byte{
		byte(width),               // Width
		byte(height),              // Height
		0x00,                      // Color palette
		0x00,                      // Reserved
		0x01, 0x00,                // Color planes
		0x20, 0x00,                // Bits per pixel (32)
		byte(bmpSize),             // Size of BMP data
		byte(bmpSize >> 8),
		byte(bmpSize >> 16),
		byte(bmpSize >> 24),
		0x16, 0x00, 0x00, 0x00,    // Offset to BMP data
	}...)

	// BITMAPINFOHEADER
	ico = append(ico, []byte{
		0x28, 0x00, 0x00, 0x00, // Header size (40)
		byte(width), 0x00, 0x00, 0x00, // Width
		byte(height * 2), 0x00, 0x00, 0x00, // Height (doubled for ICO)
		0x01, 0x00, // Planes
		0x20, 0x00, // Bits per pixel
		0x00, 0x00, 0x00, 0x00, // Compression
		0x00, 0x00, 0x00, 0x00, // Image size
		0x00, 0x00, 0x00, 0x00, // X pixels per meter
		0x00, 0x00, 0x00, 0x00, // Y pixels per meter
		0x00, 0x00, 0x00, 0x00, // Colors used
		0x00, 0x00, 0x00, 0x00, // Important colors
	}...)

	// Pixel data (BGRA, bottom-up)
	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			// Створюємо градієнт синього кольору з іконкою монітора
			isMonitor := x >= 2 && x <= 13 && y >= 4 && y <= 11
			isScreen := x >= 3 && x <= 12 && y >= 5 && y <= 10
			isStand := x >= 6 && x <= 9 && y >= 12 && y <= 13
			isBase := x >= 4 && x <= 11 && y >= 14 && y <= 14

			var b, g, r, a byte
			if isScreen {
				// Екран - світло-блакитний
				b, g, r, a = 0xFF, 0xCC, 0x66, 0xFF
			} else if isMonitor || isStand || isBase {
				// Рамка - темно-сірий
				b, g, r, a = 0x44, 0x44, 0x44, 0xFF
			} else {
				// Прозорий фон
				b, g, r, a = 0x00, 0x00, 0x00, 0x00
			}
			ico = append(ico, b, g, r, a)
		}
	}

	// AND mask (всі 0 для повної непрозорості)
	for i := 0; i < width*height/8; i++ {
		ico = append(ico, 0x00)
	}

	return ico
}
