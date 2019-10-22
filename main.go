package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/hybridgroup/mjpeg"
	"github.com/technomancers/piCamera"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// raspistill -o img.jpg --exif none --mode 7 -w 640 -h 480 -th none -q 20 -e jpg
// raspistill options:
// -o -   : capture img and output into stdout
// -e jpg : set jpeg encoding
// --exif none : doesn't save any exif metadata
// -md 6 : captures in 1280x720 resolution
// -ex off : turn off automatic exposure and gain control
// -drc off : disable dynamic range (feature useful in dark areas)

// Deps:
// https://github.com/technomancers/piCamera - raspivid wrapper to capture video frames

// Points
// Top Left: 52:37.4
// Bottom Left: 9:325.4
// Bottom Right: 616.6:326.4
// Top Right: 573.4:34.6

type Config struct {
	ScreenWidth int `json:"screenWidth"`
	ScreenHeight int `json:"screenHeight"`
	LedsX int `json:"ledsX"`
	LedsY int `json:"ledsY"`
	dir string
}

const (
	AxisX = iota
	AxisY
)

const Width = 640
const Height = 480

const ScreenWidth = 3840
const ScreenHeight = 2160

var leds = [2]int{31, 17}

var points = map[string]image.Point{
	"topLeft": image.Pt(58, 70),
	"topRight": image.Pt(537, 61),
	"bottomLeft": image.Pt(27, 343),
	"bottomRight": image.Pt(560, 334),
}

var defishStr float64
var defishZoom float64

var ledMap = generateLedMap()
var ledColors = make([]*color.RGBA, leds[AxisX]*2+leds[AxisY]*2)

var camera *piCamera.PiCamera
var stream *mjpeg.Stream
var cameraStream *mjpeg.Stream
var calibrationStream *mjpeg.Stream

var calibrationRGBA *image.RGBA
//var calibrationScreens []image.Image

var Conf = &Config{
	dir: "./.rpicam-ambilight",
}

func (c *Config) CalibrationScreenDir() string {
	return filepath.Join(c.dir, "calibration")
}

func (c *Config) CalibrationScreenPath(index int) string {
	return filepath.Join(c.CalibrationScreenDir(), fmt.Sprintf("%d.jpg", index))
}

func (c *Config) HasCalibrationSettingsSet() bool {
	return c.ScreenWidth > 0 && c.ScreenHeight > 0 && c.LedsX > 0 && c.LedsY > 0
}

func createOrCleanUpDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}


func handleError(err error) (ok bool) {
	if err != nil {
		fmt.Printf("Error occured: %s", err)
		return false
	}
	return true
}

func main() {
	if ok := handleError(Conf.Read()); !ok {
		return
	}
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "init":
			runInitCmd()
		case "calibrate":
			if !Conf.HasCalibrationSettingsSet() {
				fmt.Printf("Missing or invalid configuration: run \"./%s init\"", os.Args[0])
				return
			}
			camera, err := startCamera()
			if ok := handleError(err); !ok {
				return
			}
			defer camera.Stop()
			serveCameraStream(camera)
			serveCalibrationStream()
			go startServer()
			fmt.Println("Started camera stream at http://127.0.0.1:8081/camera")
			fmt.Println("Adjust camera placement to it's permanent position and make sure whole screen is visible")
			fmt.Println()
			fmt.Println("Started calibration server at http://127.0.0.1:8081/calibration")
			fmt.Println("Open website on calibrated screen and make it full screen")
			fmt.Println("When you are ready press enter to start calibration process")
			reader := bufio.NewReader(os.Stdin)
			_, err = reader.ReadString('\n')
			if ok := handleError(err); !ok {
				return
			}
			// todo: show each of pre-generated calibration screen images
			// todo: capture camera frame and store coordinates of all white pixels
			// todo: save coordinates into config file
			// todo: show calibrated result image with highlighted areas
		case "run":
			// todo: if config or calibration coordinates doesn't exist, show error
			// todo: capture frame
			// todo: analyze avg color of each led position based on calibration coordinates
			// todo: set led color
		default:
			fmt.Printf("Unrecognized command: %s", os.Args[1])
			return
		}
	}
