export interface Transaction {
  id: string
  buyerId: string
  sellerId: string
  stockId: string
  price: number
  quantity: number
  executedAt: Date
}
