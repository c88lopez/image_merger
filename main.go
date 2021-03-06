package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type job struct {
	rowNumber int
	record    []string
}

var wg sync.WaitGroup

// main func
func main() {
	now := time.Now()

	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)
	config := Config{}

	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	file, err = os.Open(config.CsvPath)
	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(bufio.NewReader(file))
	record, err := reader.Read()

	rowNumber := 2

	jobs := make(chan job, 100)

	for w := 1; w <= 25; w++ {
		go worker(jobs)
	}

	for {
		record, err = reader.Read()

		// only for debug
		if err == io.EOF {
			break
		}

		wg.Add(1)
		// go mergeImages(rowNumber, record[26], record[5])
		jobs <- job{rowNumber, record}
		rowNumber++
	}

	wg.Wait()

	fmt.Printf("%s\n", time.Since(now))
}

// worker func
func worker(jobs <-chan job) {
	for j := range jobs {
		mergeImages(j.rowNumber, j.record[26], j.record[5])
	}
}

// mergeImages func
func mergeImages(rowNumber int, baseURL string, logoURL string) {
	getImages(baseURL, logoURL, rowNumber)
	glueImages(rowNumber)
}

// getImage func
func getImages(baseURL string, logoURL string, rowNumber int) {
	getImage(baseURL, rowNumber, "base")
	getImage(logoURL, rowNumber, "logo")
}

// glueImages func
func glueImages(rowNumber int) {
	var basePath bytes.Buffer
	var logoPath bytes.Buffer
	var mergedPath bytes.Buffer

	basePath.WriteString("tmp/base/")
	basePath.WriteString(strconv.Itoa(rowNumber))

	bf, err := os.Open(basePath.String())
	if err != nil {
		panic(err)
	}

	baseImg, err := png.Decode(bufio.NewReader(bf))
	if err != nil {
		panic(err)
	}

	base := image.NewRGBA(image.Rect(baseImg.Bounds().Min.X, baseImg.Bounds().Min.Y, baseImg.Bounds().Max.X, baseImg.Bounds().Max.Y))

	logoPath.WriteString("tmp/logo/")
	logoPath.WriteString(strconv.Itoa(rowNumber))

	lf, err := os.Open(logoPath.String())
	if err != nil {
		panic(err)
	}

	logoImg, err := jpeg.Decode(bufio.NewReader(lf))
	if err != nil {
		panic(err)
	}

	draw.Draw(base, baseImg.Bounds(), baseImg, image.Point{0, 0}, draw.Src)
	draw.Draw(base, logoImg.Bounds(), logoImg, image.Point{15, 15}, draw.Src)

	mergedPath.WriteString("images/")
	mergedPath.WriteString(strconv.Itoa(rowNumber))
	mergedPath.WriteString(".png")

	mergeFile, err := os.Create(mergedPath.String())
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(mergeFile)
	png.Encode(w, base)

	defer os.Remove(basePath.String())
	defer os.Remove(logoPath.String())

	fmt.Printf("%s %s\n", "Generated", mergedPath.String())

	wg.Done()
}

// getBaseImage func
func getImage(URL string, rowNumber int, imageType string) {
	var imagePath bytes.Buffer
	imagePath.WriteString("tmp/")
	imagePath.WriteString(imageType)
	imagePath.WriteString("/")
	imagePath.WriteString(strconv.Itoa(rowNumber))

	out, err := os.Create(imagePath.String())
	if err != nil {
		panic(err)
	}

	response, err := http.Get(URL)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, response.Body)
	if err != nil {
		panic(err)
	}
}
