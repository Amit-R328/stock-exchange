package handlers

import (
	"log"
	"net/http"
	"stock-exchange/internal/models"
	"stock-exchange/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handlers struct {
	exchange         *services.Exchange
	upgrader         websocket.Upgrader
	algorithmManager *services.AlgorithmManager
}

func NewHandlers(exchange *services.Exchange, algorithmManager *services.AlgorithmManager) *Handlers {
	return &Handlers{
		exchange: exchange,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		algorithmManager: algorithmManager,
	}
}

// GetAllStocks
// @Summary Get all stocks
// @Description Get current data for all stocks
// @Tags stocks
// @Produce json
// @Success 200 {array} models.Stock
// @Router /stocks [get]
func (h *Handlers) GetAllStocks(c *gin.Context) {
	stocks := h.exchange.GetAllStocks()
	c.JSON(http.StatusOK, stocks)
}

// GetStock
// @Summary Get stock details
// @Description Get specific stock data including open orders and last 10 transactions
// @Tags stocks
// @Produce json
// @Param id path string true "Stock ID"
// @Success 200 {object} StockDetailsResponse
// @Router /stocks/{id} [get]
func (h *Handlers) GetStock(c *gin.Context) {
	stockID := c.Param("id")

	stock, exists := h.exchange.GetStock(stockID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	openOrders := h.exchange.GetOpenOrders(stockID)
	transactions := h.exchange.GetLastTransactions(stockID, 10)

	response := StockDetailsResponse{
		Stock:        stock,
		OpenOrders:   openOrders,
		Transactions: transactions,
	}

	c.JSON(http.StatusOK, response)
}

// PlaceOrder
// @Summary Place a new order
// @Description Place a buy or sell order
// @Tags trading
// @Accept json
// @Produce json
// @Param order body OrderRequest true "Order details"
// @Success 201 {object} models.Order
// @Router /orders [post]
func (h *Handlers) PlaceOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if trader has conflicting orders
	if h.exchange.HasConflictingOrder(req.TraderID, req.StockID, req.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot have both buy and sell orders for the same stock"})
		return
	}

	order := &models.Order{
		ID:        uuid.New().String(),
		TraderID:  req.TraderID,
		StockID:   req.StockID,
		Type:      req.Type,
		Price:     req.Price,
		Quantity:  req.Quantity,
		Status:    models.Open,
		CreatedAt: time.Now(),
	}

	if err := h.exchange.PlaceOrder(order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// CancelOrder
// @Summary Cancel an order
// @Description Cancel an existing order by ID
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id} [delete]
func (h *Handlers) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	if err := h.exchange.CancelOrder(orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Replace gin.H with proper struct
	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Order cancelled successfully",
	})
}

// GetAllTraders
// @Summary Get all traders
// @Description Get names of all traders
// @Tags traders
// @Produce json
// @Success 200 {array} TraderInfo
// @Router /traders [get]
func (h *Handlers) GetAllTraders(c *gin.Context) {
	traders := h.exchange.GetAllTraders()
	c.JSON(http.StatusOK, traders)
}

// GetTrader
// @Summary Get trader details
// @Description Get trader's open orders, holdings and cash
// @Tags traders
// @Produce json
// @Param id path string true "Trader ID"
// @Success 200 {object} TraderDetailsResponse
// @Router /traders/{id} [get]
func (h *Handlers) GetTrader(c *gin.Context) {
	traderID := c.Param("id")

	trader, exists := h.exchange.GetTrader(traderID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found"})
		return
	}

	openOrders := h.exchange.GetTraderOpenOrders(traderID)

	// Ensure I never send null for openOrders
	if openOrders == nil {
		openOrders = make([]models.Order, 0)
	}

	response := TraderDetailsResponse{
		Trader:     trader,
		OpenOrders: openOrders,
	}

	c.JSON(http.StatusOK, response)
}

// GetTraderTransactions
// @Summary Get trader transactions
// @Description Get last 8 transactions of a trader
// @Tags traders
// @Produce json
// @Param id path string true "Trader ID"
// @Success 200 {object} TraderTransactionsResponse
// @Router /traders/{id}/transactions [get]
func (h *Handlers) GetTraderTransactions(c *gin.Context) {
	traderID := c.Param("id")

	transactions := h.exchange.GetTraderTransactions(traderID, 8)
	profitLoss := h.exchange.CalculateProfitLoss(traderID)

	response := TraderTransactionsResponse{
		Transactions: transactions,
		ProfitLoss:   profitLoss,
	}

	c.JSON(http.StatusOK, response)
}

// GetStockHistory
// @Summary Get stock price history
// @Description Get historical price data for charts
// @Tags stocks
// @Produce json
// @Param id path string true "Stock ID"
// @Param days query int false "Number of days of history (default 30)"
// @Success 200 {object} StockHistoryResponse
// @Router /stocks/{id}/history [get]
func (h *Handlers) GetStockHistory(c *gin.Context) {
	stockID := c.Param("id")
	days := 30 // Default to 30 days

	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	history := h.exchange.GetStockHistory(stockID, days)

	response := StockHistoryResponse{
		StockID: stockID,
		Days:    days,
		History: history,
	}

	c.JSON(http.StatusOK, response)
}

