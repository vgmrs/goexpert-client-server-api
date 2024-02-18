package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"
)

func getQuotation() (map[string]interface{}, error) {
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	res, err := http.Get(url)
	if err != nil {
		log.Println("Error getting quote:", err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return nil, err
	}

	var result map[string]interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("Error formatting response:", err)
		return nil, err
	}

	return result, nil
}

func getDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Println("Error on open database connection:", err)
		return nil, err
	}
	defer db.Close()

	return db, err
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS quotation (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		quotation TEXT
	)
	`)
	if err != nil {
		log.Println("Error on create table:", err)
		return err
	}

	return nil
}

func registerQuotation(quotation map[string]interface{}) error {
	db, err := getDB()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO quotation (quotation) VALUES (?)", quotation)
	if err != nil {
		log.Println("Error when inserting data into table:", err)
		return err
	}

	log.Println("Data saved successfully!")

	return nil
}

func handlerQuotation(w http.ResponseWriter, r *http.Request) {
	quotation, err := getQuotation()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	registerQuotation(quotation)

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotation)
}

func init() {
	db, err := getDB()
	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}

	err = createTable(db)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}
}

func main() {
	r := chi.NewRouter()
	r.Get("/cotacao", handlerQuotation)

	log.Println("Start server...")
	http.ListenAndServe(":8080", r)
}
