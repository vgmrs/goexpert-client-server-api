package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	serverPort       string        = ":8080"
	quotationURL     string        = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	quotationTimeout time.Duration = 200 * time.Millisecond
	databaseName     string        = "database.db"
	databaseTimeout  time.Duration = 10 * time.Millisecond
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

type Quotation struct {
	ASK       string `json:"ask"`
	BID       string `json:"bid"`
	Code      string `json:"code"`
	CodeIn    string `json:"codein"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Name      string `json:"name"`
	PCTChange string `json:"pctChange"`
	VarBID    string `json:"varBid"`
}

func getQuotation(ctx context.Context) (*Quotation, error) {
	ctx, cancel := context.WithTimeout(ctx, quotationTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", quotationURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data map[string]Quotation
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	quotation, ok := data["USDBRL"]
	if !ok {
		return nil, errors.New("no quotation found for USD-BRL")
	}

	return &quotation, nil
}

func getDB() *sql.DB {
	dbOnce.Do(func() {
		log.Println("Starting database")

		conn, err := sql.Open("sqlite3", databaseName)
		if err != nil {
			log.Fatalf("Error on open database connection. %v", err)
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

	log.Println("Closing database")
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS quotation (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ask DOUBLE,
		bid DOUBLE,
		code VARCHAR(256),
		code_in VARCHAR(256),
		high DOUBLE,
		low DOUBLE,
		name VARCHAR(256),
		pct_change DOUBLE,
		var_bid DOUBLE
	)
	`)
	if err != nil {
		return err
	}

	return nil
}

func registerQuotation(ctx context.Context, q *Quotation) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	db := getDB()
	sql := `INSERT INTO quotation (ask, bid, code, code_in, high, low, name, pct_change, var_bid) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.ExecContext(ctx,
		sql,
		q.ASK,
		q.BID,
		q.Code,
		q.CodeIn,
		q.High,
		q.Low,
		q.Name,
		q.PCTChange,
		q.VarBID,
	)
	if err != nil {
		return err
	}

	log.Println("Quotation saved successfully!")
	return nil
}

func handlerQuotation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	quotation, err := getQuotation(ctx)
	if err != nil {
		log.Printf("Error getting quotation. %v", err)
		http.Error(w, fmt.Sprintf("Error getting quotation. %v", err), http.StatusInternalServerError)
		return
	}

	err = registerQuotation(ctx, quotation)
	if err != nil {
		log.Printf("Error registering quotation. %v", err)
		http.Error(w, fmt.Sprintf("Error registering quotation. %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(quotation); err != nil {
		log.Printf("Error encoding JSON response. %v", err)
		http.Error(w, fmt.Sprintf("Error encoding JSON response. %v", err), http.StatusInternalServerError)
		return
	}
}

func init() {
	log.Println("Starting sever...")

	db := getDB()
	err := createTable(db)
	if err != nil {
		log.Fatalf("Error creating table. %v", err)
		panic(err)
	}
}

func main() {
	defer closeDB()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", handlerQuotation)

	log.Printf("Running server on port %s", serverPort)

	if err := http.ListenAndServe(serverPort, mux); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server listen failed. %v", err)
		panic(err)
	}
}
