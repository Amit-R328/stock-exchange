import { ComponentFixture, TestBed } from '@angular/core/testing'
import { BrowserAnimationsModule } from '@angular/platform-browser/animations'
import { of } from 'rxjs'

import { TradersComponent } from './traders'
import { TraderService } from '../../core/services/trader/trader.service'

describe('TradersComponent', () => {
  let component: TradersComponent
  let fixture: ComponentFixture<TradersComponent>
  let mockTraderService: jasmine.SpyObj<TraderService>

  beforeEach(async () => {
    const traderServiceSpy = jasmine.createSpyObj('TraderService', [
      'getAllTraders'
    ])

    await TestBed.configureTestingModule({
      imports: [TradersComponent, BrowserAnimationsModule],
      providers: [{ provide: TraderService, useValue: traderServiceSpy }]
    }).compileComponents()

    mockTraderService = TestBed.inject(
      TraderService
    ) as jasmine.SpyObj<TraderService>
    fixture = TestBed.createComponent(TradersComponent)
    component = fixture.componentInstance
  })

  it('should create', () => {
    // Mock the required service calls
    mockTraderService.getAllTraders.and.returnValue(
      of([
        { id: 'trader1', name: 'Test Trader 1' },
        { id: 'trader2', name: 'Test Trader 2' }
      ])
    )

    fixture.detectChanges()
    expect(component).toBeTruthy()
  })
})