// GetTraderPerformance
// @Summary Get trader performance history
// @Description Get trader performance data for charts
// @Tags traders
// @Produce json
// @Param id path string true "Trader ID"
// @Param days query int false "Number of days of history (default 30)"
// @Success 200 {object} TraderPerformanceResponse
// @Router /traders/{id}/performance [get]
func (h *Handlers) GetTraderPerformance(c *gin.Context) {
	traderID := c.Param("id")
	days := 30 // Default to 30 days

	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	performance := h.exchange.GetTraderPerformance(traderID, days)
	portfolio := h.exchange.GetTraderPortfolio(traderID)
	activity := h.exchange.GetTraderActivity(traderID, 6) // Last 6 months

	response := TraderPerformanceResponse{
		TraderID:    traderID,
		Days:        days,
		Performance: performance,
		Portfolio:   portfolio,
		Activity:    activity,
	}

	c.JSON(http.StatusOK, response)
}

// HandleWebSocket handles WebSocket connections for real-time updates
func (h *Handlers) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer func() {
		log.Println("WebSocket connection closed")
		conn.Close()
	}()

	log.Println("New WebSocket connection established")

	// Subscribe to updates
	subscription := h.exchange.Subscribe()
	defer h.exchange.Unsubscribe(subscription)

	// Handle incoming messages (ping/pong)
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}
		}
	}()

	// Send updates
	for {
		select {
		case update, ok := <-subscription.GetChannel():
			if !ok {
				log.Println("Subscription channel closed")
				return
			}

			if err := conn.WriteJSON(update); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
			log.Printf("Sent WebSocket update: %s", update.Type)
		}
	}
}

// ToggleAlgorithmicTrader
// @Summary Toggle algorithmic trader
// @Description Start or stop an algorithmic trader
// @Tags algorithms
// @Param id path string true "Trader ID"
// @Success 200 {object} SuccessResponse
// @Router /algorithms/{id}/toggle [post]
func (h *Handlers) ToggleAlgorithmicTrader(c *gin.Context) {
	traderID := c.Param("id")

	if h.algorithmManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Algorithm manager not available"})
		return
	}

	err := h.algorithmManager.ToggleTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Trader toggled successfully"})
}

// StartAlgorithmManager
// @Summary Start algorithm manager
// @Description Start the algorithmic trading system
// @Tags algorithms
// @Success 200 {object} SuccessResponse
// @Router /algorithms/start [post]
func (h *Handlers) StartAlgorithmManager(c *gin.Context) {
	if h.algorithmManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Algorithm manager not available"})
		return
	}

	h.algorithmManager.Start()
	c.JSON(http.StatusOK, SuccessResponse{Message: "Algorithm manager started"})
}

// StopAlgorithmManager
// @Summary Stop algorithm manager
// @Description Stop the algorithmic trading system
// @Tags algorithms
// @Success 200 {object} SuccessResponse
// @Router /algorithms/stop [post]
func (h *Handlers) StopAlgorithmManager(c *gin.Context) {
	if h.algorithmManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Algorithm manager not available"})
		return
	}

	h.algorithmManager.Stop()
	c.JSON(http.StatusOK, SuccessResponse{Message: "Algorithm manager stopped"})
}

// GetAlgorithms
// @Summary Get all algorithmic traders
// @Description Get list of all algorithmic trading bots
// @Tags algorithms
// @Produce json
// @Success 200 {array} AlgorithmicTraderResponse
// @Router /algorithms [get]
func (h *Handlers) GetAlgorithms(c *gin.Context) {
	algorithms := h.algorithmManager.GetAlgoTraders()

	var response []AlgorithmicTraderResponse
	for _, algo := range algorithms {
		response = append(response, AlgorithmicTraderResponse{
			ID:                algo.ID,
			Name:              algo.Name,
			Strategy:          algo.Strategy,
			Active:            algo.Active,
			OrdersPlaced:      algo.OrdersPlaced,
			ProfitLoss:        algo.ProfitLoss,
			LastAction:        algo.LastAction,
			MaxOrderValue:     algo.Config.MaxOrderValue,
			MinOrderValue:     algo.Config.MinOrderValue,
			RiskThreshold:     algo.Config.RiskThreshold,
			CooldownSeconds:   algo.Config.CooldownSeconds,
			MomentumThreshold: algo.Config.MomentumThreshold,
			ContrarianSpread:  algo.Config.ContrarianSpread,
		})
	}

	c.JSON(http.StatusOK, response)
}

// StartAlgorithm
// @Summary Start algorithmic trader
// @Description Start a specific algorithmic trading bot
// @Tags algorithms
// @Param id path string true "Algorithm ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Router /algorithms/{id}/start [post]
func (h *Handlers) StartAlgorithm(c *gin.Context) {
	id := c.Param("id")

	// Get the algorithm first to check if it exists and if it's not already active
	algorithms := h.algorithmManager.GetAlgoTraders()
	var targetAlgo *services.AlgoTrader
	for _, algo := range algorithms {
		if algo.ID == id {
			targetAlgo = algo
			break
		}
	}

	if targetAlgo == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Algorithm not found"})
		return
	}

	if targetAlgo.Active {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Algorithm is already active"})
		return
	}

	err := h.algorithmManager.ToggleTrader(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Algorithm started successfully"})
}

