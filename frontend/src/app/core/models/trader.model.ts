import { Order } from "./order.model"
import { Transaction } from "./transaction.model"

export interface Trader {
  id: string
  name: string
  money: number
  holdings: { [stockId: string]: number }
}

export interface TraderInfo {
  id: string
  name: string
}

export interface TraderDetails extends Trader {
  openOrders: Order[] | null
}

export interface TraderTransactionsResponse {
  transactions: Transaction[]
  profitLoss: number
}
