import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-stock-list',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatProgressSpinnerModule
  ],
  templateUrl: './stock-list.html',
  styles: [`
    .stock-list-container {
      padding: 20px;
      max-width: 1200px;
      margin: 0 auto;
    }
    
    .loading {
      display: flex;
      justify-content: center;
      padding: 40px;
    }
    
    .error {
      color: red;
      text-align: center;
      padding: 20px;
    }
    
    .stocks-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
      gap: 20px;
      margin-top: 20px;
    }
    
    .stock-card {
      cursor: pointer;
      transition: transform 0.2s;
    }
    
    .stock-card:hover {
      transform: translateY(-4px);
    }
    
    .price {
      font-size: 24px;
      font-weight: bold;
      margin: 10px 0;
    }
    
    .shares {
      color: #666;
    }
  `]
})
export class StockListComponent implements OnInit {
  stocks: any[] = [];
  loading = true;
  error: string | null = null;

  constructor(
    private apiService: ApiService,
    private router: Router
  ) {}

  ngOnInit() {
    this.loadStocks();
  }

  loadStocks() {
    this.loading = true;
    this.error = null;
    
    this.apiService.getStocks().subscribe({
      next: (stocks) => {
        this.stocks = stocks;
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Failed to load stocks. Please try again.';
        this.loading = false;
      }
    });
  }

  viewStock(stock: any) {
    this.router.navigate(['/stocks', stock.id]);
  }
}