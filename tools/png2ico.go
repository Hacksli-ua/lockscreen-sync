package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"os"

	"golang.org/x/image/draw"
)

func main() {
	// Читаємо PNG
	f, err := os.Open("icon.png")
	if err != nil {
		println("Error opening icon.png:", err.Error())
		return
	}
	srcImg, err := png.Decode(f)
	f.Close()
	if err != nil {
		println("Error decoding PNG:", err.Error())
		return
	}

	// Створюємо різні розміри
	sizes := []int{256, 48, 32, 16}
	pngImages := make([][]byte, len(sizes))

	for i, size := range sizes {
		// Масштабуємо
		dst := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.CatmullRom.Scale(dst, dst.Bounds(), srcImg, srcImg.Bounds(), draw.Over, nil)

		var buf bytes.Buffer
		png.Encode(&buf, dst)
		pngImages[i] = buf.Bytes()
	}

	// Створюємо ICO
	ico := createICO(pngImages, sizes)
	os.WriteFile("icon.ico", ico, 0644)
	println("icon.ico created successfully!")
}

func createICO(pngImages [][]byte, sizes []int) []byte {
	var buf bytes.Buffer

	// ICO Header
	binary.Write(&buf, binary.LittleEndian, uint16(0))          // Reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1))          // Type (1 = ICO)
	binary.Write(&buf, binary.LittleEndian, uint16(len(sizes))) // Number of images

	// Обчислюємо офсети
	headerSize := 6 + len(sizes)*16
	offset := headerSize

	// ICO Directory Entries
	for i, size := range sizes {
		w := uint8(size)
		h := uint8(size)
		if size >= 256 {
			w, h = 0, 0 // 0 означає 256
		}
		buf.WriteByte(w)                                                   // Width
		buf.WriteByte(h)                                                   // Height
		buf.WriteByte(0)                                                   // Color palette
		buf.WriteByte(0)                                                   // Reserved
		binary.Write(&buf, binary.LittleEndian, uint16(1))                 // Color planes
		binary.Write(&buf, binary.LittleEndian, uint16(32))                // Bits per pixel
		binary.Write(&buf, binary.LittleEndian, uint32(len(pngImages[i]))) // Size
		binary.Write(&buf, binary.LittleEndian, uint32(offset))            // Offset
		offset += len(pngImages[i])
	}

	// Image data
	for _, pngData := range pngImages {
		buf.Write(pngData)
	}

	return buf.Bytes()
}
