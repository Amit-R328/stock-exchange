import { TestBed } from '@angular/core/testing'
import { AuthService } from './auth.service'

describe('AuthService', () => {
  let service: AuthService

  beforeEach(() => {
    TestBed.configureTestingModule({})
    service = TestBed.inject(AuthService)

    // Clear localStorage before each test
    localStorage.clear()
  })

  afterEach(() => {
    // Clean up localStorage after each test
    localStorage.clear()
  })

  it('should be created', () => {
    expect(service).toBeTruthy()
  })

  describe('login', () => {
    it('should store trader ID in localStorage and update subject', () => {
      const traderId = 'trader123'

      service.login(traderId)

      expect(localStorage.getItem('traderId')).toBe(traderId)
      expect(service.getCurrentTraderId()).toBe(traderId)
    })

    it('should emit the trader ID through currentTraderId$ observable', done => {
      const traderId = 'trader456'

      service.currentTraderId$.subscribe(id => {
        if (id === traderId) {
          expect(id).toBe(traderId)
          done()
        }
      })

      service.login(traderId)
    })

    it('should overwrite existing trader ID', () => {
      const firstTraderId = 'trader1'
      const secondTraderId = 'trader2'

      service.login(firstTraderId)
      expect(service.getCurrentTraderId()).toBe(firstTraderId)

      service.login(secondTraderId)
      expect(service.getCurrentTraderId()).toBe(secondTraderId)
      expect(localStorage.getItem('traderId')).toBe(secondTraderId)
    })
  })

  describe('logout', () => {
    it('should remove trader ID from localStorage and update subject', () => {
      const traderId = 'trader123'

      // First login
      service.login(traderId)
      expect(service.getCurrentTraderId()).toBe(traderId)

      // Then logout
      service.logout()
      expect(localStorage.getItem('traderId')).toBeNull()
      expect(service.getCurrentTraderId()).toBeNull()
    })

    it('should emit null through currentTraderId$ observable', done => {
      const traderId = 'trader123'
      let emissionCount = 0

      service.currentTraderId$.subscribe(id => {
        emissionCount++
        if (emissionCount === 1) {
          // Initial null value
          expect(id).toBeNull()
        } else if (emissionCount === 2) {
          // After login
          expect(id).toBe(traderId)
        } else if (emissionCount === 3) {
          // After logout
          expect(id).toBeNull()
          done()
        }
      })

      service.login(traderId)
      service.logout()
    })

    it('should handle logout when not logged in', () => {
      // Logout when no trader is logged in
      service.logout()

      expect(localStorage.getItem('traderId')).toBeNull()
      expect(service.getCurrentTraderId()).toBeNull()
    })
  })

  describe('getCurrentTraderId', () => {
    it('should return current trader ID when logged in', () => {
      const traderId = 'trader123'
      service.login(traderId)

      expect(service.getCurrentTraderId()).toBe(traderId)
    })

    it('should return null when not logged in', () => {
      expect(service.getCurrentTraderId()).toBeNull()
    })
  })

  describe('isLoggedIn', () => {
    it('should return true when trader is logged in', () => {
      const traderId = 'trader123'
      service.login(traderId)

      expect(service.isLoggedIn()).toBe(true)
    })

    it('should return false when no trader is logged in', () => {
      expect(service.isLoggedIn()).toBe(false)
    })

    it('should return false after logout', () => {
      const traderId = 'trader123'
      service.login(traderId)
      expect(service.isLoggedIn()).toBe(true)

      service.logout()
      expect(service.isLoggedIn()).toBe(false)
    })
  })

  describe('localStorage persistence', () => {
    it('should initialize with existing localStorage value', () => {
      const traderId = 'persistedTrader'
      localStorage.setItem('traderId', traderId)

      // Create a new service instance
      const newService = new AuthService()

      expect(newService.getCurrentTraderId()).toBe(traderId)
      expect(newService.isLoggedIn()).toBe(true)
    })

    it('should initialize with null when localStorage is empty', () => {
      // Ensure localStorage is empty
      localStorage.removeItem('traderId')

      // Create a new service instance
      const newService = new AuthService()

      expect(newService.getCurrentTraderId()).toBeNull()
      expect(newService.isLoggedIn()).toBe(false)
    })
  })

  describe('currentTraderId$ observable', () => {
    it('should emit initial value immediately', done => {
      service.currentTraderId$.subscribe(id => {
        expect(id).toBeNull() // Initial value should be null
        done()
      })
    })

    it('should emit changes when trader ID changes', () => {
      const emittedValues: (string | null)[] = []

      service.currentTraderId$.subscribe(id => {
        emittedValues.push(id)
      })

      service.login('trader1')
      service.login('trader2')
      service.logout()

      expect(emittedValues).toEqual([null, 'trader1', 'trader2', null])
    })
  })
})
