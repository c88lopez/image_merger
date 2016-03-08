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

	for {
		record, err = reader.Read()

		// only for debug
		if err == io.EOF {
			break
		}

		wg.Add(1)
		go mergeImages(rowNumber, record[26], record[5])
		rowNumber++
	}

	wg.Wait()

	fmt.Printf("%s\n", time.Since(now))
}

// mergeImages func
func mergeImages(rowNumber int, baseURL string, logoURL string) {
	getImages(baseURL, logoURL, rowNumber)
	glueImages(rowNumber)
}

// getImage func
func getImages(baseRUL string, logoURL string, rowNumber int) {
	getBaseImage(baseRUL, rowNumber)
	getLogoImage(logoURL, rowNumber)
}

// glueImages func
func glueImages(rowNumber int) {
	var basePath bytes.Buffer
	var logoPath bytes.Buffer
	var mergedPath bytes.Buffer

	basePath.WriteString("tmp/images/base/")
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

	logoPath.WriteString("tmp/images/logo/")
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

	mergedPath.WriteString("tmp/images/merged/")
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

	wg.Done()
}

// getBaseImage func
func getBaseImage(baseRUL string, rowNumber int) {
	var imagePath bytes.Buffer
	imagePath.WriteString("tmp/images/base/")
	imagePath.WriteString(strconv.Itoa(rowNumber))

	out, err := os.Create(imagePath.String())
	if err != nil {
		panic(err)
	}

	response, err := http.Get(baseRUL)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, response.Body)
	if err != nil {
		panic(err)
	}
}

// getLogoImage func
func getLogoImage(logoURL string, rowNumber int) {
	var imagePath bytes.Buffer
	imagePath.WriteString("tmp/images/logo/")
	imagePath.WriteString(strconv.Itoa(rowNumber))

	out, err := os.Create(imagePath.String())
	if err != nil {
		panic(err)
	}

	response, err := http.Get(logoURL)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, response.Body)
	if err != nil {
		panic(err)
	}
}
