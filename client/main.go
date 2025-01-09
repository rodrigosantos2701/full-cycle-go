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

type PriceType struct {
	Bid string `json:"bid"`
}

func fetchPrice() (*PriceType, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/price", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var price PriceType
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return nil, err
	}
	return &price, nil
}

func savePrice(price *PriceType) error {
	content := fmt.Sprintf("Dolar: %s", price.Bid)
	return os.WriteFile("dolarPrice.txt", []byte(content), 0666)
}

func main() {
	price, err := fetchPrice()
	if err != nil {
		log.Printf("Error trying fetch")
	}

	err = savePrice(price)
	if err != nil {
		log.Println("Error trying save")
	}
}
