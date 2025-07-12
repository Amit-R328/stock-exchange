import { Injectable } from '@angular/core'
import { HttpClient } from '@angular/common/http'
import { Observable } from 'rxjs'
import {
  TraderInfo,
  TraderDetails,
  TraderTransactionsResponse
} from '../../models/trader.model'
import { environment } from '../../../../environments/environment'

@Injectable({
  providedIn: 'root'
})
export class TraderService {
  private apiUrl = `${environment.apiUrl}/traders`

  constructor (private http: HttpClient) {}

  getAllTraders (): Observable<TraderInfo[]> {
    return this.http.get<TraderInfo[]>(this.apiUrl)
  }

  getTrader (id: string): Observable<TraderDetails> {
    return this.http.get<TraderDetails>(`${this.apiUrl}/${id}`)
  }

  getTraderTransactions (id: string): Observable<TraderTransactionsResponse> {
    return this.http.get<TraderTransactionsResponse>(
      `${this.apiUrl}/${id}/transactions`
    )
  }

  getTraderPerformance(id: string, days: number = 30): Observable<TraderPerformanceResponse> {
    return this.http.get<TraderPerformanceResponse>(
      `${this.apiUrl}/${id}/performance?days=${days}`
    )
  }
}

export interface PerformanceData {
  date: string;
  portfolioValue: number;
  profitLoss: number;
  cashBalance: number;
}

export interface PortfolioHolding {
  stockId: string;
  stockName: string;
  quantity: number;
  value: number;
  percentage: number;
}

export interface PortfolioData {
  holdings: PortfolioHolding[];
  totalValue: number;
  cashBalance: number;
}

export interface ActivityLog {
  period: string;
  buyOrders: number;
  sellOrders: number;
  volume: number;
  value: number;
}

export interface TraderPerformanceResponse {
  traderId: string;
  days: number;
  performance: PerformanceData[];
  portfolio: PortfolioData;
  activity: ActivityLog[];
}
