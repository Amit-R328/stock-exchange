// app/app.ts
import { Component, OnInit, OnDestroy } from '@angular/core'
import { CommonModule } from '@angular/common'
import { Router, RouterOutlet, RouterLink } from '@angular/router'
import { MatToolbarModule } from '@angular/material/toolbar'
import { MatButtonModule } from '@angular/material/button'
import { AuthService } from './core/services/auth/auth.service'
import { Observable, Subject, takeUntil } from 'rxjs'

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    RouterLink,
    MatToolbarModule,
    MatButtonModule
  ],
  templateUrl: './app.html',
  styleUrls: ['./app.scss']
})
export class App implements OnInit, OnDestroy {
  isLoggedIn$: Observable<string | null>
  private destroy$ = new Subject<void>()

  constructor (private authService: AuthService, private router: Router) {
    // Initialize the Observable in constructor to avoid race conditions
    this.isLoggedIn$ = this.authService.currentTraderId$
  }

  ngOnInit (): void {
    this.authService.currentTraderId$
      .pipe(takeUntil(this.destroy$))
      .subscribe(traderId => {
        if (!traderId && !this.router.url.includes('login')) {
          this.router.navigate(['/login'])
        }
      })
  }

  ngOnDestroy (): void {
    this.destroy$.next()
    this.destroy$.complete()
  }

  logout (): void {
    this.authService.logout()
    this.router.navigate(['/login'])
  }
}
