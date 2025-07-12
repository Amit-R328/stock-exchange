export type OrderType = 'buy' | 'sell'
export type OrderStatus = 'open' | 'filled' | 'cancelled'

export interface Order {
  id: string
  traderId: string
  stockId: string
  type: OrderType
  price: number
  quantity: number
  status: OrderStatus
  createdAt: Date
}

export interface CreateOrderRequest {
  traderId: string
  stockId: string
  type: OrderType
  price: number
  quantity: number
}
