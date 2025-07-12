import { Injectable } from '@angular/core'
import { BehaviorSubject, Observable } from 'rxjs'

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private currentTraderIdSubject = new BehaviorSubject<string | null>(
    localStorage.getItem('traderId')
  )
  public currentTraderId$ = this.currentTraderIdSubject.asObservable()

  login (traderId: string): void {
    localStorage.setItem('traderId', traderId)
    this.currentTraderIdSubject.next(traderId)
  }

  logout (): void {
    localStorage.removeItem('traderId')
    this.currentTraderIdSubject.next(null)
  }

  getCurrentTraderId (): string | null {
    return this.currentTraderIdSubject.value
  }

  isLoggedIn (): boolean {
    return !!this.getCurrentTraderId()
  }
}
