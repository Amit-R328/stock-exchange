import { Injectable } from '@angular/core'
import { HttpClient } from '@angular/common/http'
import { Observable } from 'rxjs'
import { Stock, StockDetails } from '../../models/stock.model'
import { Order } from '../../models/order.model'
import { Transaction } from '../../models/transaction.model'
import { environment } from '../../../../environments/environment'

@Injectable({
  providedIn: 'root'
})
export class StockService {
  private apiUrl = `${environment.apiUrl}/stocks`

  constructor (private http: HttpClient) {}

  getAllStocks (): Observable<Stock[]> {
    return this.http.get<Stock[]>(this.apiUrl)
  }

  getStock (id: string): Observable<StockDetails> {
    return this.http.get<StockDetails>(`${this.apiUrl}/${id}`)
  }

  getStockHistory (
    id: string,
    days: number = 30
  ): Observable<StockHistoryResponse> {
    return this.http.get<StockHistoryResponse>(
      `${this.apiUrl}/${id}/history?days=${days}`
    )
  }
}

export interface PriceQuote {
  timestamp: string
  price: number
  volume: number
}

export interface StockHistoryResponse {
  stockId: string
  days: number
  history: PriceQuote[]
}
