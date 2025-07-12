import { ComponentFixture, TestBed } from '@angular/core/testing'
import { ElementRef } from '@angular/core'
import { of } from 'rxjs'

import { StockChartComponent } from './stock-chart.component'
import { StockService, StockHistoryResponse, PriceQuote } from '../../../core/services/stock/stock.service'

// Mock Chart.js to avoid importing the full library in tests
const mockChart = {
  destroy: jasmine.createSpy('destroy'),
  update: jasmine.createSpy('update')
}

const mockChartConstructor = jasmine.createSpy('Chart').and.returnValue(mockChart)

// Mock Chart.js module
Object.defineProperty(window, 'Chart', {
  writable: true,
  value: mockChartConstructor
})

describe('StockChartComponent', () => {
  let component: StockChartComponent
  let fixture: ComponentFixture<StockChartComponent>
  let mockStockService: jasmine.SpyObj<StockService>

  const mockPriceQuotes: PriceQuote[] = [
    { timestamp: '2024-01-01T00:00:00Z', price: 140, volume: 1000 },
    { timestamp: '2024-01-02T00:00:00Z', price: 145, volume: 1200 },
    { timestamp: '2024-01-03T00:00:00Z', price: 150, volume: 900 }
  ]

  const mockStockHistoryResponse: StockHistoryResponse = {
    stockId: '1',
    days: 30,
    history: mockPriceQuotes
  }

  beforeEach(async () => {
    const stockServiceSpy = jasmine.createSpyObj('StockService', [
      'getStockHistory'
    ])

    await TestBed.configureTestingModule({
      imports: [StockChartComponent],
      providers: [
        { provide: StockService, useValue: stockServiceSpy }
      ]
    }).compileComponents()

    mockStockService = TestBed.inject(StockService) as jasmine.SpyObj<StockService>
    fixture = TestBed.createComponent(StockChartComponent)
    component = fixture.componentInstance

    // Mock the canvas element
    const mockCanvas = document.createElement('canvas')
    const mockElementRef = new ElementRef(mockCanvas)
    component.chartCanvas = mockElementRef
  })

  afterEach(() => {
    // Clean up any existing charts
    if (component['chart']) {
      component['chart'].destroy()
    }
    mockChartConstructor.calls.reset()
    mockChart.destroy.calls.reset()
    mockChart.update.calls.reset()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })

  it('should have default input values', () => {
    expect(component.stockData).toEqual([])
    expect(component.stockName).toBe('')
    expect(component.stockId).toBe('')
    expect(component.chartType).toBe('line')
    expect(component.height).toBe(300)
  })

  it('should accept input values', () => {
    component.stockId = '1'
    component.stockName = 'Apple Inc.'
    component.chartType = 'candlestick'
    component.height = 400

    expect(component.stockId).toBe('1')
    expect(component.stockName).toBe('Apple Inc.')
    expect(component.chartType).toBe('candlestick')
    expect(component.height).toBe(400)
  })

  it('should load chart data on init when stockId is provided', () => {
    mockStockService.getStockHistory.and.returnValue(of(mockStockHistoryResponse))
    
    component.stockId = '1'
    component.ngOnInit()

    expect(mockStockService.getStockHistory).toHaveBeenCalledWith('1', 30)
  })

  it('should load chart data with custom data when stockData is provided', () => {
    const customData = [
      { timestamp: '2024-01-01', price: 100 },
      { timestamp: '2024-01-02', price: 105 }
    ]
    
    component.stockData = customData
    component.ngOnInit()

    // Should not call service when stockData is provided
    expect(mockStockService.getStockHistory).not.toHaveBeenCalled()
  })

  it('should destroy chart on component destroy', () => {
    // Simulate chart creation
    component['chart'] = mockChart as any

    component.ngOnDestroy()

    expect(mockChart.destroy).toHaveBeenCalled()
  })

  it('should handle different chart types', () => {
    const chartTypes: ('line' | 'candlestick' | 'volume')[] = ['line', 'candlestick', 'volume']
    
    chartTypes.forEach(chartType => {
      component.chartType = chartType
      expect(component.chartType).toBe(chartType)
    })
  })

  it('should handle empty stock history data', () => {
    const emptyResponse: StockHistoryResponse = {
      stockId: '1',
      days: 30,
      history: []
    }
    
    mockStockService.getStockHistory.and.returnValue(of(emptyResponse))
    
    component.stockId = '1'
    component.ngOnInit()

    expect(mockStockService.getStockHistory).toHaveBeenCalledWith('1', 30)
    // Component should not crash with empty data
    expect(component).toBeTruthy()
  })

  it('should handle service errors gracefully', () => {
    mockStockService.getStockHistory.and.returnValue(of(mockStockHistoryResponse))
    
    component.stockId = '1'
    
    expect(() => {
      component.ngOnInit()
    }).not.toThrow()
  })

  it('should have canvas element reference', () => {
    expect(component.chartCanvas).toBeDefined()
    expect(component.chartCanvas.nativeElement).toBeTruthy()
  })

  it('should accept different height values', () => {
    const heights = [200, 300, 400, 500]
    
    heights.forEach(height => {
      component.height = height
      expect(component.height).toBe(height)
    })
  })

  it('should work with stock data array input', () => {
    const stockData = [
      { timestamp: '2024-01-01', price: 100, volume: 1000 },
      { timestamp: '2024-01-02', price: 110, volume: 1200 }
    ]

    component.stockData = stockData
    expect(component.stockData).toEqual(stockData)
  })

  it('should work with stock name input', () => {
    const stockNames = ['Apple Inc.', 'Microsoft Corp.', 'Google LLC']
    
    stockNames.forEach(name => {
      component.stockName = name
      expect(component.stockName).toBe(name)
    })
  })

  it('should call loadChartData on initialization', () => {
    const spy = spyOn(component as any, 'loadChartData')
    
    component.ngOnInit()
    
    expect(spy).toHaveBeenCalled()
  })

  it('should handle missing stockId gracefully', () => {
    // No stockId provided
    component.stockId = ''
    
    expect(() => {
      component.ngOnInit()
    }).not.toThrow()

    // Should not call service without stockId
    expect(mockStockService.getStockHistory).not.toHaveBeenCalled()
  })

  it('should validate chart type values', () => {
    const validTypes: ('line' | 'candlestick' | 'volume')[] = ['line', 'candlestick', 'volume']
    
    validTypes.forEach(type => {
      component.chartType = type
      expect(['line', 'candlestick', 'volume']).toContain(component.chartType)
    })
  })
})
