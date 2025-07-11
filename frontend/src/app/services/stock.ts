import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Stock } from '../models/stock.model';

@Injectable({
  providedIn: 'root'
})
export class StockService {
  private apiUrl = 'http://localhost:8080/api/v1/stocks';

  constructor(private http: HttpClient) {}

  getAllStocks(): Observable<Stock[]> {
    return this.http.get<Stock[]>(this.apiUrl);
  }

  getStock(id: string): Observable<Stock> {
    return this.http.get<Stock>(`${this.apiUrl}/${id}`);
  }
}
