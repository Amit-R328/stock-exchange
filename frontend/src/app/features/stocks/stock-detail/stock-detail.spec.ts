import { ComponentFixture, TestBed } from '@angular/core/testing'
import { ActivatedRoute } from '@angular/router'
import { BrowserAnimationsModule } from '@angular/platform-browser/animations'
import { of } from 'rxjs'

import { StockDetailComponent } from './stock-detail'
import { StockService } from '../../../core/services/stock/stock.service'
import { OrderService } from '../../../core/services/order/order.service'
import { AuthService } from '../../../core/services/auth/auth.service'
import { TraderService } from '../../../core/services/trader/trader.service'

describe('StockDetailComponent', () => {
  let component: StockDetailComponent
  let fixture: ComponentFixture<StockDetailComponent>
  let mockStockService: jasmine.SpyObj<StockService>
  let mockOrderService: jasmine.SpyObj<OrderService>
  let mockAuthService: jasmine.SpyObj<AuthService>
  let mockTraderService: jasmine.SpyObj<TraderService>
  let mockActivatedRoute: jasmine.SpyObj<ActivatedRoute>

  beforeEach(async () => {
    const stockServiceSpy = jasmine.createSpyObj('StockService', [
      'getStock',
      'getStockHistory'
    ])
    const orderServiceSpy = jasmine.createSpyObj('OrderService', [
      'createOrder'
    ])
    const authServiceSpy = jasmine.createSpyObj(
      'AuthService',
      ['getCurrentTraderId'],
      {
        currentTraderId$: of('trader123')
      }
    )
    const traderServiceSpy = jasmine.createSpyObj('TraderService', [
      'getTrader'
    ])
    const activatedRouteSpy = jasmine.createSpyObj('ActivatedRoute', [], {
      params: of({ id: '1' }),
      snapshot: {
        paramMap: jasmine.createSpyObj('ParamMap', ['get'])
      }
    })

    // Set up the paramMap.get method to return the stock ID
    ;(activatedRouteSpy.snapshot.paramMap.get as jasmine.Spy).and.returnValue(
      '1'
    )

    await TestBed.configureTestingModule({
      imports: [StockDetailComponent, BrowserAnimationsModule],
      providers: [
        { provide: StockService, useValue: stockServiceSpy },
        { provide: OrderService, useValue: orderServiceSpy },
        { provide: AuthService, useValue: authServiceSpy },
        { provide: TraderService, useValue: traderServiceSpy },
        { provide: ActivatedRoute, useValue: activatedRouteSpy }
      ]
    }).compileComponents()

    mockStockService = TestBed.inject(
      StockService
    ) as jasmine.SpyObj<StockService>
    mockOrderService = TestBed.inject(
      OrderService
    ) as jasmine.SpyObj<OrderService>
    mockAuthService = TestBed.inject(AuthService) as jasmine.SpyObj<AuthService>
    mockTraderService = TestBed.inject(
      TraderService
    ) as jasmine.SpyObj<TraderService>
    mockActivatedRoute = TestBed.inject(
      ActivatedRoute
    ) as jasmine.SpyObj<ActivatedRoute>

    fixture = TestBed.createComponent(StockDetailComponent)
    component = fixture.componentInstance
  })

  it('should create', () => {
    // Mock the required service calls
    mockStockService.getStock.and.returnValue(
      of({
        id: '1',
        name: 'Test Stock',
        currentPrice: 100,
        amount: 1000,
        openOrders: [],
        lastTransactions: []
      })
    )
    mockStockService.getStockHistory.and.returnValue(
      of({
        stockId: '1',
        days: 30,
        history: [
          { timestamp: '2024-01-01T00:00:00Z', price: 100, volume: 1000 },
          { timestamp: '2024-01-02T00:00:00Z', price: 105, volume: 1200 }
        ]
      })
    )
    mockTraderService.getTrader.and.returnValue(
      of({
        id: 'trader123',
        name: 'Test Trader',
        money: 5000,
        holdings: {},
        openOrders: []
      })
    )

    fixture.detectChanges()
    expect(component).toBeTruthy()
  })
})