return
	//var err error
	/*led, err := NewWS2801Led(rpio.Spi0, 96)
	if err != nil {
		log.Printf("error occurred: %q", err)
		return
	}
	defer led.Close()
	for i := 0; i < led.Count; i++ {
		err = led.UpdatePixel(i, 0, 255, 0)
		if err != nil {
			log.Printf("error occurred: %q", err)
			return
		}
	}*/
	//calibrationRGBA = drawCalibrationImage(leds[AxisX], leds[AxisY], ScreenWidth, ScreenHeight)
	/*args := piCamera.NewArgs()
	args.Width = 1640
	args.Height = 1232
	args.Mode = 4
	args.ExposureMode = piCamera.ExpVerylong
	args.Brightness = 60
	args.Contrast = 40
	camera, err = piCamera.New(nil, args)
	if err != nil {
		log.Printf("error occurred: %q", err)
		return
	}
	err = camera.Start()
	defer camera.Stop()
	if err != nil {
		log.Printf("error occurred: %q", err)
		return
	}

	//stream = mjpeg.NewStream()
	calibrationStream = mjpeg.NewStream()*/

	//go mjpegCapture()

	//http.Handle("/stream", stream)
	/*http.Handle("/calibration", calibrationStream)
	http.HandleFunc("/calibrate", func(w http.ResponseWriter, r *http.Request) {
		go calibrate()
		http.ServeFile(w, r, "./calibration.html")
	})
	http.HandleFunc("/leds", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		response := struct {
			Positions []*image.Rectangle `json:"positions"`
			Colors    []*color.RGBA      `json:"colors"`
		}{ledMap, ledColors}
		encoder := json.NewEncoder(w)
		err := encoder.Encode(response)
		if err != nil {
			log.Printf("error occurred: %q", err)
		}
	})
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Fatal(http.ListenAndServe(":8081", nil))*/
}

func serveCameraStream(cam *piCamera.PiCamera) {
	cameraStream = mjpeg.NewStream()
	http.Handle("/camera-stream", cameraStream)
	http.HandleFunc("/camera", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "camera.html")
	})
	go mjpegCapture(cam)
}

func serveCalibrationStream() {
	calibrationStream = mjpeg.NewStream()
	http.Handle("/calibration-stream", calibrationStream)
	http.HandleFunc("/calibration", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "calibration.html")
	})
}

