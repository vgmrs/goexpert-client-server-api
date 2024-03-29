package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

type Quotation struct {
	ASK         float32   `json:"USDBRL.ask"`
	BID         float32   `json:"USDBRL.bid"`
	Code        string    `json:"USDBRL.code"`
	CodeIn      string    `json:"USDBRL.code_in"`
	CreatedDate time.Time `json:"USDBRL.create_date"`
	High        float32   `json:"USDBRL.high"`
	Low         float32   `json:"USDBRL.low"`
	Name        string    `json:"USDBRL.name"`
	PCTChange   float32   `json:"USDBRL.pct_change"`
	Timestamp   string    `json:"USDBRL.timestamp"`
	VarBID      float32   `json:"USDBRL.var_bid"`
}

func getQuotation() ([]byte, error) {
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

	return body, nil
}

func getDB() *sql.DB {
	dbOnce.Do(func() {
		conn, err := sql.Open("sqlite3", "database.db")
		if err != nil {
			log.Fatal("Error on open database connection:", err)
			panic(err)
		}

		db = conn
	})

	return db
}

func closeDB() {
	if db != nil {
		db.Close()

		db = nil
	}
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS quotation (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ask DOUBLE,
		bid DOUBLE,
		code TEXT,
		code_in TEXT,
		created_date DATE,
		high DOUBLE,
		low DOUBLE,
		name TEXT,
		pct_change DOUBLE,
		timestamp TEXT,
		var_bid DOUBLE
	)
	`)
	if err != nil {
		log.Println("Error on create table:", err)
		return err
	}

	return nil
}

func registerQuotation(q *Quotation) error {
	db := getDB()

	sql := `INSERT INTO quotation (ask, bid, code, code_in, created_date, high, low, name, pct_change, timestamp, var_bid) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	data := db.QueryRow(
		sql,
		q.ASK,
		q.BID,
		q.Code,
		q.CodeIn,
		q.CreatedDate,
		q.High,
		q.Low,
		q.Name,
		q.PCTChange,
		q.Timestamp,
		q.VarBID,
	)

	log.Println("Quotation saved successfully!", data)

	return nil
}

func handlerQuotation(w http.ResponseWriter, r *http.Request) {
	resp, err := getQuotation()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	quotation := Quotation{}

	err = json.Unmarshal(resp, &quotation)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	registerQuotation(&quotation)

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotation)
}

func init() {
	db := getDB()

	err := createTable(db)
	if err != nil {
		log.Fatal("Error creating table:", err)
		panic(err)
	}
}

func main() {
	r := chi.NewRouter()
	r.Get("/cotacao", handlerQuotation)

	log.Println("Start server...")

	if err := http.ListenAndServe(":8080", r); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server listen failed:", err)
		closeDB()
		panic(err)
	}
}
