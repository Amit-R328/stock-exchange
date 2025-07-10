package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Stock struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CurrentPrice float64 `json:"currentPrice"`
	Amount       int     `json:"amount"`
}

type Trader struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Money    float64        `json:"money"`
	Holdings map[string]int `json:"holdings"`
}

type Config struct {
	Shares  []Stock  `json:"shares"`
	Traders []Trader `json:"traders"`
}

type Exchange struct {
	Stocks  map[string]*Stock
	Traders map[string]*Trader
}

var exchange *Exchange

func main() {
	// Initialize our exchange
	exchange = &Exchange{
		Stocks:  make(map[string]*Stock),
		Traders: make(map[string]*Trader),
	}

	// Load configuration
	if err := loadConfig("../../config.json"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Let's start with a simple test endpoint
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Stock Exchange API is running! We have %d stocks and %d traders",
			len(exchange.Stocks), len(exchange.Traders))
	})

	// Add a simple endpoint to see our stocks
	http.HandleFunc("/stocks", handleGetStocks)

	log.Println("Server starting on :8080")
	log.Printf("Loaded %d stocks and %d traders", len(exchange.Stocks), len(exchange.Traders))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return err
	}

	for _, stock := range config.Shares {
		s := stock // Important: create a copy
		exchange.Stocks[s.ID] = &s
		log.Printf("Loaded stock: %s - %s at $%.2f", s.ID, s.Name, s.CurrentPrice)
	}

	// Load traders
	for _, trader := range config.Traders {
		t := trader                       // Important: create a copy
		t.Holdings = make(map[string]int) // Initialize empty holdings
		exchange.Traders[t.ID] = &t
		log.Printf("Loaded trader: %s - %s with $%.2f", t.ID, t.Name, t.Money)
	}

	return nil
}

func handleGetStocks(w http.ResponseWriter, r *http.Request) {
	// Simple handler to return all stocks as JSON
	w.Header().Set("Content-Type", "application/json")

	// Convert map to slice for easier JSON encoding
	stocks := make([]Stock, 0, len(exchange.Stocks))
	for _, stock := range exchange.Stocks {
		stocks = append(stocks, *stock)
	}

	if err := json.NewEncoder(w).Encode(stocks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
