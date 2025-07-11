import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { ApiService } from '../../services/api.service'

@Component({
  selector: 'app-stock-list',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="stock-list">
      <h2>Stocks</h2>

      <div *ngIf="loading">Loading...</div>

      <div *ngIf="!loading" class="stocks">
        <div *ngFor="let stock of stocks" class="stock-item">
          <h3>{{ stock.name }}</h3>
          <p>Price: \${{ stock.currentPrice }}</p>
          <p>Available: {{ stock.amount }}</p>
          <button (click)="viewStock(stock)">View Details</button>
        </div>
      </div>
    </div>
  `,
  styles: [
    `
      .stock-list {
        padding: 20px;
      }
      .stocks {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
        gap: 20px;
      }
      .stock-item {
        border: 1px solid #ddd;
        padding: 15px;
        border-radius: 4px;
      }
    `
  ]
})
export class StockListComponent implements OnInit {
  stocks: any[] = []
  loading = true

  constructor (private apiService: ApiService) {}

  ngOnInit () {
    this.loadStocks()
  }

  loadStocks () {
    this.apiService.getStocks().subscribe({
      next: stocks => {
        this.stocks = stocks
        this.loading = false
      },
      error: err => {
        console.error('Failed to load stocks:', err)
        this.loading = false
      }
    })
  }

  viewStock (stock: any) {
    console.log('View stock:', stock)
    // TODO: Navigate to detail page
  }
}
