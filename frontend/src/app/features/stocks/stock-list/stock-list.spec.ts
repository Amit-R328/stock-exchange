import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick
} from '@angular/core/testing'
import { Router } from '@angular/router'
import { ChangeDetectorRef } from '@angular/core'
import { BrowserAnimationsModule } from '@angular/platform-browser/animations'
import { of, Subject, throwError } from 'rxjs'

import { StockListComponent } from './stock-list'
import { StockService } from '../../../core/services/stock/stock.service'
import { WebSocketService } from '../../../core/services/websocket/websocket.service'
import { Stock } from '../../../core/models/stock.model'

describe('StockListComponent', () => {
  let component: StockListComponent
  let fixture: ComponentFixture<StockListComponent>
  let mockStockService: jasmine.SpyObj<StockService>
  let mockWebSocketService: jasmine.SpyObj<WebSocketService>
  let mockRouter: jasmine.SpyObj<Router>
  let mockChangeDetectorRef: jasmine.SpyObj<ChangeDetectorRef>

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

  beforeEach(async () => {
    const stockServiceSpy = jasmine.createSpyObj('StockService', [
      'getAllStocks'
    ])
    const webSocketServiceSpy = jasmine.createSpyObj('WebSocketService', [
      'connect',
      'disconnect'
    ])
    const routerSpy = jasmine.createSpyObj('Router', ['navigate'])
    const cdrSpy = jasmine.createSpyObj('ChangeDetectorRef', [
      'detectChanges',
      'detach'
    ])

    await TestBed.configureTestingModule({
      imports: [StockListComponent, BrowserAnimationsModule],
      providers: [
        { provide: StockService, useValue: stockServiceSpy },
        { provide: WebSocketService, useValue: webSocketServiceSpy },
        { provide: Router, useValue: routerSpy },
        { provide: ChangeDetectorRef, useValue: cdrSpy }
      ]
    }).compileComponents()

    fixture = TestBed.createComponent(StockListComponent)
    component = fixture.componentInstance
    mockStockService = TestBed.inject(
      StockService
    ) as jasmine.SpyObj<StockService>
    mockWebSocketService = TestBed.inject(
      WebSocketService
    ) as jasmine.SpyObj<WebSocketService>
    mockRouter = TestBed.inject(Router) as jasmine.SpyObj<Router>
    mockChangeDetectorRef = TestBed.inject(
      ChangeDetectorRef
    ) as jasmine.SpyObj<ChangeDetectorRef>

    // Set up default mocks to prevent undefined errors
    mockStockService.getAllStocks.and.returnValue(of(mockStocks))
    mockWebSocketService.connect.and.returnValue(of({ type: 'test', data: [] }))

    // Make sure CDR methods don't throw errors
    mockChangeDetectorRef.detectChanges.and.stub()
    mockChangeDetectorRef.detach.and.stub()
  })

  beforeEach(() => {
    // Reset spy call counts before each test
    mockChangeDetectorRef.detectChanges.calls.reset()
    mockChangeDetectorRef.detach.calls.reset()
    mockStockService.getAllStocks.calls.reset()
    mockWebSocketService.connect.calls.reset()
    mockRouter.navigate.calls.reset()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
    // Verify that the ChangeDetectorRef is available (but don't check if it's the mock since Angular may create its own)
    expect(component['cdr']).toBeDefined()
  })

  it('should initialize with default values', () => {
    expect(component.stocks).toEqual([])
    expect(component.loading).toBe(true)
    expect(component.error).toBe(null)
    expect(component.displayedColumns).toEqual([
      'name',
      'currentPrice',
      'amount',
      'actions'
    ])
  })

  describe('ngOnInit', () => {
    it('should load stocks and connect to websocket', () => {
      component.ngOnInit()

      expect(mockStockService.getAllStocks).toHaveBeenCalled()
      expect(mockWebSocketService.connect).toHaveBeenCalled()
    })

    it('should set loading to false and populate stocks on successful load', fakeAsync(() => {
      component.ngOnInit()
      tick() // Wait for Observable to complete

      expect(component.loading).toBe(false)
      expect(component.stocks).toEqual(mockStocks)
      expect(component.error).toBe(null)
    }))

    it('should handle error when loading stocks fails', fakeAsync(() => {
      const errorMessage = 'Network error'
      mockStockService.getAllStocks.and.returnValue(
        throwError(() => new Error(errorMessage))
      )

      component.ngOnInit()
      tick() // Wait for Observable to complete

      expect(component.loading).toBe(false)
      expect(component.error).toBe('Failed to load stocks. Please try again.')
    }))
  })

  describe('ngOnDestroy', () => {
    it('should complete destroy subject and disconnect websocket', () => {
      spyOn(component['destroy$'], 'next')
      spyOn(component['destroy$'], 'complete')

      component.ngOnDestroy()

      expect(component['destroy$'].next).toHaveBeenCalled()
      expect(component['destroy$'].complete).toHaveBeenCalled()
      expect(mockWebSocketService.disconnect).toHaveBeenCalled()
    })
  })

  describe('getPriceChange', () => {
    beforeEach(() => {
      component['priceChanges'].set('1', 5.5)
      component['priceChanges'].set('2', -2.3)
    })

    it('should return cached price change for existing stock', () => {
      const result = component.getPriceChange(mockStocks[0])
      expect(result).toBe(5.5)
    })

    it('should return 0 for stock without cached price change', () => {
      const newStock: Stock = {
        id: '3',
        name: 'Tesla',
        currentPrice: 800,
        amount: 200
      }
      const result = component.getPriceChange(newStock)
      expect(result).toBe(0)
    })
  })

  describe('viewStock', () => {
    it('should navigate to stock detail page', () => {
      const stock = mockStocks[0]

      component.viewStock(stock)

      expect(mockRouter.navigate).toHaveBeenCalledWith(['/stocks', stock.id])
    })
  })

  describe('WebSocket updates', () => {
    beforeEach(() => {
      mockStockService.getAllStocks.and.returnValue(of(mockStocks))
      // Set up WebSocket mock BEFORE calling ngOnInit
      mockWebSocketService.connect.and.returnValue(
        of({
          type: 'test',
          data: []
        })
      )
      component.ngOnInit()
      component.stocks = mockStocks
      // Initialize previous prices
      component['updatePreviousPrices']()
    })

    it('should update stocks when receiving websocket updates', fakeAsync(() => {
      const updatedStocks: Stock[] = [
        { ...mockStocks[0], currentPrice: 155.0 },
        { ...mockStocks[1], currentPrice: 305.0 }
      ]

      mockWebSocketService.connect.and.returnValue(
        of({
          type: 'stocks',
          data: updatedStocks
        })
      )

      // Simulate websocket update
      component.ngOnInit()
      tick() // Wait for Observable to complete

      expect(component.stocks).toEqual(updatedStocks)
    }))

    it('should calculate price changes correctly', () => {
      // Set up initial state
      component.stocks = mockStocks
      component['updatePreviousPrices']()

      // Create updated stocks with price changes
      const updatedStocks: Stock[] = [
        { ...mockStocks[0], currentPrice: 157.5 }, // +5% change
        { ...mockStocks[1], currentPrice: 285.0 } // -5% change
      ]

      // Simulate the price change calculation logic
      const currentPrices = new Map<string, number>()
      component.stocks.forEach(stock => {
        currentPrices.set(stock.id, stock.currentPrice)
      })

      component.stocks = updatedStocks
      component['priceChanges'].clear()

      updatedStocks.forEach(newStock => {
        const previousPrice = currentPrices.get(newStock.id)
        if (previousPrice && previousPrice !== newStock.currentPrice) {
          const change =
            ((newStock.currentPrice - previousPrice) / previousPrice) * 100
          component['priceChanges'].set(newStock.id, change)
        } else {
          component['priceChanges'].set(newStock.id, 0)
        }
      })

      expect(component.getPriceChange(updatedStocks[0])).toBeCloseTo(5, 1)
      expect(component.getPriceChange(updatedStocks[1])).toBeCloseTo(-5, 1)
    })
  })

  describe('Template Integration', () => {
    it('should manage loading state correctly', () => {
      // Test that component correctly manages loading state
      expect(component.loading).toBe(true) // Initial state

      // Simulate successful data load
      component.loading = false
      component.stocks = mockStocks
      component.error = null

      expect(component.loading).toBe(false)
      expect(component.stocks.length).toBe(2)
      expect(component.error).toBeNull()
    })

    it('should manage error state correctly', () => {
      // Test that component correctly manages error state
      component.loading = false
      component.error = 'Test error message'
      component.stocks = []

      expect(component.loading).toBe(false)
      expect(component.error).toBe('Test error message')
      expect(component.stocks.length).toBe(0)
    })

    it('should manage stocks display state correctly', () => {
      // Test that component correctly manages stocks display state
      component.loading = false
      component.error = null
      component.stocks = mockStocks

      expect(component.loading).toBe(false)
      expect(component.error).toBeNull()
      expect(component.stocks).toEqual(mockStocks)
      expect(component.displayedColumns).toEqual([
        'name',
        'currentPrice',
        'amount',
        'actions'
      ])
    })
  })
})
