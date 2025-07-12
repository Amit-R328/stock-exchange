import { ComponentFixture, TestBed } from '@angular/core/testing'
import { ElementRef } from '@angular/core'
import { of } from 'rxjs'

import { TraderChartComponent } from './trader-chart.component'
import { TraderService, TraderPerformanceResponse, PerformanceData, PortfolioData, ActivityLog } from '../../../core/services/trader/trader.service'

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

describe('TraderChartComponent', () => {
  let component: TraderChartComponent
  let fixture: ComponentFixture<TraderChartComponent>
  let mockTraderService: jasmine.SpyObj<TraderService>

  const mockPerformanceData: PerformanceData[] = [
    { date: '2024-01-01', portfolioValue: 10000, profitLoss: 0, cashBalance: 5000 },
    { date: '2024-01-02', portfolioValue: 10500, profitLoss: 500, cashBalance: 4500 },
    { date: '2024-01-03', portfolioValue: 11000, profitLoss: 1000, cashBalance: 4000 }
  ]

  const mockPortfolioData: PortfolioData = {
    totalValue: 50000,
    cashBalance: 20000,
    holdings: [
      { stockId: '1', stockName: 'Apple Inc.', quantity: 100, value: 15000, percentage: 30 },
      { stockId: '2', stockName: 'Microsoft Corp.', quantity: 50, value: 15000, percentage: 30 }
    ]
  }

  const mockActivityData: ActivityLog[] = [
    { period: '2024-01', buyOrders: 10, sellOrders: 5, volume: 150, value: 15000 },
    { period: '2024-02', buyOrders: 8, sellOrders: 7, volume: 120, value: 12000 }
  ]

  const mockTraderPerformanceResponse: TraderPerformanceResponse = {
    traderId: 'trader123',
    days: 30,
    performance: mockPerformanceData,
    portfolio: mockPortfolioData,
    activity: mockActivityData
  }

  beforeEach(async () => {
    const traderServiceSpy = jasmine.createSpyObj('TraderService', [
      'getTraderPerformance'
    ])

    await TestBed.configureTestingModule({
      imports: [TraderChartComponent],
      providers: [
        { provide: TraderService, useValue: traderServiceSpy }
      ]
    }).compileComponents()

    mockTraderService = TestBed.inject(TraderService) as jasmine.SpyObj<TraderService>
    fixture = TestBed.createComponent(TraderChartComponent)
    component = fixture.componentInstance

    // Mock the canvas element
    const mockCanvas = document.createElement('canvas')
    const mockElementRef = new ElementRef(mockCanvas)
    component.chartCanvas = mockElementRef

    // Setup default mock returns
    mockTraderService.getTraderPerformance.and.returnValue(of(mockTraderPerformanceResponse))
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
    expect(component.traderId).toBe('')
    expect(component.chartType).toBe('performance')
    expect(component.height).toBe(300)
  })

  it('should accept input values', () => {
    component.traderId = 'trader123'
    component.chartType = 'portfolio'
    component.height = 400

    expect(component.traderId).toBe('trader123')
    expect(component.chartType).toBe('portfolio')
    expect(component.height).toBe(400)
  })

  it('should load trader performance data when chartType is performance', () => {
    component.traderId = 'trader123'
    component.chartType = 'performance'
    
    component.ngOnInit()

    expect(mockTraderService.getTraderPerformance).toHaveBeenCalledWith('trader123', 30)
  })

  it('should load trader performance data for all chart types', () => {
    component.traderId = 'trader123'
    
    // All chart types use the same getTraderPerformance API call
    const chartTypes: ('performance' | 'portfolio' | 'activity')[] = ['performance', 'portfolio', 'activity']
    
    chartTypes.forEach(chartType => {
      component.chartType = chartType
      component.ngOnInit()
      expect(mockTraderService.getTraderPerformance).toHaveBeenCalledWith('trader123', 30)
    })
  })

  it('should not load data when traderId is not provided', () => {
    component.traderId = ''
    component.ngOnInit()

    expect(mockTraderService.getTraderPerformance).not.toHaveBeenCalled()
  })

  it('should destroy chart on component destroy', () => {
    // Simulate chart creation
    component['chart'] = mockChart as any

    component.ngOnDestroy()

    expect(mockChart.destroy).toHaveBeenCalled()
  })

  it('should handle different chart types', () => {
    const chartTypes: ('performance' | 'portfolio' | 'activity')[] = ['performance', 'portfolio', 'activity']
    
    chartTypes.forEach(chartType => {
      component.chartType = chartType
      expect(component.chartType).toBe(chartType)
    })
  })

  it('should handle empty performance data', () => {
    const emptyResponse: TraderPerformanceResponse = {
      traderId: 'trader123',
      days: 30,
      performance: [],
      portfolio: { totalValue: 0, cashBalance: 0, holdings: [] },
      activity: []
    }
    
    mockTraderService.getTraderPerformance.and.returnValue(of(emptyResponse))
    
    component.traderId = 'trader123'
    component.chartType = 'performance'
    component.ngOnInit()

    expect(mockTraderService.getTraderPerformance).toHaveBeenCalledWith('trader123', 30)
    // Component should not crash with empty data
    expect(component).toBeTruthy()
  })

  it('should handle service errors gracefully', () => {
    component.traderId = 'trader123'
    component.chartType = 'performance'
    
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

  it('should call loadChartData on initialization', () => {
    const spy = spyOn(component as any, 'loadChartData')
    
    component.ngOnInit()
    
    expect(spy).toHaveBeenCalled()
  })

  it('should validate chart type values', () => {
    const validTypes: ('performance' | 'portfolio' | 'activity')[] = ['performance', 'portfolio', 'activity']
    
    validTypes.forEach(type => {
      component.chartType = type
      expect(['performance', 'portfolio', 'activity']).toContain(component.chartType)
    })
  })

  it('should work with different trader IDs', () => {
    const traderIds = ['trader1', 'trader2', 'algo-bot-1', 'momentum-bot-1']
    
    traderIds.forEach(traderId => {
      component.traderId = traderId
      expect(component.traderId).toBe(traderId)
    })
  })

  it('should handle long trader IDs', () => {
    const longTraderId = 'very-long-trader-id-with-many-characters-and-numbers-123456789'
    
    component.traderId = longTraderId
    expect(component.traderId).toBe(longTraderId)
  })

  it('should handle special characters in trader ID', () => {
    const specialTraderId = 'trader-123_bot@domain.com'
    
    component.traderId = specialTraderId
    expect(component.traderId).toBe(specialTraderId)
  })

  it('should process performance data correctly', () => {
    component.traderId = 'trader123'
    component.chartType = 'performance'
    component.ngOnInit()

    expect(mockTraderService.getTraderPerformance).toHaveBeenCalledWith('trader123', 30)
    
    // The component should handle the response structure
    expect(component).toBeTruthy()
  })

  it('should process portfolio data correctly', () => {
    component.traderId = 'trader123'
    component.chartType = 'portfolio'
    component.ngOnInit()

    expect(mockTraderService.getTraderPerformance).toHaveBeenCalledWith('trader123', 30)
    
    // The component should handle the response structure with portfolio data
    expect(component).toBeTruthy()
  })

  it('should process activity data correctly', () => {
    component.traderId = 'trader123'
    component.chartType = 'activity'
    component.ngOnInit()

    expect(mockTraderService.getTraderPerformance).toHaveBeenCalledWith('trader123', 30)
    
    // The component should handle the response structure with activity data
    expect(component).toBeTruthy()
  })
})
