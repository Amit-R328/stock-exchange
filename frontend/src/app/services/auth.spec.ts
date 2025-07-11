import { Injectable } from '@angular/core'
import { BehaviorSubject } from 'rxjs'

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private traderIdSubject = new BehaviorSubject<string | null>(
    localStorage.getItem('traderId')
  )

  currentTraderId$ = this.traderIdSubject.asObservable()

  login (traderId: string): void {
    localStorage.setItem('traderId', traderId)
    this.traderIdSubject.next(traderId)
  }

  logout (): void {
    localStorage.removeItem('traderId')
    this.traderIdSubject.next(null)
  }

  getCurrentTraderId (): string | null {
    return this.traderIdSubject.value
  }
}
