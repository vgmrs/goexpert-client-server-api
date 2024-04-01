package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	quotationURL string        = "https://localhost:8080/cotacao"
	timeout      time.Duration = 300 * time.Millisecond
	fileName     string        = "cotacao.txt"
)

func getQuotation() (map[string]string, error) {
	var quotation map[string]string

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", quotationURL, nil)
	if err != nil {
		return quotation, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return quotation, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&quotation)
	if err != nil {
		return quotation, err
	}

	log.Println("Quotation getting successfully!")

	return quotation, nil
}

func saveQuotation(bid string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("Dólar: %s", bid)
	data := []byte(text)

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	log.Println("Quotation saved successfully!")

	return nil
}

func main() {
	log.Println("Starting client...")

	quotation, err := getQuotation()
	if err != nil {
		log.Fatalf("Error getting quotation. %v", err)
		panic(err)
	}

	err = saveQuotation(quotation["bid"])
	if err != nil {
		log.Fatalf("Error saving quotation. %v", err)
		panic(err)
	}

	log.Println("Closing client!")
}
