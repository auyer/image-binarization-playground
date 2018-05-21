package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math"
	"os"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var infile = flag.String("infile", "img.png", "path to image (gif, jpeg, png)")
var nNeighborhood = flag.Int("n", 10, "n-neighborhood (Integer)")
var k = 0.5
var r = float64(128)

// Reduce data from a 256 length array to a 50 length
func reduce(list []int) []int {
	if len(list) > 5 {

		var v int
		for i := range list[:5] {
			v += list[i]
		}
		return append(reduce(list[5:]), v)
	}
	return make([]int, 0)

}

func average(list []int) int {
	var a int
	for i := range list {
		a += list[i]
	}
	return a / len(list)
}

func flatten(img *image.Gray, x int, y int, n int) []int {
	// img.Bounds().Max.Y
	var list []int
	for yi := y - n/2; yi < y+n/2; yi++ {
		for xi := x - n/2; xi < x+n/2; xi++ {
			list = append(list, int(img.GrayAt(yi, yi).Y))
		}
	}
	return list
}
func max(list []int) float64 {
	if len(list) > 1 {
		return math.Max(max(list[1:]), float64(list[len(list)-1]))
	}
	return float64(list[len(list)-1])
}
func min(list []int) float64 {
	if len(list) > 1 {
		return math.Min(min(list[1:]), float64(list[len(list)-1]))
	}
	return float64(list[len(list)-1])
}

func floatify(ints []int) []float64 {
	floats := make([]float64, len(ints))
	for idx, val := range ints {
		floats[idx] = float64(val)
	}
	return floats
}

func barnsen(oldPixel color.Gray, pixlist []int, limiarImg *image.Gray, x int, y int) {
	if int(oldPixel.Y) >= int((max(pixlist)+min(pixlist))/2) {
		limiarImg.Set(x, y, color.GrayModel.Convert(color.RGBA{255, 255, 255, 255}))
	} else {
		limiarImg.Set(x, y, color.GrayModel.Convert(color.RGBA{0, 0, 0, 255}))
	}
}
func niblack(oldPixel color.Gray, pixlist []int, limiarImg *image.Gray, x int, y int) {
	if int(oldPixel.Y) >= average(pixlist)+int(k*math.Sqrt(stat.Variance(floatify(pixlist), nil))) {
		limiarImg.Set(x, y, color.GrayModel.Convert(color.RGBA{255, 255, 255, 255}))
	} else {
		limiarImg.Set(x, y, color.GrayModel.Convert(color.RGBA{0, 0, 0, 255}))
	}

}

func suapiet(oldPixel color.Gray, pixlist []int, limiarImg *image.Gray, x int, y int) {
	if int(oldPixel.Y) >= average(pixlist)+int(1+k*(math.Sqrt((stat.Variance(floatify(pixlist), nil))/r)-1)) {
		limiarImg.Set(x, y, color.GrayModel.Convert(color.RGBA{255, 255, 255, 255}))
	} else {
		limiarImg.Set(x, y, color.GrayModel.Convert(color.RGBA{0, 0, 0, 255}))
	}

}

func globalLimiar(oldPixel color.Gray, avg int, limiarImg *image.Gray, x int, y int) {
	if int(oldPixel.Y) >= avg {
		limiarImg.Set(x, y, color.GrayModel.Convert(color.RGBA{255, 255, 255, 255}))
	} else {
		limiarImg.Set(x, y, color.GrayModel.Convert(color.RGBA{0, 0, 0, 255}))
	}
}

func main() {

	flag.Parse()
	reader, err := os.Open(*infile)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	bounds := m.Bounds()
	img := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldPixel := m.At(x, y)
			pixel := color.GrayModel.Convert(oldPixel)
			img.Set(x, y, pixel)
		}
	}

	outGray, err := os.Create("grayScale.png")
	if err != nil {
		log.Fatalf("Error creating file %s: %v", "grayScale.png", err)
	}
	png.Encode(outGray, img)
	// HISTOGRAM

	histogram := make([]int, 256)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			histogram[img.GrayAt(x, y).Y]++
		}
	}
	//CONDENSED HISTOGRAM
	cHistogram := reduce(histogram)

	for i, x := range cHistogram {
		fmt.Printf("%d: %6d \n", i, x)
		//		fmt.Printf("0x%04x-0x%04x: %6d \n", i, (i + 1), x)
	}

	// calculate average color

	//Print the results
	val := make(plotter.Values, 50)
	for i := range val {
		val[i] = float64(cHistogram[i])
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Histogram"
	h, err := plotter.NewBarChart(val, vg.Points(50))
	if err != nil {
		panic(err)
	}
	p.Add(h)
	// Save the plot to a PNG file.
	if err := p.Save(50*vg.Inch, 30*vg.Inch, "hist.png"); err != nil {
		panic(err)
	}

	// GLOBAL LIMIAR
	// Average luminosity
	var a int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			a += int(img.GrayAt(x, y).Y)
		}
	}
	avg := a / (bounds.Max.Y * bounds.Max.X)
	print(avg)

	limiarImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldPixel := img.GrayAt(x, y)
			go globalLimiar(oldPixel, avg, limiarImg, x, y)
		}
	}
	outFile, err := os.Create("limiarGlobal.png")
	if err != nil {
		log.Fatalf("Error creating file %s: %v", "limiarGlobal.png", err)
	}
	png.Encode(outFile, limiarImg)

	// LOCAL LIMIAR (Bernsen)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixlist := flatten(img, x, y, *nNeighborhood)
			oldPixel := img.GrayAt(x, y)
			go barnsen(oldPixel, pixlist, limiarImg, x, y)

		}
	}
	outFile2, err := os.Create("binarBernsen.png")
	if err != nil {
		log.Fatalf("Error creating file %s: %v", "binarBernsen.png", err)
	}
	png.Encode(outFile2, limiarImg)

	// LOCAL LIMIAR (Niblack)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixlist := flatten(img, x, y, *nNeighborhood)
			oldPixel := img.GrayAt(x, y)
			go niblack(oldPixel, pixlist, limiarImg, x, y)
		}
	}
	outFile3, err := os.Create("binarNiblack.png")
	if err != nil {
		log.Fatalf("Error creating file %s: %v", "binarNiblack.png", err)
	}
	png.Encode(outFile3, limiarImg)

	//Sauvola e Pietaksinen

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixlist := flatten(img, x, y, *nNeighborhood)
			oldPixel := img.GrayAt(x, y)
			go suapiet(oldPixel, pixlist, limiarImg, x, y)

		}
	}
	outFile4, err := os.Create("binarSauvola.png")
	if err != nil {
		log.Fatalf("Error creating file %s: %v", "binarSauvola.png", err)
	}
	png.Encode(outFile4, limiarImg)

}
