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

func fetchPrice(ctx context.Context) (*PriceType, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
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
		return nil, fmt.Errorf("price not found")
	}

	return &price, nil
}

func savePrice(ctx context.Context, price *PriceType) error {

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

	_, err = db.ExecContext(ctx, "INSERT INTO dolarPrice (bid) VALUES (?)", price.Bid)

	return err
}

func PriceHandler(w http.ResponseWriter, r *http.Request) {
	ctxAPI, cancelAPI := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancelAPI()

	price, err := fetchPrice(ctxAPI)
	if err != nil {
		log.Printf("Error fetching price: %v", err)
		http.Error(w, "Failed to fetch price", http.StatusInternalServerError)
		return
	}

	ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelDB()

	if err := savePrice(ctxDB, price); err != nil {
		log.Printf("Error saving price: %v", err)
		http.Error(w, "Failed to save price", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}

func main() {
	http.HandleFunc("/price", PriceHandler)

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
