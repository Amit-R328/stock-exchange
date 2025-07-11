import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { Router } from '@angular/router'
import { ApiService } from '../../services/api.service'

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="login-container">
      <h2>Select a Trader</h2>
      <div class="traders-list">
        <button
          *ngFor="let trader of traders"
          (click)="selectTrader(trader.id)"
        >
          {{ trader.name }}
        </button>
      </div>
    </div>
  `,
  styles: [
    `
      .login-container {
        max-width: 400px;
        margin: 50px auto;
        padding: 20px;
      }
      .traders-list {
        display: flex;
        flex-direction: column;
        gap: 10px;
      }
      button {
        padding: 10px;
        cursor: pointer;
      }
    `
  ]
})
export class LoginComponent implements OnInit {
  traders: any[] = []

  constructor (private apiService: ApiService, private router: Router) {}

  ngOnInit () {
    this.apiService.getTraders().subscribe(
      traders => (this.traders = traders),
      error => console.error('Failed to load traders:', error)
    )
  }

  selectTrader (traderId: string) {
    localStorage.setItem('traderId', traderId)
    this.router.navigate(['/stocks'])
  }
}