func startServer() {
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func startCamera() (*piCamera.PiCamera, error) {
	args := piCamera.NewArgs()
	args.Width = 1640
	args.Height = 1232
	args.Mode = 4
	args.ExposureMode = piCamera.ExpVerylong
	args.Brightness = 60
	args.Contrast = 40
	camera, err := piCamera.New(nil, args)
	if err != nil {
		return nil, err
	}
	err = camera.Start()
	if err != nil {
		return nil, err
	}
	return camera, nil
}

func runInitCmd() {
	if len(os.Args) < 6 {
		fmt.Printf("Usage: %s init {screen_width} {screen_height} {amount_of_leds_x} {amount_of_leds_y}", os.Args[0])
		return
	}
	var err error
	Conf.ScreenWidth, err = strconv.Atoi(os.Args[2])
	if ok := handleError(err); !ok {
		return
	}
	Conf.ScreenHeight, err = strconv.Atoi(os.Args[3])
	if ok := handleError(err); !ok {
		return
	}
	Conf.LedsX, err = strconv.Atoi(os.Args[4])
	if ok := handleError(err); !ok {
		return
	}
	Conf.LedsY, err = strconv.Atoi(os.Args[5])
	if ok := handleError(err); !ok {
		return
	}
	if !Conf.HasCalibrationSettingsSet() {
		handleError(fmt.Errorf("entered arguments must be greater than zero"))
		return
	}
	if ok := handleError(Conf.Write()); !ok {
		return
	}
	dir := Conf.CalibrationScreenDir()
	if ok := handleError(createOrCleanUpDir(dir)); !ok {
		return
	}
	fmt.Println("Generating calibration screens...")
	var wg sync.WaitGroup
	buffer := make(chan *CalibrationJpegImage)
	generateCalibrationImages(buffer, Conf)
	for img := range buffer {
		wg.Add(1)
		go func(img *CalibrationJpegImage) {
			defer wg.Done()
			err := ioutil.WriteFile(Conf.CalibrationScreenPath(img.Index()), img.Bytes(), 0644)
			handleError(err)
		}(img)
	}
	wg.Wait()
	fmt.Println("\nDone.")
}

func (c *Config) Read() error {
	f, err := os.Open(c.Dest())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()
	d := json.NewDecoder(bufio.NewReader(f))
	err = d.Decode(c)
	return err
}

func (c *Config) Write() error {
	err := os.MkdirAll(c.dir, os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.Create(c.Dest())
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()
	w := bufio.NewWriter(f)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(c)
	if err != nil {
		return err
	}
	err = w.Flush()
	return err
}

func (c *Config) Dest() string {
	return filepath.Join(c.dir, "config.json")
}


func mjpegCapture(cam *piCamera.PiCamera) {
	for {
		b, err := cam.GetFrame()
		if err != nil {
			log.Printf("error occurred: %q", err)
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}

		//img, err := jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			log.Printf("error occurred: %q", err)
			continue
		}

		/*log.Printf("Type %s", reflect.TypeOf(img))
		if rgba, ok := img.(*image.RGBA); ok {
			log.Printf("Defishing...")
		}*/
		//img = defish(img, defishStr, defishZoom)

		//buffer := new(bytes.Buffer)
		//err = jpeg.Encode(buffer, img, nil)
		if err != nil {
			log.Printf("error occurred: %q", err)
			continue
		}

		/*for i, rect := range ledMap {
			go computeColor(i, &img, rect)
		}*/
		//stream.UpdateJPEG(buffer.Bytes())
		cameraStream.UpdateJPEG(b)
	}
}

func defish(img image.Image, strength, zoom float64) *image.RGBA {
	b := img.Bounds()
	b2 := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	rgba2 := image.NewRGBA(image.Rect(0, 0, b2.Dx(), b2.Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)
	draw.Draw(rgba2, rgba2.Bounds(), img, b2.Min, draw.Src)

	if strength == 0 {
		strength = 0.00001
	}

	width := rgba.Rect.Dx()
	height := rgba.Rect.Dy()

	halfW, halfH := float64(width)/2, float64(height)/2

	correctionRadius := math.Sqrt(float64(width*width+height*height))/strength
	for x := rgba.Rect.Min.X; x < rgba.Rect.Max.X; x++ {
		for y := rgba.Rect.Min.Y; y < rgba.Rect.Max.Y; y++ {
			newX, newY := float64(x) - halfW, float64(y) -halfH
			distance := math.Sqrt(newX*newX+newY*newY)
			r := distance / correctionRadius

			var theta float64
			if r == 0 {
				theta = 1
			} else {
				theta = math.Atan(r) / r
			}

			srcX := halfW + theta * newX * zoom
			srcY := halfH + theta * newY * zoom

			var col color.Color = color.RGBA{50, 150, 100, 255}
			//if srcX >= float64(rgba.Bounds().Min.X) && srcX < float64(rgba.Bounds().Max.X) && srcY >= float64(rgba.Bounds().Min.Y) && srcY < float64(rgba.Bounds().Max.Y) {
			if image.Pt(int(srcX), int(srcY)).In(rgba.Rect) {
				//min, max := math.Ceil(srcX)
				// cases:
				// (x1.1, y1.1) = (x1*0.9, x2*0.1, y1*0.9, y2*0.1)
				// (x1.9, y1.9) = (x1*0.1, x2*0.9, y1*0.1, y2*0.9)
				x0, x1 := math.Floor(srcX), math.Ceil(srcX)
				y0, y1 := math.Floor(srcY), math.Ceil(srcY)
				wx0, wx1 := x1 - srcX, srcX - x0
				wy0, wy1 := y1 - srcY, srcY - y0
				wx0y0 := wx0*wy0
				wx0y1 := wx0*wy1
				wx1y0 := wx1*wy0
				wx1y1 := wx1*wy1
				//srx = math.Round()
				// todo: blend colors
				col1 := rgba2.RGBAAt(int(x0), int(y0))
				col2 := rgba2.RGBAAt(int(x0), int(y1))
				col3 := rgba2.RGBAAt(int(x1), int(y0))
				col4 := rgba2.RGBAAt(int(x1), int(y1))

				cR := float64(col1.R)*float64(col1.R)*wx0y0 + float64(col2.R)*float64(col2.R)*wx0y1 + float64(col3.R)*float64(col3.R)*wx1y0 + float64(col4.R)*float64(col4.R)*wx1y1
				cG := float64(col1.G)*float64(col1.G)*wx0y0 + float64(col2.G)*float64(col2.G)*wx0y1 + float64(col3.G)*float64(col3.G)*wx1y0 + float64(col4.G)*float64(col4.G)*wx1y1
				cB := float64(col1.B)*float64(col1.B)*wx0y0 + float64(col2.B)*float64(col2.B)*wx0y1 + float64(col3.B)*float64(col3.B)*wx1y0 + float64(col4.B)*float64(col4.B)*wx1y1
				cA := float64(col1.A)*float64(col1.A)*wx0y0 + float64(col2.A)*float64(col2.A)*wx0y1 + float64(col3.A)*float64(col3.A)*wx1y0 + float64(col4.A)*float64(col4.A)*wx1y1

				col = color.RGBA{
					R: uint8(math.Sqrt(cR / 4)),
					G: uint8(math.Sqrt(cG / 4)),
					B: uint8(math.Sqrt(cB / 4)),
					A: uint8(math.Sqrt(cA / 4)),
				}
			}

			rgba.Set(x, y, col)
		}
	}
	return rgba
}

func computeColor(index int, img *image.Image, rect *image.Rectangle) {
	rgba := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(rgba, rgba.Bounds(), *img, rect.Min, draw.Src)
	var sumR, sumG, sumB, sumA int
	for x := 0; x < rgba.Rect.Max.X; x++ {
		for y := 0; y < rgba.Rect.Max.Y; y++ {
			col := rgba.RGBAAt(x, y)
			sumR += int(col.R) * int(col.R)
			sumG += int(col.G) * int(col.G)
			sumB += int(col.B) * int(col.B)
			sumA += int(col.A) * int(col.A)
		}
	}
	totalPixels :=  rgba.Rect.Max.X * rgba.Rect.Max.Y
	avgColor := color.RGBA{
		R: uint8(math.Sqrt(float64(sumR / totalPixels))),
		G: uint8(math.Sqrt(float64(sumG / totalPixels))),
		B: uint8(math.Sqrt(float64(sumB / totalPixels))),
		A: uint8(math.Sqrt(float64(sumA / totalPixels))),
	}
	ledColors[index] = &avgColor
}

func avgColor(c1, c2 *color.RGBA) *color.RGBA {
	r := float64(c1.R) * float64(c2.R)
	g := float64(c1.G) * float64(c2.G)
	b := float64(c1.B) * float64(c2.B)
	a := float64(c1.A) * float64(c2.A)

	return &color.RGBA{
		R: uint8(math.Sqrt(r / 2)),
		G: uint8(math.Sqrt(g / 2)),
		B: uint8(math.Sqrt(b / 2)),
		A: uint8(math.Sqrt(a / 2)),
	}
}

func generateLedMap() []*image.Rectangle {
	ledMap := make([]*image.Rectangle, leds[AxisX]*2+leds[AxisY]*2)

	edges := [4]struct {
		from image.Point
		to image.Point
		axis int
		calcRectDest func(x0, y0, stepX, stepY int) (x1, y1 int)
	}{
		{points["topLeft"], points["topRight"], AxisX, func(x0, y0, stepX, stepY int) (x1, y1 int) {
			return x0 + stepX, y0 + 50
		}},
		{points["topRight"], points["bottomRight"], AxisY, func(x0, y0, stepX, stepY int) (x1, y1 int) {
			return x0 - 50, y0 + stepY
		}},
		{points["bottomRight"], points["bottomLeft"], AxisX, func(x0, y0, stepX, stepY int) (x1, y1 int) {
			return x0 - stepX, y0 - 50
		}},
		{points["bottomLeft"], points["topLeft"], AxisY, func(x0, y0, stepX, stepY int) (x1, y1 int) {
			return x0 + 50, y0 - stepY
		}},
	}

	idx := 0
	for _, edge := range edges {
		amount := leds[edge.axis]
		diffX := edge.to.X - edge.from.X
		diffY := edge.to.Y - edge.from.Y
		stepX := diffX / amount
		stepY := diffY / amount
		remainderX := diffX % amount
		remainderY := diffY % amount

		x0 := edge.from.X
		y0 := edge.from.Y
		for i := 0; i < amount; i++ {
			if i != 0 {
				x0 = x0 + stepX
				y0 = y0 + stepY
			}
			if remainderX != 0 {
				x0 += signum(remainderX)
				remainderX += -signum(remainderX)
			}
			if remainderY != 0 {
				y0 += signum(remainderY)
				remainderY += -signum(remainderY)
			}
			x1, y1 := edge.calcRectDest(x0, y0, stepX, stepY)
			rect := image.Rect(x0, y0, x1, y1)
			ledMap[idx] = &rect
			idx++
		}
	}
	return ledMap
}

func signum(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}