package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Amit-R328/stock-exchange/internal/handlers"
	"github.com/Amit-R328/stock-exchange/internal/middleware"
	"github.com/Amit-R328/stock-exchange/internal/services"
)

func main() {
	// Initialize exchange
	exchange := services.NewExchange()

	// Load configuration
	if err := exchange.LoadConfig("config.json"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Start price updater
	priceUpdater := services.NewPriceUpdater(exchange, 10*time.Second)
	priceUpdater.Start()
	defer priceUpdater.Stop()

	// Initialize handlers
	h := handlers.NewHandlers(exchange)

	// Setup routes with middleware
	setupRoutes(h)

	log.Println("🚀 Stock Exchange Server starting on :8080")
	log.Println("📊 Available endpoints:")
	log.Println("   GET    /api/v1/stocks")
	log.Println("   GET    /api/v1/stocks/{id}")
	log.Println("   POST   /api/v1/orders")
	log.Println("   DELETE /api/v1/orders/{id}")
	log.Println("   GET    /api/v1/traders")
	log.Println("   GET    /api/v1/traders/{id}")
	log.Println("   GET    /api/v1/traders/{id}/transactions")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func setupRoutes(h *handlers.Handlers) {
	// Stock endpoints
	http.HandleFunc("/api/v1/stocks", middleware.CORS(h.GetAllStocks))
	http.HandleFunc("/api/v1/stocks/", middleware.CORS(h.GetStock))

	// Order endpoints
	http.HandleFunc("/api/v1/orders", middleware.CORS(h.PlaceOrder))
	http.HandleFunc("/api/v1/orders/", middleware.CORS(h.CancelOrder))

	// Trader endpoints
	http.HandleFunc("/api/v1/traders", middleware.CORS(h.GetAllTraders))
	http.HandleFunc("/api/v1/traders/", middleware.CORS(h.GetTrader))

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
