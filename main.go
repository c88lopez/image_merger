package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
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

	file, err := os.Open("pf_mobile_tw_corrected.csv")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", file)

	reader := csv.NewReader(bufio.NewReader(file))

	record, err := reader.Read()

	rowNumber := 2

	for {
		record, err = reader.Read()

		if err == io.EOF {
			break
		}

		wg.Add(1)
		go mergeImages(rowNumber, record[22], record[5])
		rowNumber++
	}

	wg.Wait()

	fmt.Printf("%s\n", time.Since(now))
}

// mergeImages func
func mergeImages(rowNumber int, baseURL string, logoURL string) {
	fmt.Println("@mergeImages")

	getImages(baseURL, logoURL, rowNumber)

	wg.Done()
}

// getImage func
func getImages(baseRUL string, logoURL string, rowNumber int) {
	getBaseImage(baseRUL, rowNumber)
	getLogoImage(logoURL, rowNumber)
}

// getBaseImage func
func getBaseImage(baseRUL string, rowNumber int) {
	fmt.Printf("> Image: %s\n", baseRUL)

	var imagePath bytes.Buffer
	imagePath.WriteString("images/base/")

	imagePath.WriteString(strconv.Itoa(rowNumber))
	imagePath.WriteString(".png")

	fmt.Printf("> Image path: %s\n", imagePath.String())
	out, err := os.Create(imagePath.String())
	if err != nil {
		panic(err)
	}

	fmt.Printf("> Getting: %s\n", baseRUL)
	response, err := http.Get(baseRUL)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Base %d downloaded.\n", rowNumber)
}

// getLogoImage func
func getLogoImage(logoURL string, rowNumber int) {
	fmt.Printf("> Image: %s\n", logoURL)

	var imagePath bytes.Buffer
	imagePath.WriteString("images/logo/")

	imagePath.WriteString(strconv.Itoa(rowNumber))
	imagePath.WriteString(".gif")

	fmt.Printf("> Image path: %s\n", imagePath.String())
	out, err := os.Create(imagePath.String())
	if err != nil {
		panic(err)
	}

	fmt.Printf("> Getting: %s\n", logoURL)
	response, err := http.Get(logoURL)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Logo %d downloaded.\n", rowNumber)
}
