import { Injectable } from '@angular/core'
import { HttpClient } from '@angular/common/http'
import { Observable } from 'rxjs'
import { Order, CreateOrderRequest } from '../../models/order.model'
import { environment } from '../../../../environments/environment'

@Injectable({
  providedIn: 'root'
})
export class OrderService {
  private apiUrl = `${environment.apiUrl}/orders`

  constructor (private http: HttpClient) {}

  placeOrder (order: CreateOrderRequest): Observable<Order> {
    return this.http.post<Order>(this.apiUrl, order)
  }

  cancelOrder (orderId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${orderId}`)
  }
}
