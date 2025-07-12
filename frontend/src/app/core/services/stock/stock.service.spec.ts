import { TestBed } from '@angular/core/testing'
import {
  HttpClientTestingModule,
  HttpTestingController
} from '@angular/common/http/testing'
import { StockService } from './stock.service'
import { Stock, StockDetails } from '../../models/stock.model'
import { environment } from '../../../../environments/environment'

describe('StockService', () => {
  let service: StockService
  let httpMock: HttpTestingController

  const mockStocks: Stock[] = [
    {
      id: '1',
      name: 'Apple Inc.',
      currentPrice: 150.0,
      amount: 1000
    },
    {
      id: '2',
      name: 'Microsoft Corp.',
      currentPrice: 300.0,
      amount: 500
    }
  ]

  const mockStockDetails: StockDetails = {
    id: '1',
    name: 'Apple Inc.',
    currentPrice: 150.0,
    amount: 1000,
    openOrders: [],
    lastTransactions: []
  }

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [StockService]
    })
    service = TestBed.inject(StockService)
    httpMock = TestBed.inject(HttpTestingController)
  })

  afterEach(() => {
    httpMock.verify()
  })

  it('should be created', () => {
    expect(service).toBeTruthy()
  })

  describe('getAllStocks', () => {
    it('should return an array of stocks', () => {
      service.getAllStocks().subscribe(stocks => {
        expect(stocks).toEqual(mockStocks)
        expect(stocks.length).toBe(2)
      })

      const req = httpMock.expectOne(`${environment.apiUrl}/stocks`)
      expect(req.request.method).toBe('GET')
      req.flush(mockStocks)
    })

    it('should handle empty stocks array', () => {
      service.getAllStocks().subscribe(stocks => {
        expect(stocks).toEqual([])
        expect(stocks.length).toBe(0)
      })

      const req = httpMock.expectOne(`${environment.apiUrl}/stocks`)
      expect(req.request.method).toBe('GET')
      req.flush([])
    })

    it('should handle HTTP error', () => {
      const errorMessage = 'Server error'

      service.getAllStocks().subscribe({
        next: () => fail('Expected an error'),
        error: error => {
          expect(error.status).toBe(500)
          expect(error.statusText).toBe('Internal Server Error')
        }
      })

      const req = httpMock.expectOne(`${environment.apiUrl}/stocks`)
      req.flush(errorMessage, {
        status: 500,
        statusText: 'Internal Server Error'
      })
    })
  })

  describe('getStock', () => {
    it('should return stock details for a valid ID', () => {
      const stockId = '1'

      service.getStock(stockId).subscribe(stockDetails => {
        expect(stockDetails).toEqual(mockStockDetails)
        expect(stockDetails.id).toBe(stockId)
      })

      const req = httpMock.expectOne(`${environment.apiUrl}/stocks/${stockId}`)
      expect(req.request.method).toBe('GET')
      req.flush(mockStockDetails)
    })

    it('should handle 404 error for non-existent stock', () => {
      const stockId = 'non-existent'

      service.getStock(stockId).subscribe({
        next: () => fail('Expected an error'),
        error: error => {
          expect(error.status).toBe(404)
          expect(error.statusText).toBe('Not Found')
        }
      })

      const req = httpMock.expectOne(`${environment.apiUrl}/stocks/${stockId}`)
      req.flush('Stock not found', { status: 404, statusText: 'Not Found' })
    })

    it('should make request with correct URL format', () => {
      const stockId = 'AAPL'

      service.getStock(stockId).subscribe()

      const req = httpMock.expectOne(`${environment.apiUrl}/stocks/AAPL`)
      expect(req.request.method).toBe('GET')
      expect(req.request.url).toBe(`${environment.apiUrl}/stocks/AAPL`)
      req.flush(mockStockDetails)
    })
  })

  describe('API URL configuration', () => {
    it('should use the correct base API URL', () => {
      service.getAllStocks().subscribe()

      const req = httpMock.expectOne(`${environment.apiUrl}/stocks`)
      expect(req.request.url).toContain(environment.apiUrl)
      req.flush([])
    })
  })
})
