package cache

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
)

// DetectTerminalProtocol checks the environment to choose the best rendering method.
func DetectTerminalProtocol() string {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	if strings.Contains(term, "kitty") {
		return "kitty"
	}
	if termProgram == "iTerm.app" || termProgram == "WezTerm" {
		return "iterm2"
	}
	// Fallback to ANSI Truecolor block characters
	return "ansi"
}

// RenderImage returns the raw escape sequences or ASCII blocks to display the image.
func RenderImage(imagePath string, cellWidth, cellHeight int) string {
	file, err := os.Open(imagePath)
	if err != nil {
		return renderFallbackPlaceholder(cellWidth, cellHeight, "Image Unavailable")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return renderFallbackPlaceholder(cellWidth, cellHeight, "Decode Error")
	}

	protocol := DetectTerminalProtocol()
	switch protocol {
	case "kitty":
		return renderKitty(imagePath, cellWidth, cellHeight)
	case "iterm2":
		return renderITerm2(imagePath, cellWidth, cellHeight)
	default:
		return renderANSIBlocks(img, cellWidth, cellHeight)
	}
}

func renderKitty(imagePath string, cellWidth, cellHeight int) string {
	// Base64 encode the file content
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return renderFallbackPlaceholder(cellWidth, cellHeight, "IO Error")
	}

	b64Data := base64.StdEncoding.EncodeToString(data)
	
	// Kitty Graphics Protocol command format:
	// \x1b_Gf=100,a=T,t=d,c=W,r=H;<base64>\x1b\\
	// chunking into blocks of 4096 bytes
	var sb strings.Builder
	chunkSize := 4096
	totalLen := len(b64Data)

	for i := 0; i < totalLen; i += chunkSize {
		end := i + chunkSize
		m := 1 // More chunks
		if end >= totalLen {
			end = totalLen
			m = 0 // Last chunk
		}

		if i == 0 {
			// First chunk has layout params (f=100 means png/jpeg payload, a=T means display, c=cols, r=rows)
			sb.WriteString(fmt.Sprintf("\x1b_Gf=100,a=T,t=d,c=%d,r=%d,m=%d;", cellWidth, cellHeight, m))
		} else {
			sb.WriteString(fmt.Sprintf("\x1b_Gm=%d;", m))
		}
		sb.WriteString(b64Data[i:end])
		sb.WriteString("\x1b\\")
	}
	return sb.String()
}

func renderITerm2(imagePath string, cellWidth, cellHeight int) string {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return renderFallbackPlaceholder(cellWidth, cellHeight, "IO Error")
	}
	b64Data := base64.StdEncoding.EncodeToString(data)

	// iTerm2 inline images protocol:
	// \x1b]1337;File=inline=1;width=W;height=H;preserveAspectRatio=1:<base64>\a
	return fmt.Sprintf("\x1b]1337;File=inline=1;width=%d;height=%d;preserveAspectRatio=1:%s\a", cellWidth, cellHeight, b64Data)
}

func renderANSIBlocks(img image.Image, cellWidth, cellHeight int) string {
	bounds := img.Bounds()
	imgWidth := bounds.Max.X - bounds.Min.X
	imgHeight := bounds.Max.Y - bounds.Min.Y

	// Each TUI row has 2 block vertical segments (upper and lower half block)
	// Thus, pixel height = cellHeight * 2
	pixelWidth := cellWidth
	pixelHeight := cellHeight * 2

	var sb strings.Builder
	for y := 0; y < pixelHeight; y += 2 {
		for x := 0; x < pixelWidth; x++ {
			// Map destination pixel coordinates back to source image (nearest neighbor)
			srcX := bounds.Min.X + (x * imgWidth / pixelWidth)
			srcY1 := bounds.Min.Y + (y * imgHeight / pixelHeight)
			srcY2 := bounds.Min.Y + ((y + 1) * imgHeight / pixelHeight)

			c1 := img.At(srcX, srcY1)
			c2 := img.At(srcX, srcY2)

			r1, g1, b1, _ := c1.RGBA()
			r2, g2, b2, _ := c2.RGBA()

			// Convert from uint32 [0..65535] to uint8 [0..255]
			R1, G1, B1 := uint8(r1>>8), uint8(g1>>8), uint8(b1>>8)
			R2, G2, B2 := uint8(r2>>8), uint8(g2>>8), uint8(b2>>8)

			// Half block character renders foreground on top, background on bottom
			// Character: ▄ (lower half block)
			// Foreground color matches bottom pixel (c2)
			// Background color matches top pixel (c1)
			sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▄", R2, G2, B2, R1, G1, B1))
		}
		sb.WriteString("\x1b[0m\n")
	}

	return sb.String()
}

func renderFallbackPlaceholder(width, height int, text string) string {
	var buf bytes.Buffer
	padding := (width - len(text) - 2) / 2
	if padding < 0 {
		padding = 0
	}

	border := "+" + strings.Repeat("-", width-2) + "+"
	buf.WriteString(border)
	buf.WriteString("\n")
	
	middleLineY := height / 2
	for y := 1; y < height-1; y++ {
		if y == middleLineY {
			buf.WriteString("|")
			buf.WriteString(strings.Repeat(" ", padding))
			buf.WriteString(text)
			buf.WriteString(strings.Repeat(" ", width-2-padding-len(text)))
			buf.WriteString("|\n")
		} else {
			buf.WriteString("|")
			buf.WriteString(strings.Repeat(" ", width-2))
			buf.WriteString("|\n")
		}
	}
	buf.WriteString(border)
	return buf.String()
}
