package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func getQuotation(w http.ResponseWriter, r *http.Request) {
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	res, err := http.Get(url)
	if err != nil {
		log.Println("Error getting quote:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("Error formatting response:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	r := chi.NewRouter()
	r.Get("/cotacao", getQuotation)

	log.Println("Start server...")
	http.ListenAndServe(":8080", r)
}
