# Stock Exchange Application

## Project Description
A complete stock exchange system built with Go (Backend) and Angular (Frontend) that enables real-time stock trading, graphical data visualization, and investment portfolio management.

## Technologies
### Backend (Go)
- **Gin Framework** - HTTP server
- **WebSocket** - Real-time updates
- **Swagger** - Automatic API documentation
- **CORS** - Cross-Origin Request support

### Frontend (Angular)
- **Angular 20** - Modern SPA development framework
- **Material Design** - Professional and clean design
- **Chart.js** - Interactive charts
- **TypeScript** - Structured and safe development
- **SCSS** - Advanced styling

## System Requirements
- **Node.js** (version 18 and above)
- **Go** (version 1.19 and above)
- **Angular CLI** (`npm install -g @angular/cli`)

## Running the Application

### Step 1: Starting the Backend (Go)

1. **Open terminal and navigate to the Backend directory:**
   ```bash
   cd stock-exchange-backend
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the server:**
   ```bash
   go run cmd/server/main.go
   ```

4. **Verify the server is running:**
   - Server will start on port 8080
   - Check: http://localhost:8080/api/v1/test
   - API Documentation: http://localhost:8080/swagger/index.html
   - **Algorithmic Bots**: Two trading bots start automatically!

### Step 2: Starting the Frontend (Angular)

1. **Open another terminal and navigate to the Frontend directory:**
   ```bash
   cd stock-exchange-frontend/stock-exchange-frontend
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Run the application:**
   ```bash
   ng serve
   ```

4. **Open browser:**
   - Go to: http://localhost:4200
   - The application will start loading

## Project Structure

```
stock-exchange-try/
├── stock-exchange-backend/          # Go Server
│   ├── cmd/server/main.go          # Server entry point
│   ├── internal/
│   │   ├── handlers/               # API controllers
│   │   ├── models/                 # Data models
│   │   ├── services/               # Business logic
│   │   └── middleware/             # Middleware (CORS, etc.)
│   ├── config/config.json          # Initial configuration
│   └── docs/                       # Swagger documentation
│
└── stock-exchange-frontend/        # Angular Application
    └── stock-exchange-frontend/
        ├── src/app/
        │   ├── core/               # Services and guards
        │   ├── features/           # Application pages
        │   ├── shared/             # Shared components
        │   └── styles/             # Global styling
        └── public/                 # Static files
```

## Key Features

### 📊 **Charts and Data**
- **Stock Charts**: Track historical prices
- **Trader Charts**: Performance, portfolio distribution, activity
- **Real-time Updates**: Prices update automatically

### 💰 **Stock Trading**
- **Buy and Sell Orders**: Simple and convenient interface
- **View Open Orders**: Manage existing orders
- **Transaction History**: Track all activities

### 👤 **Trader Management**
- **Personal Dashboard**: Portfolio overview
- **Trader List**: View all traders in the system
- **Performance Analysis**: Detailed performance charts

### 🤖 **Algorithmic Trading Bots**
- **Auto-Starting Bots**: Two trading algorithms start automatically
- **Real-time Strategy Execution**: Bots analyze market data and execute trades
- **Configurable Parameters**: Customize trading thresholds and behaviors
- **Performance Monitoring**: Track bot trading activity and profitability

## Application Entry Points

### API Endpoints (Backend)
- `GET /api/v1/stocks` - List all stocks
- `GET /api/v1/stocks/:id` - Specific stock details
- `GET /api/v1/stocks/:id/history` - Price history
- `GET /api/v1/traders` - List of traders
- `GET /api/v1/traders/:id` - Trader details
- `GET /api/v1/traders/:id/performance` - Trader performance
- `POST /api/v1/orders` - New order
- `DELETE /api/v1/orders/:id` - Cancel order
- `GET /api/v1/algorithms` - List algorithmic trading bots
- `POST /api/v1/algorithms/:id/start` - Start a trading bot
- `POST /api/v1/algorithms/:id/stop` - Stop a trading bot
- `GET /api/v1/algorithms/:id/status` - Get algorithm status
- `WebSocket: ws://localhost:8080/ws` - Real-time updates

### Frontend Routes
- `/` - Home page
- `/stocks` - Stock list
- `/stocks/:id` - Stock details
- `/traders` - Trader list
- `/personal` - Personal dashboard

## Troubleshooting Common Issues

