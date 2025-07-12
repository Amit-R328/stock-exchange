import { Component, OnInit, OnDestroy } from '@angular/core'
import { CommonModule } from '@angular/common'
import { MatCardModule } from '@angular/material/card'
import { MatIconModule } from '@angular/material/icon'
import { MatButtonModule } from '@angular/material/button'
import { MatDividerModule } from '@angular/material/divider'
import { MatChipsModule } from '@angular/material/chips'
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner'
import {
  TraderDetails,
  TraderTransactionsResponse
} from '../../core/models/trader.model'
import { Transaction } from '../../core/models/transaction.model'
import { Stock } from '../../core/models/stock.model'
import { AuthService } from '../../core/services/auth/auth.service'
import { TraderService } from '../../core/services/trader/trader.service'
import { StockService } from '../../core/services/stock/stock.service'
import { OrderService } from '../../core/services/order/order.service'
import { WebSocketService } from '../../core/services/websocket/websocket.service'
import { forkJoin, Subscription } from 'rxjs'
import { TraderChartComponent } from '../../shared/components/trader-chart/trader-chart.component'

interface HoldingDisplay {
  stockId: string
  stockName: string
  quantity: number
  currentValue: number
  purchasePrice: number // Average purchase price for P&L calculation
}

@Component({
  selector: 'app-personal',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatIconModule,
    MatButtonModule,
    MatDividerModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    TraderChartComponent
  ],
  templateUrl: './personal.html',
  styleUrls: ['./personal.scss']
})
export class PersonalComponent implements OnInit, OnDestroy {
  traderId: string | null = null
  traderDetails?: TraderDetails
  loading = true
  transactions: Transaction[] = []
  profitLoss = 0
  holdings: HoldingDisplay[] = []
  stocks: Stock[] = [] // Add stocks array

  private wsSubscription?: Subscription
  private stockPrices = new Map<string, number>() // Cache for current stock prices
  private baseProfitLoss = 0 // P&L from completed transactions

  // Helper method to get stock name from stock ID
  getStockName (stockId: string): string {
    const stock = this.stocks.find(s => s.id === stockId)
    return stock ? stock.name : `Stock ${stockId}`
  }

  constructor (
    private authService: AuthService,
    private traderService: TraderService,
    private stockService: StockService,
    private orderService: OrderService,
    private webSocketService: WebSocketService
  ) {}

  ngOnInit (): void {
    this.traderId = this.authService.getCurrentTraderId()
    if (this.traderId) {
      this.loadTraderData()
      this.setupWebSocketConnection()
    }
  }

  ngOnDestroy (): void {
    if (this.wsSubscription) {
      this.wsSubscription.unsubscribe()
    }
    this.webSocketService.disconnect()
  }

  private setupWebSocketConnection (): void {
    this.wsSubscription = this.webSocketService.connect().subscribe({
      next: message => {
        this.handleWebSocketMessage(message)
      },
      error: error => {
        console.error('WebSocket error:', error)
      }
    })
  }

  private handleWebSocketMessage (message: any): void {
    if (message.type === 'stocks') {
      // Update stock prices from the WebSocket data
      const stocks = message.data as Stock[]
      this.stocks = stocks // Update stocks array for name lookups
      stocks.forEach(stock => {
        this.stockPrices.set(stock.id, stock.currentPrice)
      })
      this.updateHoldingsAndPL()
    }
  }

  private updateHoldingsAndPL (): void {
    // Update holdings with new prices
    this.holdings.forEach(holding => {
      const currentPrice = this.stockPrices.get(holding.stockId)
      if (currentPrice !== undefined) {
        holding.currentValue = holding.quantity * currentPrice
      }
    })

    // Recalculate P&L: base P&L + unrealized gains/losses from holdings
    let unrealizedPL = 0
    this.holdings.forEach(holding => {
      const currentPrice = this.stockPrices.get(holding.stockId)
      if (currentPrice !== undefined) {
        unrealizedPL +=
          holding.quantity * (currentPrice - holding.purchasePrice)
      }
    })

    this.profitLoss = this.baseProfitLoss + unrealizedPL
  }

  private loadTraderData (): void {
    if (!this.traderId) return

    this.loading = true
    forkJoin({
      details: this.traderService.getTrader(this.traderId),
      transactions: this.traderService.getTraderTransactions(this.traderId),
      stocks: this.stockService.getAllStocks()
    }).subscribe({
      next: ({ details, transactions, stocks }) => {
        this.traderDetails = details
        this.transactions = transactions.transactions
        this.baseProfitLoss = transactions.profitLoss
        this.profitLoss = transactions.profitLoss
        this.stocks = stocks // Store stocks in component property

        // Cache current stock prices
        stocks.forEach(stock => {
          this.stockPrices.set(stock.id, stock.currentPrice)
        })

        this.processHoldings(details, stocks)
        this.loading = false
      },
      error: err => {
        console.error('Error loading trader data:', err)
        this.loading = false
      }
    })
  }

  private processHoldings (trader: TraderDetails, stocks: Stock[]): void {
    const stockMap = new Map(stocks.map(s => [s.id, s]))
    this.holdings = []

    Object.entries(trader.holdings).forEach(([stockId, quantity]) => {
      if (quantity > 0) {
        const stock = stockMap.get(stockId)
        if (stock) {
          const avgPurchasePrice = this.calculateAveragePurchasePrice(
            stockId,
            quantity
          )
          this.holdings.push({
            stockId,
            stockName: stock.name,
            quantity,
            currentValue: quantity * stock.currentPrice,
            purchasePrice: avgPurchasePrice
          })
        }
      }
    })
  }

  private calculateAveragePurchasePrice (
    stockId: string,
    currentQuantity: number
  ): number {
    // Calculate average purchase price from transactions where this trader was the buyer
    const stockTransactions = this.transactions.filter(
      t => t.stockId === stockId && t.buyerId === this.traderId
    )

    if (stockTransactions.length === 0) {
      // Fallback to current price if no transactions found
      return this.stockPrices.get(stockId) || 0
    }

    let totalCost = 0
    let totalQuantity = 0

    stockTransactions.forEach(transaction => {
      totalCost += transaction.quantity * transaction.price
      totalQuantity += transaction.quantity
    })

    return totalQuantity > 0 ? totalCost / totalQuantity : 0
  }

  cancelOrder (orderId: string): void {
    this.orderService.cancelOrder(orderId).subscribe(() => {
      this.loadTraderData()
    })
  }

  getOpenOrders () {
    return this.traderDetails?.openOrders || []
  }

  getHoldingProfitLoss (holding: HoldingDisplay): number {
    const currentPrice = this.stockPrices.get(holding.stockId) || 0
    return holding.quantity * (currentPrice - holding.purchasePrice)
  }
}
