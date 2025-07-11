export interface Stock {
  id: string
  name: string
  currentPrice: number
  amount: number
}

export interface Trader {
  id: string
  name: string
}

export interface Order {
  id: string
  traderId: string
  stockId: string
  type: 'buy' | 'sell'
  price: number
  quantity: number
  status: 'open' | 'filled' | 'cancelled'
  createdAt: Date
}
