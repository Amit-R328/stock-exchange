import { Component } from '@angular/core'
import { CommonModule } from '@angular/common'
import { FormsModule } from '@angular/forms'
import { Router } from '@angular/router'
import { MatCardModule } from '@angular/material/card'
import { MatFormFieldModule } from '@angular/material/form-field'
import { MatInputModule } from '@angular/material/input'
import { MatButtonModule } from '@angular/material/button'
import { AuthService } from '../../core/services/auth/auth.service'
import { TraderService } from '../../core/services/trader/trader.service'

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule
  ],
  templateUrl: './login.html',
  styleUrls: ['./login.scss']
})
export class LoginComponent {
  traderId = ''
  error = ''

  constructor (
    private authService: AuthService,
    private traderService: TraderService,
    private router: Router
  ) {}

  login (): void {
    if (!this.traderId) {
      this.error = 'Please enter a trader ID'
      return
    }

    this.traderService.getTrader(this.traderId).subscribe({
      next: () => {
        this.authService.login(this.traderId)
        this.router.navigate(['/stocks'])
      },
      error: () => {
        this.error = 'Invalid trader ID'
      }
    })
  }
}
