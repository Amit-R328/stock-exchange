import {
  Component,
  Input,
  OnInit,
  OnDestroy,
  ViewChild,
  ElementRef
} from '@angular/core'
import { CommonModule } from '@angular/common'
import { MatIconModule } from '@angular/material/icon'
import { Chart, ChartConfiguration, ChartType, registerables } from 'chart.js'
import { TraderService } from '../../../core/services/trader/trader.service'

// Register Chart.js components
Chart.register(...registerables)

@Component({
  selector: 'app-trader-chart',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './trader-chart.component.html',
  styleUrls: ['./trader-chart.component.scss']
})
export class TraderChartComponent implements OnInit, OnDestroy {
  @Input() traderData: any = {}
  @Input() traderId: string = ''
  @Input() chartType: 'performance' | 'portfolio' | 'activity' = 'performance'
  @Input() height: number = 300

  @ViewChild('chartCanvas', { static: true })
  chartCanvas!: ElementRef<HTMLCanvasElement>

  private chart?: Chart
  private performanceData: any = null

  constructor (private traderService: TraderService) {}

  ngOnInit () {
    this.loadChartData()
  }

  ngOnDestroy () {
    if (this.chart) {
      this.chart.destroy()
    }
  }

  private loadChartData () {
    if (this.traderId) {
      this.traderService.getTraderPerformance(this.traderId, 30).subscribe({
        next: response => {
          this.performanceData = response
          this.createChart()
        },
        error: error => {
          console.error('Error loading trader performance:', error)
          // Fallback to sample data
          this.createChart()
        }
      })
    } else {
      // Use existing data or generate sample data
      this.createChart()
    }
  }

  private createChart () {
    // Destroy existing chart before creating a new one
    if (this.chart) {
      this.chart.destroy()
      this.chart = undefined
    }

    const ctx = this.chartCanvas.nativeElement.getContext('2d')
    if (!ctx) return

    let config: ChartConfiguration

    switch (this.chartType) {
      case 'performance':
        config = this.createPerformanceChart()
        break
      case 'portfolio':
        config = this.createPortfolioChart()
        break
      case 'activity':
        config = this.createActivityChart()
        break
      default:
        config = this.createPerformanceChart()
    }

    this.chart = new Chart(ctx, config)
  }

  private createPerformanceChart (): ChartConfiguration {
    const data =
      this.performanceData?.performance || this.generatePerformanceData()

    return {
      type: 'line' as ChartType,
      data: {
        labels: data.map((item: any, index: number) => {
          if (item.date) {
            const date = new Date(item.date)
            return date.toLocaleDateString()
          }
          return `Day ${index + 1}`
        }),
        datasets: [
          {
            label: 'Portfolio Value',
            data: data.map((item: any) => item.portfolioValue),
            borderColor: '#4caf50',
            backgroundColor: 'rgba(76, 175, 80, 0.1)',
            borderWidth: 2,
            fill: true,
            tension: 0.1
          },
          {
            label: 'P&L',
            data: data.map((item: any) => item.profitLoss),
            borderColor: '#f44336',
            backgroundColor: 'rgba(244, 67, 54, 0.1)',
            borderWidth: 2,
            fill: false,
            tension: 0.1
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Trader Performance',
            font: { size: 16, weight: 'bold' }
          },
          legend: {
            display: true,
            position: 'top'
          }
        },
        scales: {
          x: {
            title: { display: true, text: 'Time Period' }
          },
          y: {
            title: { display: true, text: 'Value ($)' },
            ticks: {
              callback: function (value) {
                return '$' + value
              }
            }
          }
        }
      }
    }
  }

  private createPortfolioChart (): ChartConfiguration {
    const data =
      this.performanceData?.portfolio?.holdings || this.generatePortfolioData()

    return {
      type: 'doughnut' as ChartType,
      data: {
        labels: data.map((item: any) => item.stockName || item.stock),
        datasets: [
          {
            data: data.map((item: any) => item.value),
            backgroundColor: [
              '#1976d2',
              '#4caf50',
              '#ff9800',
              '#f44336',
              '#9c27b0',
              '#00bcd4'
            ],
            borderWidth: 2,
            borderColor: '#ffffff'
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Portfolio Distribution',
            font: { size: 16, weight: 'bold' }
          },
          legend: {
            display: true,
            position: 'right'
          },
          tooltip: {
            callbacks: {
              label: context => {
                const value = context.parsed as number
                const total = context.dataset.data.reduce(
                  (a: number, b: unknown) => {
                    const num = typeof b === 'number' ? b : 0
                    return a + num
                  },
                  0
                )
                const percentage =
                  total > 0 ? ((value / total) * 100).toFixed(1) : '0'
                return `${
                  context.label
                }: $${value.toLocaleString()} (${percentage}%)`
              }
            }
          }
        }
      }
    }
  }

  private createActivityChart (): ChartConfiguration {
    const data = this.performanceData?.activity || this.generateActivityData()

    return {
      type: 'bar' as ChartType,
      data: {
        labels: data.map((item: any) => item.period || item.month),
        datasets: [
          {
            label: 'Buy Orders',
            data: data.map((item: any) => item.buyOrders),
            backgroundColor: '#4caf50',
            borderRadius: 4
          },
          {
            label: 'Sell Orders',
            data: data.map((item: any) => item.sellOrders),
            backgroundColor: '#f44336',
            borderRadius: 4
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Trading Activity',
            font: { size: 16, weight: 'bold' }
          },
          legend: {
            display: true,
            position: 'top'
          }
        },
        scales: {
          x: {
            title: { display: true, text: 'Month' }
          },
          y: {
            title: { display: true, text: 'Number of Orders' },
            beginAtZero: true
          }
        }
      }
    }
  }

  private generatePerformanceData () {
    const data = []
    let portfolioValue = 100000
    let profitLoss = 0

    for (let i = 0; i < 30; i++) {
      const change = (Math.random() - 0.5) * 5000
      portfolioValue += change
      profitLoss += change

      data.push({
        portfolioValue: Math.max(50000, portfolioValue),
        profitLoss: profitLoss,
        date: new Date(Date.now() - (29 - i) * 24 * 60 * 60 * 1000)
      })
    }

    return data
  }

  private generatePortfolioData () {
    return [
      { stock: 'Apple', value: 25000 },
      { stock: 'Microsoft', value: 20000 },
      { stock: 'Tesla', value: 15000 },
      { stock: 'Google', value: 18000 },
      { stock: 'Amazon', value: 12000 },
      { stock: 'Other', value: 10000 }
    ]
  }

  private generateActivityData () {
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun']
    return months.map(month => ({
      month,
      buyOrders: Math.floor(Math.random() * 50) + 10,
      sellOrders: Math.floor(Math.random() * 40) + 5
    }))
  }

  updateChartType (type: 'performance' | 'portfolio' | 'activity') {
    this.chartType = type
    if (this.chart) {
      this.chart.destroy()
      this.createChart()
    }
  }
}
