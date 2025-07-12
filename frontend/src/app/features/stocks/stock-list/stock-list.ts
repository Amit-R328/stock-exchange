import {
  Component,
  OnInit,
  OnDestroy,
  ChangeDetectorRef,
  ChangeDetectionStrategy
} from '@angular/core'
import { CommonModule } from '@angular/common'
import { Router } from '@angular/router'
import { MatCardModule } from '@angular/material/card'
import { MatTableModule } from '@angular/material/table'
import { MatButtonModule } from '@angular/material/button'
import { MatIconModule } from '@angular/material/icon'
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner'
import { Subject, takeUntil } from 'rxjs'
import { Stock } from '../../../core/models/stock.model'
import { StockService } from '../../../core/services/stock/stock.service'
import { WebSocketService } from '../../../core/services/websocket/websocket.service'

@Component({
  selector: 'app-stock-list',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    CommonModule,
    MatCardModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule
  ],
  templateUrl: './stock-list.html',
  styleUrls: ['./stock-list.scss']
})
export class StockListComponent implements OnInit, OnDestroy {
  stocks: Stock[] = []
  loading = true
  error: string | null = null
  private previousPrices: Map<string, number> = new Map()
  private priceChanges: Map<string, number> = new Map()
  displayedColumns: string[] = ['name', 'currentPrice', 'amount', 'actions']
  private destroy$ = new Subject<void>()

  constructor (
    private stockService: StockService,
    private wsService: WebSocketService,
    private router: Router,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit (): void {
    // Detach from automatic change detection for full control
    this.cdr.detach()
    this.loadStocks()
    this.subscribeToUpdates()
  }

  ngOnDestroy (): void {
    this.destroy$.next()
    this.destroy$.complete()
    this.wsService.disconnect()
  }

  private loadStocks (): void {
    this.loading = true
    this.error = null

    this.stockService
      .getAllStocks()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: stocks => {
          this.stocks = stocks
          // Set initial previous prices to current prices so I start with 0% change
          this.updatePreviousPrices()
          console.log(
            '🔧 Initial previous prices set, map size:',
            this.previousPrices.size
          )
          this.loading = false
          // Trigger change detection after initial load
          this.cdr.detectChanges()
        },
        error: err => {
          this.error = 'Failed to load stocks. Please try again.'
          this.loading = false
          this.cdr.detectChanges()
        },
        complete: () => {
          // Ensure loading is always set to false when complete
          this.loading = false
          this.cdr.detectChanges()
        }
      })
  }

  private subscribeToUpdates (): void {
    this.wsService
      .connect()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: update => {
          console.log('🔄 WebSocket update received:', update)

          // Handle connection status messages
          if (update.type === 'connection') {
            if (update.status === 'connected') {
              console.log('🟢 WebSocket reconnected - refreshing data')
              // Refresh stocks data when reconnected
              this.loadStocks()
            } else if (update.status === 'disconnected') {
              console.log('🔴 WebSocket disconnected')
            }
            return // Exit early for connection status messages
          }

          if (update.type === 'stocks') {
            console.log(
              '📊 Before update - Current stocks count:',
              this.stocks.length
            )
            console.log(
              '📊 Previous prices map size:',
              this.previousPrices.size
            )

            // Store current prices as previous before updating
            const currentPrices = new Map<string, number>()
            this.stocks.forEach(stock => {
              currentPrices.set(stock.id, stock.currentPrice)
            })

            // Update stocks data
            const oldStocks = [...this.stocks]
            this.stocks = update.data

            // Check if there are any actual price changes
            let hasChanges = false
            console.log('🔍 Price comparisons:')
            this.priceChanges.clear()

            this.stocks.forEach(newStock => {
              const oldStock = oldStocks.find(s => s.id === newStock.id)
              const previousPrice = currentPrices.get(newStock.id)

              // Calculate and cache price change
              if (previousPrice && previousPrice !== newStock.currentPrice) {
                const change =
                  ((newStock.currentPrice - previousPrice) / previousPrice) *
                  100
                this.priceChanges.set(newStock.id, change)
                hasChanges = true
                console.log(
                  `💹 PRICE CHANGE for ${newStock.id}: $${previousPrice} -> $${
                    newStock.currentPrice
                  } = ${change.toFixed(2)}%`
                )
              } else {
                this.priceChanges.set(newStock.id, 0)
              }
            })

            // Now update previous prices with the stored current prices
            this.previousPrices = currentPrices

            // Only trigger change detection if there were actual changes
            if (hasChanges) {
              console.log('✅ Changes detected - triggering re-render')
              this.cdr.detectChanges()
            } else {
              console.log(
                '⏭️  No price changes detected - but checking for other updates'
              )
              // Even if no price changes, there might be other updates (new stocks, removed stocks)
              if (this.stocks.length !== oldStocks.length) {
                console.log('📊 Stock count changed, triggering re-render')
                this.cdr.detectChanges()
              }
            }

            console.log(
              '📊 After update - New stocks count:',
              this.stocks.length
            )
          }
        },
        error: error => {
          console.error('🔌 WebSocket connection error:', error)
          // Could add UI notification here for connection issues
        },
        complete: () => {
          console.log('🔌 WebSocket connection closed')
        }
      })
  }

  getPriceChange (stock: Stock): number {
    // Return cached price change to avoid recalculating multiple times
    const change = this.priceChanges.get(stock.id)
    if (change === undefined) {
      console.warn(
        `⚠️  Price change not found for ${stock.id}, returning 0. Map size: ${this.priceChanges.size}`
      )
      console.warn(
        '💾 Current priceChanges map keys:',
        Array.from(this.priceChanges.keys())
      )
      return 0
    }
    return change
  }

  private updatePreviousPrices (): void {
    this.priceChanges.clear()
    this.stocks.forEach(stock => {
      this.previousPrices.set(stock.id, stock.currentPrice)
      this.priceChanges.set(stock.id, 0) // Initial load = 0% change
      console.log(
        `🔧 Set previous price for ${stock.id}: $${stock.currentPrice}`
      )
    })
  }

  viewStock (stock: Stock): void {
    this.router.navigate(['/stocks', stock.id])
  }
}
