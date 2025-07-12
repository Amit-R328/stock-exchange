import { ComponentFixture, TestBed } from '@angular/core/testing'
import { BrowserAnimationsModule } from '@angular/platform-browser/animations'
import { of } from 'rxjs'

import { PersonalComponent } from './personal'
import { AuthService } from '../../core/services/auth/auth.service'
import { TraderService } from '../../core/services/trader/trader.service'
import { StockService } from '../../core/services/stock/stock.service'
import { OrderService } from '../../core/services/order/order.service'
import { WebSocketService } from '../../core/services/websocket/websocket.service'

describe('PersonalComponent', () => {
  let component: PersonalComponent
  let fixture: ComponentFixture<PersonalComponent>
  let mockAuthService: jasmine.SpyObj<AuthService>
  let mockTraderService: jasmine.SpyObj<TraderService>
  let mockStockService: jasmine.SpyObj<StockService>
  let mockOrderService: jasmine.SpyObj<OrderService>
  let mockWebSocketService: jasmine.SpyObj<WebSocketService>

  beforeEach(async () => {
    const authServiceSpy = jasmine.createSpyObj(
      'AuthService',
      ['getCurrentTraderId'],
      {
        currentTraderId$: of('trader123')
      }
    )
    const traderServiceSpy = jasmine.createSpyObj('TraderService', [
      'getTrader',
      'getTraderTransactions',
      'getTraderPerformance'
    ])
    const stockServiceSpy = jasmine.createSpyObj('StockService', [
      'getAllStocks'
    ])
    const orderServiceSpy = jasmine.createSpyObj('OrderService', [
      'cancelOrder'
    ])
    const webSocketServiceSpy = jasmine.createSpyObj('WebSocketService', [
      'connect',
      'disconnect'
    ])

    await TestBed.configureTestingModule({
      imports: [PersonalComponent, BrowserAnimationsModule],
      providers: [
        { provide: AuthService, useValue: authServiceSpy },
        { provide: TraderService, useValue: traderServiceSpy },
        { provide: StockService, useValue: stockServiceSpy },
        { provide: OrderService, useValue: orderServiceSpy },
        { provide: WebSocketService, useValue: webSocketServiceSpy }
      ]
    }).compileComponents()

    mockAuthService = TestBed.inject(AuthService) as jasmine.SpyObj<AuthService>
    mockTraderService = TestBed.inject(
      TraderService
    ) as jasmine.SpyObj<TraderService>
    mockStockService = TestBed.inject(
      StockService
    ) as jasmine.SpyObj<StockService>
    mockOrderService = TestBed.inject(
      OrderService
    ) as jasmine.SpyObj<OrderService>
    mockWebSocketService = TestBed.inject(
      WebSocketService
    ) as jasmine.SpyObj<WebSocketService>

    fixture = TestBed.createComponent(PersonalComponent)
    component = fixture.componentInstance
  })

  it('should create', () => {
    // Mock the required service calls
    mockAuthService.getCurrentTraderId.and.returnValue('trader123')
    mockTraderService.getTrader.and.returnValue(
      of({
        id: 'trader123',
        name: 'Test Trader',
        money: 5000,
        holdings: { '1': 10 },
        openOrders: []
      })
    )
    mockTraderService.getTraderTransactions.and.returnValue(
      of({
        transactions: [],
        profitLoss: 0
      })
    )
    mockTraderService.getTraderPerformance.and.returnValue(
      of({
        traderId: 'trader123',
        days: 30,
        performance: [
          { date: '2024-01-01', portfolioValue: 10000, profitLoss: 0, cashBalance: 5000 }
        ],
        portfolio: { totalValue: 10000, cashBalance: 5000, holdings: [] },
        activity: []
      })
    )
    mockStockService.getAllStocks.and.returnValue(
      of([{ id: '1', name: 'Test Stock', currentPrice: 100, amount: 1000 }])
    )
    mockWebSocketService.connect.and.returnValue(of({ type: 'test', data: {} }))

    fixture.detectChanges()
    expect(component).toBeTruthy()
  })
})