// StopAlgorithm
// @Summary Stop algorithmic trader
// @Description Stop a specific algorithmic trading bot
// @Tags algorithms
// @Param id path string true "Algorithm ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Router /algorithms/{id}/stop [post]
func (h *Handlers) StopAlgorithm(c *gin.Context) {
	id := c.Param("id")

	// Get the algorithm first to check if it exists and if it's active
	algorithms := h.algorithmManager.GetAlgoTraders()
	var targetAlgo *services.AlgoTrader
	for _, algo := range algorithms {
		if algo.ID == id {
			targetAlgo = algo
			break
		}
	}

	if targetAlgo == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Algorithm not found"})
		return
	}

	if !targetAlgo.Active {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Algorithm is already inactive"})
		return
	}

	err := h.algorithmManager.ToggleTrader(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Algorithm stopped successfully"})
}

// GetAlgorithmStatus
// @Summary Get algorithm status
// @Description Get detailed status of a specific algorithmic trading bot
// @Tags algorithms
// @Param id path string true "Algorithm ID"
// @Success 200 {object} AlgorithmicTraderResponse
// @Failure 404 {object} ErrorResponse
// @Router /algorithms/{id}/status [get]
func (h *Handlers) GetAlgorithmStatus(c *gin.Context) {
	id := c.Param("id")

	algorithms := h.algorithmManager.GetAlgoTraders()
	var targetAlgo *services.AlgoTrader
	for _, algo := range algorithms {
		if algo.ID == id {
			targetAlgo = algo
			break
		}
	}

	if targetAlgo == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Algorithm not found"})
		return
	}

	response := AlgorithmicTraderResponse{
		ID:                targetAlgo.ID,
		Name:              targetAlgo.Name,
		Strategy:          targetAlgo.Strategy,
		Active:            targetAlgo.Active,
		OrdersPlaced:      targetAlgo.OrdersPlaced,
		ProfitLoss:        targetAlgo.ProfitLoss,
		LastAction:        targetAlgo.LastAction,
		MaxOrderValue:     targetAlgo.Config.MaxOrderValue,
		MinOrderValue:     targetAlgo.Config.MinOrderValue,
		RiskThreshold:     targetAlgo.Config.RiskThreshold,
		CooldownSeconds:   targetAlgo.Config.CooldownSeconds,
		MomentumThreshold: targetAlgo.Config.MomentumThreshold,
		ContrarianSpread:  targetAlgo.Config.ContrarianSpread,
	}

	c.JSON(http.StatusOK, response)
}

// Request/Response types
type OrderRequest struct {
	TraderID string           `json:"traderId" binding:"required"`
	StockID  string           `json:"stockId" binding:"required"`
	Type     models.OrderType `json:"type" binding:"required"`
	Price    float64          `json:"price" binding:"required,gt=0"`
	Quantity int              `json:"quantity" binding:"required,gt=0"`
}

type StockDetailsResponse struct {
	*models.Stock
	OpenOrders   []models.Order       `json:"openOrders"`
	Transactions []models.Transaction `json:"lastTransactions"`
}

type TraderInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TraderDetailsResponse struct {
	*models.Trader
	OpenOrders []models.Order `json:"openOrders"`
}

type TraderTransactionsResponse struct {
	Transactions []models.Transaction `json:"transactions"`
	ProfitLoss   float64              `json:"profitLoss"`
}

type StockHistoryResponse struct {
	StockID string              `json:"stockId"`
	Days    int                 `json:"days"`
	History []models.PriceQuote `json:"history"`
}

type TraderPerformanceResponse struct {
	TraderID    string                   `json:"traderId"`
	Days        int                      `json:"days"`
	Performance []models.PerformanceData `json:"performance"`
	Portfolio   models.PortfolioData     `json:"portfolio"`
	Activity    []models.ActivityLog     `json:"activity"`
}

// AlgorithmicTraderResponse represents an algorithmic trader response
type AlgorithmicTraderResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Strategy     string    `json:"strategy"`
	Active       bool      `json:"active"`
	OrdersPlaced int       `json:"ordersPlaced"`
	ProfitLoss   float64   `json:"profitLoss"`
	LastAction   time.Time `json:"lastAction"`
	// Configuration parameters
	MaxOrderValue     float64 `json:"maxOrderValue"`     // Maximum value per order
	MinOrderValue     float64 `json:"minOrderValue"`     // Minimum value per order
	RiskThreshold     float64 `json:"riskThreshold"`     // Risk percentage of portfolio
	CooldownSeconds   int     `json:"cooldownSeconds"`   // Cooldown between orders
	MomentumThreshold float64 `json:"momentumThreshold"` // Momentum strategy threshold
	ContrarianSpread  float64 `json:"contrarianSpread"`  // Contrarian strategy spread
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// Replace gin.H usage with proper structs for Swagger
