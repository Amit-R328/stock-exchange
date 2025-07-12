import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { FormsModule } from '@angular/forms'
import { ActivatedRoute, RouterLink } from '@angular/router'
import { MatCardModule } from '@angular/material/card'
import { MatButtonModule } from '@angular/material/button'
import { MatFormFieldModule } from '@angular/material/form-field'
import { MatInputModule } from '@angular/material/input'
import { MatRadioModule } from '@angular/material/radio'
import { MatIconModule } from '@angular/material/icon'
import { MatDividerModule } from '@angular/material/divider'
import { MatChipsModule } from '@angular/material/chips'
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner'
import { StockDetails } from '../../../core/models/stock.model'
import { CreateOrderRequest, OrderType } from '../../../core/models/order.model'
import { StockService } from '../../../core/services/stock/stock.service'
import { OrderService } from '../../../core/services/order/order.service'
import { AuthService } from '../../../core/services/auth/auth.service'
import { TraderService } from '../../../core/services/trader/trader.service'
import { StockChartComponent } from '../../../shared/components/stock-chart/stock-chart.component'

@Component({
  selector: 'app-stock-detail',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    MatCardModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatRadioModule,
    MatIconModule,
    MatDividerModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    StockChartComponent
  ],
  templateUrl: './stock-detail.html',
  styleUrls: ['./stock-detail.scss']
})
export class StockDetailComponent implements OnInit {
  stock?: StockDetails
  loading = true
  orderForm = {
    type: 'buy' as OrderType,
    price: 0,
    quantity: 1
  }
  trader?: any
  error = ''

  constructor (
    private route: ActivatedRoute,
    private stockService: StockService,
    private orderService: OrderService,
    private authService: AuthService,
    private traderService: TraderService
  ) {}

  ngOnInit (): void {
    const stockId = this.route.snapshot.paramMap.get('id')
    if (stockId) {
      this.loadStock(stockId)
      this.loadTraderInfo()
    }
  }

  private loadStock (id: string): void {
    this.loading = true
    this.stockService.getStock(id).subscribe({
      next: stock => {
        this.stock = stock
        this.orderForm.price = Math.round(stock.currentPrice * 100) / 100
        this.loading = false
      },
      error: err => {
        this.error = 'Failed to load stock details'
        this.loading = false
        console.error('Error loading stock:', err)
      }
    })
  }

  private loadTraderInfo (): void {
    const traderId = this.authService.getCurrentTraderId()
    if (traderId) {
      this.traderService.getTrader(traderId).subscribe(trader => {
        this.trader = trader
      })
    }
  }

  placeOrder (): void {
    if (!this.stock || !this.trader) return

    const order: CreateOrderRequest = {
      traderId: this.trader.id,
      stockId: this.stock.id,
      type: this.orderForm.type,
      price: this.orderForm.price,
      quantity: this.orderForm.quantity
    }

    this.orderService.placeOrder(order).subscribe({
      next: () => {
        this.loadStock(this.stock!.id)
        this.loadTraderInfo()
        this.error = ''
      },
      error: err => {
        this.error = err.error.error || 'Failed to place order'
      }
    })
  }

  cancelOrder (orderId: string): void {
    this.orderService.cancelOrder(orderId).subscribe(() => {
      this.loadStock(this.stock!.id)
      this.loadTraderInfo()
      this.clearError() // Clear error when order is cancelled
    })
  }

  clearError (): void {
    this.error = ''
  }

  onPriceChange (): void {
    // Round to 2 decimal places
    this.orderForm.price = Math.round(this.orderForm.price * 100) / 100
    this.clearError()
  }

  canCancelOrder (order: any): boolean {
    return order.traderId === this.trader?.id && order.status === 'open'
  }
}
