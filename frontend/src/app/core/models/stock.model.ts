import { Order } from "./order.model"
import { Transaction } from "./transaction.model"

export interface Stock {
  id: string
  name: string
  currentPrice: number
  amount: number
}

export interface StockDetails {
  id: string
  name: string
  currentPrice: number
  amount: number
  openOrders: Order[]
  lastTransactions: Transaction[]
}
