package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	// Створюємо PNG іконки різних розмірів
	sizes := []int{256, 48, 32, 16}

	for _, size := range sizes {
		img := createMonitorIcon(size)
		filename := "icon.png"
		if size != 256 {
			continue // Створюємо тільки 256x256 для go-winres
		}
		f, _ := os.Create(filename)
		png.Encode(f, img)
		f.Close()

		// Копіюємо в winres
		f2, _ := os.Create("winres/icon.png")
		f, _ = os.Open(filename)
		img2, _ := png.Decode(f)
		png.Encode(f2, img2)
		f.Close()
		f2.Close()
	}

	println("Icon created successfully!")
}

func createMonitorIcon(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Масштабування
	scale := float64(size) / 32.0

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Нормалізовані координати (0-32)
			nx := float64(x) / scale
			ny := float64(y) / scale

			var c color.RGBA

			// Монітор (рамка)
			isFrame := nx >= 4 && nx <= 28 && ny >= 5 && ny <= 21
			isScreen := nx >= 6 && nx <= 26 && ny >= 7 && ny <= 19
			isStand := nx >= 14 && nx <= 18 && ny >= 22 && ny <= 25
			isBase := nx >= 10 && nx <= 22 && ny >= 26 && ny <= 28

			if isScreen {
				// Градієнт екрану (синьо-фіолетовий як Windows)
				t := (nx - 6) / 20.0
				r := uint8(100 + t*80)
				g := uint8(120 + t*60)
				b := uint8(200 + t*40)
				c = color.RGBA{r, g, b, 255}
			} else if isFrame && !isScreen {
				// Рамка монітора - темно-сірий
				c = color.RGBA{50, 50, 55, 255}
			} else if isStand {
				// Ніжка - сірий
				c = color.RGBA{60, 60, 65, 255}
			} else if isBase {
				// База - сірий
				c = color.RGBA{70, 70, 75, 255}
			} else {
				// Прозорий фон
				c = color.RGBA{0, 0, 0, 0}
			}

			img.SetRGBA(x, y, c)
		}
	}

	return img
}
