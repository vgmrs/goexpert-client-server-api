package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("[SERVER]")

	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("[SERVER] Error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[SERVER] Error:", err)
		return
	}

	var result map[string]interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("[SERVER] Error:", err)
		return
	}

	fmt.Println(result)
}
