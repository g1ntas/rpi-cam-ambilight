package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"sync"
	"time"
)

//var calibrationRects = calculateLedAreas(leds[AxisX], leds[AxisY], ScreenWidth, ScreenHeight, 300)
//var screens = generateScreens()

type CalibrationJpegImage struct {
	buf   []byte
	index int
}

func (img *CalibrationJpegImage) Bytes() []byte {
	return img.buf
}

func (img *CalibrationJpegImage) Index() int {
	return img.index
}

func NewCalibrationJpegImage(index int, bytes []byte) *CalibrationJpegImage {
	if index < 0 {
		index = 0
	}
	return &CalibrationJpegImage{bytes, index}
}

func fillRGBARect(img *image.RGBA, rect *image.Rectangle, color color.Color) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			img.Set(x, y, color)
		}
	}
}

func generateCalibrationImages(buffer chan<- *CalibrationJpegImage, c *Config) {
	var wg sync.WaitGroup
	rects := calculateLedAreas(c.LedsX, c.LedsY, c.ScreenWidth, c.ScreenHeight, 300)
	for i, rect := range rects {
		wg.Add(1)
		go func(i int, rect *image.Rectangle) {
			defer wg.Done()
			rgba := image.NewRGBA(image.Rect(0, 0, c.ScreenWidth, c.ScreenHeight))
			b, err := createJpegWithFilledArea(rgba, rect, color.White)
			if err != nil {
				log.Printf("error occurred: %q", err)
				return
			}
			buffer <- NewCalibrationJpegImage(i, b)
		}(i, rect)
	}
	go func() {
		wg.Wait()
		close(buffer)
	}()
}

func createJpegWithFilledArea(rgba *image.RGBA, area *image.Rectangle, color color.Color) ([]byte, error) {
	partial := rgba.SubImage(*area).(*image.RGBA)
	draw.Draw(partial, partial.Bounds(), image.White, area.Min, draw.Src)
	fillRGBARect(rgba, area, color)
	buffer := new(bytes.Buffer)
	err := jpeg.Encode(buffer, rgba, nil)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func calibrate() {
	time.Sleep(time.Duration(5000) * time.Millisecond)
	// calculate position for each led within the screen
	// render image marking led position in white for each led
	// analyse camera image to determine led position within the camera
	//rgba := image.NewRGBA(image.Rect(0, 0, ScreenWidth, ScreenHeight))
	areas := make([][]image.Point, 96)
	//draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)
	//white := color.RGBA{255, 255, 255, 255}
	//black := color.RGBA{255, 255, 255, 255}
	/*for _, rect := range calibrationRects {
		//square := rgba.SubImage(rect).(*image.RGBA)
		//draw.Draw(square, square.Bounds(), image.White, rect.Min, draw.Src)
		fillRGBARect(rgba, rect, color.White)

		buffer := new(bytes.Buffer)
		err := jpeg.Encode(buffer, rgba, nil)
		if err != nil {
			log.Printf("error occurred: %q", err)
			continue
		}

		calibrationStream.UpdateJPEG(buffer.Bytes())
		time.Sleep(time.Duration(500) * time.Millisecond)
		areas = append(areas, findWhiteAreaInFrame())
		fillRGBARect(rgba, rect, color.Black)
	}*/

	/*for _, screen := range screens {
		calibrationStream.UpdateJPEG(screen)
		time.Sleep(time.Duration(500) * time.Millisecond)
		areas = append(areas, findWhiteAreaInFrame())
		//fillRGBARect(rgba, rect, color.Black)
	}*/

	b, err := camera.GetFrame()
	if err != nil {
		log.Printf("error occurred: %q", err)
		return
	}
	img, err := jpeg.Decode(bytes.NewReader(b))
	bd := img.Bounds()
	frame := image.NewRGBA(image.Rect(0, 0, bd.Dx(), bd.Dy()))
	draw.Draw(frame, frame.Bounds(), img, bd.Min, draw.Src)
	green := color.RGBA{0, 255, 0, 255}
	for _, area := range areas {
		for _, pt := range area {
			frame.Set(pt.X, pt.Y, green)
		}
	}
	buffer := new(bytes.Buffer)
	err = jpeg.Encode(buffer, frame, nil)
	if err != nil {
		log.Printf("error occurred: %q", err)
		return
	}
	calibrationStream.UpdateJPEG(buffer.Bytes())
	calibrationStream.UpdateJPEG(buffer.Bytes())
}

func findWhiteAreaInFrame() []image.Point {
	pixels := make([]image.Point, 0, 200)
	b, err := camera.GetFrame()
	if err != nil {
		log.Printf("error occurred: %q", err)
		return pixels
	}
	img, err := jpeg.Decode(bytes.NewReader(b))
	if err != nil {
		log.Printf("error occurred: %q", err)
		return pixels
	}
	bd := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bd.Dx(), bd.Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, bd.Min, draw.Src)
	for x := rgba.Rect.Min.X; x < rgba.Rect.Max.X; x++ {
		for y := rgba.Rect.Min.Y; y < rgba.Rect.Max.Y; y++ {
			col := rgba.RGBAAt(x, y)
			avg := float64(int(col.R)+int(col.G)+int(col.B)+int(col.A)) / 4
			if avg > 150 {
				pixels = append(pixels, image.Pt(x, y))
			}
		}
	}

	col := color.RGBA{0, 255, 0, 255}
	for _, px := range pixels {
		rgba.Set(px.X, px.Y, col)
	}
	//buffer := new(bytes.Buffer)
	//err = jpeg.Encode(buffer, rgba, nil)
	//if err != nil {
	//	log.Printf("error occurred: %q", err)
	//	return pixels
	//}
	//calibrationStream.UpdateJPEG(buffer.Bytes())
	//time.Sleep(time.Duration(500) * time.Millisecond)

	return pixels
}

func calculateLedAreas(ledsX, ledsY, screenWidth, screenHeight, ledDepth int) []*image.Rectangle {
	pos := make([]*image.Rectangle, 0, ledsX*2+ledsY*2)
	// horizontal edges
	widthX := divideScreenForLeds(screenWidth, ledsX)
	for _, y := range []int{0, screenHeight - ledDepth} {
		for xi, x := 0, 0; x < screenWidth; x, xi = x+widthX[xi], xi+1 {
			r := image.Rect(x, y, x+widthX[xi], y+ledDepth)
			pos = append(pos, &r)
		}
	}
	// vertical edges
	widthY := divideScreenForLeds(screenHeight, ledsY)
	for _, x := range []int{0, screenWidth - ledDepth} {
		for yi, y := 0, 0; y < screenHeight; y, yi = y+widthY[yi], yi+1 {
			r := image.Rect(x, y, x+ledDepth, y+widthY[yi])
			pos = append(pos, &r)
		}
	}
	return pos
}

func drawCalibrationImage(x0, y0, x1, y1, screenWidth, screenHeight int) *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))
	draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)
	square := rgba.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
	draw.Draw(square, square.Bounds(), image.White, image.Point{x0, y0}, draw.Src)
	return rgba
}

func divideScreenForLeds(maxPixels, amountOfLeds int) []int {
	r := make([]int, amountOfLeds)
	val := maxPixels / amountOfLeds
	for i := range r {
		r[i] = val
	}
	remainder := maxPixels % amountOfLeds
	for i := 0; i < remainder; i++ {
		r[i]++
	}
	return r
}
