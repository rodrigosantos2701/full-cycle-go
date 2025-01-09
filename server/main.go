package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type PriceType struct {
	Bid string `json:"bid"`
}

func fetchPrice() (*PriceType, error) {
	resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]PriceType
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	price, ok := result["USDBRL"]
	if !ok {
		return nil, fmt.Errorf("cotacao not found")
	}

	return &price, nil
}

func savePrice(ctx context.Context, cotacao *PriceType) error {

	db, err := sql.Open("sqlite3", "./price.db")
	if err != nil {
		log.Printf("Error trying open datebase")
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS dolarPrice (id INTEGER PRIMARY KEY, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "INSERT INTO dolarPrice (bid) VALUES (?)", cotacao.Bid)

	return err
}

func PriceHandler(w http.ResponseWriter, r *http.Request) {
	price, err := fetchPrice()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	savePrice(ctx, price)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}

func main() {
	http.HandleFunc("/price", PriceHandler)

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
