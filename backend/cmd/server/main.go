package main

import (
	"log"
	"net/http"
	"stock-exchange/internal/handlers"
	"stock-exchange/internal/middleware"
	"stock-exchange/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "stock-exchange/docs" // This line is crucial
)

// @title Stock Exchange API
// @version 1.0
// @description A stock exchange simulation API
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Initialize exchange
	exchange := services.NewExchange()
	if err := exchange.LoadConfig("./config/config.json"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Start price updater
	priceUpdater := services.NewPriceUpdater(exchange, 10*time.Second)
	priceUpdater.Start()

	// Initialize and start algorithm manager
	algorithmManager := services.NewAlgorithmManager(exchange)

	// Register algorithmic traders in the exchange system
	for _, trader := range algorithmManager.GetAlgoTraders() {
		exchange.RegisterTrader(trader.ID, trader.Name, trader.InitialMoney)
		log.Printf("🤖 Registered algorithmic trader: %s with $%.2f", trader.Name, trader.InitialMoney)
	}

	algorithmManager.Start()

	// Setup router
	router := gin.Default()
	router.Use(middleware.CORS())

	// Initialize handlers
	h := handlers.NewHandlers(exchange, algorithmManager)

	// API routes
	api := router.Group("/api/v1")
	{
		// Stock endpoints
		api.GET("/stocks", h.GetAllStocks)
		api.GET("/stocks/:id", h.GetStock)
		api.GET("/stocks/:id/history", h.GetStockHistory)

		// Trading endpoints
		api.POST("/orders", h.PlaceOrder)
		api.DELETE("/orders/:id", h.CancelOrder)

		// Trader endpoints
		api.GET("/traders", h.GetAllTraders)
		api.GET("/traders/:id", h.GetTrader)
		api.GET("/traders/:id/transactions", h.GetTraderTransactions)
		api.GET("/traders/:id/performance", h.GetTraderPerformance)

		// Algorithmic trading endpoints
		api.GET("/algorithms", h.GetAlgorithms)
		api.POST("/algorithms/:id/start", h.StartAlgorithm)
		api.POST("/algorithms/:id/stop", h.StopAlgorithm)
		api.GET("/algorithms/:id/status", h.GetAlgorithmStatus)
	}

	// Swagger documentation - ONLY this line, remove the Static route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// REMOVE THIS LINE - it conflicts with Swagger:
	// router.Static("/swagger", "./docs")

	// WebSocket for real-time updates
	router.GET("/ws", h.HandleWebSocket)

	// Simple test endpoint
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API working"})
	})

	// Debug Swagger endpoint
	router.GET("/swagger-debug", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Swagger route accessible"})
	})

	log.Println("Server starting on :8080")
	log.Println("API Endpoints:")
	log.Println("- API: http://localhost:8080/api/v1/stocks")
	log.Println("- Swagger: http://localhost:8080/swagger/index.html")
	log.Println("- WebSocket: ws://localhost:8080/ws")
	log.Println("Test API: http://localhost:8080/api/v1/test")
	log.Println("Swagger Debug: http://localhost:8080/swagger-debug")
	log.Println("Swagger JSON: http://localhost:8080/swagger/swagger.json")
	router.Run(":8080")
}