### Backend not starting
1. Verify Go is installed: `go version`
2. Check that port 8080 is available
3. Ensure `config/config.json` file exists

### Frontend not starting
1. Verify Node.js is installed: `node --version`
2. Run `npm install` again
3. Clear cache: `npm cache clean --force`
4. Check that port 4200 is available

### CORS Issues
- Ensure Backend is running before Frontend
- Check CORS settings in `middleware/cors.go`

### WebSocket not connecting
- Ensure Backend is running
- Check browser Console for errors
- Try refreshing the page

## Further Development

## 🤖 Algorithmic Trading System

### Overview
The system includes two auto-starting trading bots that analyze market data and execute trades automatically:

1. **🚀 Momentum Hunter Bot** - Rides the wave of price increases
2. **📉 Contrarian Trader Bot** - Buys when others sell (contrarian strategy)

### Bot Configuration

Both bots start automatically when the server launches and can be configured via API.

#### Momentum Hunter 🚀
```json
{
  "id": "momentum_hunter",
  "name": "Momentum Hunter 🚀",
  "strategy": "momentum",
  "active": true,
  "config": {
    "momentum_threshold": 2.5,    // Minimum % price increase to trigger buy (actual: 0.025 = 2.5%)
    "sell_threshold": 1.0,        // % profit to trigger sell
    "max_position_size": 0.1,     // Max 10% of available money per trade
    "analysis_window": 10         // Number of price points to analyze
  }
}
```

**Strategy**: Buys stocks showing upward momentum (price increasing by 2.5% or more), sells when profit reaches 1%.

#### Contrarian Trader 📉
```json
{
  "id": "contrarian_trader", 
  "name": "Contrarian Trader 📉",
  "strategy": "contrarian",
  "active": true,
  "config": {
    "buy_spread_threshold": 0.7,  // Min % spread between high/low to buy (actual: 0.007 = 0.7%)
    "sell_threshold": 2.0,        // % profit to trigger sell
    "max_position_size": 0.15,    // Max 15% of available money per trade
    "analysis_window": 15         // Number of price points to analyze
  }
}
```

**Strategy**: Buys during price volatility (when spread > 0.7%), sells when profit reaches 2%.

### Managing Bots

#### View Bot Status
```bash
curl http://localhost:8080/api/v1/algorithms
```

#### Get Specific Bot Status
```bash
curl http://localhost:8080/api/v1/algorithms/momentum-bot-1/status
```

#### Start/Stop Bots
```bash
# Stop a bot
curl -X POST http://localhost:8080/api/v1/algorithms/momentum-bot-1/stop

# Start a bot  
curl -X POST http://localhost:8080/api/v1/algorithms/momentum-bot-1/start
```

**Note**: Bot configuration is currently set in the source code. To modify trading parameters, you would need to update the values in `backend/internal/services/algo_trader.go` and restart the server.
```

### Performance Tips

**For Conservative Trading:**
- Lower `momentum_threshold` to 1.5% (less aggressive)
- Lower `sell_threshold` to 0.5% (take profits quickly)
- Reduce `max_position_size` to 0.05 (smaller positions)

**For Aggressive Trading:**
- Increase `momentum_threshold` to 4.0% (wait for strong signals)
- Increase `sell_threshold` to 3.0% (let winners run)
- Increase `max_position_size` to 0.2 (bigger bets)

**Pro Tip**: The bots work best when there's actual price movement. Want to see some action? Play with the thresholds! Set `momentum_threshold` to 0.5% and `max_position_size` to 0.8 to watch them go YOLO on all their money 💸, or crank up `momentum_threshold` to 10% and `max_position_size` to 0.01 to make them as careful as your grandmother investing her pension (or me)

### Adding New Stocks
Edit `config/config.json` file:
```json
{
  "shares": [
    {
      "id": "AAPL",
      "name": "Apple Inc.",
      "currentPrice": 150.0,
      "amount": 1000
    }
  ]
}
```

### Adding New Traders
```json
{
  "traders": [
    {
      "id": "trader1",
      "name": "John Doe",
      "money": 100000,
      "holdings": {}
    }
  ]
}
```

## Support and Help

- **API Documentation**: http://localhost:8080/swagger/index.html
- **Algorithmic Trading**: Bots start automatically - watch them trade!
- **Logs**: Error messages appear in terminal

---

**Note**: This is a demo application for educational purposes. Some historical data is simulated to demonstrate technical capabilities.
