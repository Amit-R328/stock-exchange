import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { MatCardModule } from '@angular/material/card'
import { MatListModule } from '@angular/material/list'
import { MatIconModule } from '@angular/material/icon'
import { MatButtonModule } from '@angular/material/button'
import { MatTooltipModule } from '@angular/material/tooltip'
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner'
import { TraderService } from '../../core/services/trader/trader.service'

interface Trader {
  id: string
  name: string
}

@Component({
  selector: 'app-traders',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatListModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    MatProgressSpinnerModule
  ],
  templateUrl: './traders.html',
  styleUrls: ['./traders.scss']
})
export class TradersComponent implements OnInit {
  traders: Trader[] = []
  loading = true
  error: string | null = null

  constructor (private tradersService: TraderService) {}

  ngOnInit (): void {
    this.loadTraders()
  }

  private loadTraders (): void {
    this.loading = true
    this.error = null

    this.tradersService.getAllTraders().subscribe({
      next: traders => {
        this.traders = traders
        this.loading = false
      },
      error: err => {
        this.error = 'Failed to load traders. Please try again.'
        this.loading = false
        console.error('Error loading traders:', err)
      }
    })
  }
}
