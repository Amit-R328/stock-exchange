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
import { StockService } from '../../../core/services/stock/stock.service'

// Register Chart.js components
Chart.register(...registerables)

@Component({
  selector: 'app-stock-chart',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './stock-chart.component.html',
  styleUrls: ['./stock-chart.component.scss']
})
export class StockChartComponent implements OnInit, OnDestroy {
  @Input() stockData: any[] = []
  @Input() stockName: string = ''
  @Input() stockId: string = ''
  @Input() chartType: 'line' | 'candlestick' | 'volume' = 'line'
  @Input() height: number = 300

  @ViewChild('chartCanvas', { static: true })
  chartCanvas!: ElementRef<HTMLCanvasElement>

  private chart?: Chart

  constructor (private stockService: StockService) {}

  ngOnInit () {
    this.loadChartData()
  }

  ngOnDestroy () {
    if (this.chart) {
      this.chart.destroy()
    }
  }

  private loadChartData () {
    if (this.stockId) {
      this.stockService.getStockHistory(this.stockId, 30).subscribe({
        next: response => {
          this.stockData = response.history
          this.createChart()
        },
        error: error => {
          console.error('Error loading stock history:', error)
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

    // Generate sample data if none provided
    const data =
      this.stockData.length > 0 ? this.stockData : this.generateSampleData()

    let config: ChartConfiguration

    switch (this.chartType) {
      case 'volume':
        config = this.createVolumeChart(data)
        break
      case 'candlestick':
        config = this.createCandlestickChart(data)
        break
      case 'line':
      default:
        config = this.createLineChart(data)
        break
    }

    this.chart = new Chart(ctx, config)
  }

  private createLineChart(data: any[]): ChartConfiguration {
    return {
      type: 'line' as ChartType,
      data: {
        labels: data.map((item: any, index: number) => {
          if (item.timestamp) {
            const date = new Date(item.timestamp)
            return date.toLocaleDateString()
          }
          return `Day ${index + 1}`
        }),
        datasets: [
          {
            label: `${this.stockName} Price`,
            data: data.map((item: any) => item.price),
            borderColor: '#1976d2',
            backgroundColor: 'rgba(25, 118, 210, 0.1)',
            borderWidth: 2,
            fill: true,
            tension: 0.1,
            pointBackgroundColor: '#1976d2',
            pointBorderColor: '#fff',
            pointBorderWidth: 2,
            pointRadius: 4,
            pointHoverRadius: 6
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: `${this.stockName} Price Chart`,
            font: {
              size: 16,
              weight: 'bold'
            }
          },
          legend: {
            display: false
          },
          tooltip: {
            mode: 'index',
            intersect: false,
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            titleColor: '#fff',
            bodyColor: '#fff',
            borderColor: '#1976d2',
            borderWidth: 1,
            callbacks: {
              label: context => {
                const value = context.parsed.y
                return `Price: $${value.toFixed(2)}`
              }
            }
          }
        },
        scales: {
          x: {
            display: true,
            title: {
              display: true,
              text: 'Time Period'
            },
            grid: {
              color: 'rgba(0, 0, 0, 0.1)'
            }
          },
          y: {
            display: true,
            title: {
              display: true,
              text: 'Price ($)'
            },
            grid: {
              color: 'rgba(0, 0, 0, 0.1)'
            },
            ticks: {
              callback: function (value) {
                return '$' + value
              }
            }
          }
        },
        interaction: {
          mode: 'nearest',
          axis: 'x',
          intersect: false
        }
      }
    }
  }

  private createVolumeChart(data: any[]): ChartConfiguration {
    return {
      type: 'bar' as ChartType,
      data: {
        labels: data.map((item: any, index: number) => {
          if (item.timestamp) {
            const date = new Date(item.timestamp)
            return date.toLocaleDateString()
          }
          return `Day ${index + 1}`
        }),
        datasets: [
          {
            label: `${this.stockName} Volume`,
            data: data.map((item: any) => item.volume || Math.floor(Math.random() * 1000000)),
            backgroundColor: 'rgba(76, 175, 80, 0.6)',
            borderColor: '#4caf50',
            borderWidth: 1,
            borderRadius: 2
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: `${this.stockName} Trading Volume`,
            font: {
              size: 16,
              weight: 'bold'
            }
          },
          legend: {
            display: false
          },
          tooltip: {
            mode: 'index',
            intersect: false,
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            titleColor: '#fff',
            bodyColor: '#fff',
            borderColor: '#4caf50',
            borderWidth: 1,
            callbacks: {
              label: context => {
                const value = context.parsed.y
                return `Volume: ${value.toLocaleString()} shares`
              }
            }
          }
        },
        scales: {
          x: {
            display: true,
            title: {
              display: true,
              text: 'Time Period'
            },
            grid: {
              color: 'rgba(0, 0, 0, 0.1)'
            }
          },
          y: {
            display: true,
            title: {
              display: true,
              text: 'Volume (shares)'
            },
            grid: {
              color: 'rgba(0, 0, 0, 0.1)'
            },
            ticks: {
              callback: function (value) {
                if (typeof value === 'number') {
                  return value.toLocaleString()
                }
                return value
              }
            }
          }
        },
        interaction: {
          mode: 'nearest',
          axis: 'x',
          intersect: false
        }
      }
    }
  }

  private createCandlestickChart(data: any[]): ChartConfiguration {
    // Create a simplified candlestick-style chart using overlapping area charts
    // to show high/low price ranges
    const priceData = data.map((item: any) => item.price)
    const highData = data.map((item: any) => {
      const price = item.price
      return price + (price * 0.02) // 2% higher
    })
    const lowData = data.map((item: any) => {
      const price = item.price
      return price - (price * 0.02) // 2% lower
    })

    return {
      type: 'line' as ChartType,
      data: {
        labels: data.map((item: any, index: number) => {
          if (item.timestamp) {
            const date = new Date(item.timestamp)
            return date.toLocaleDateString()
          }
          return `Day ${index + 1}`
        }),
        datasets: [
          {
            label: `${this.stockName} High`,
            data: highData,
            borderColor: '#4caf50',
            backgroundColor: 'rgba(76, 175, 80, 0.1)',
            borderWidth: 1,
            fill: '+1',
            tension: 0.1,
            pointRadius: 0,
            pointHoverRadius: 4
          },
          {
            label: `${this.stockName} Low`,
            data: lowData,
            borderColor: '#f44336',
            backgroundColor: 'rgba(244, 67, 54, 0.1)',
            borderWidth: 1,
            fill: false,
            tension: 0.1,
            pointRadius: 0,
            pointHoverRadius: 4
          },
          {
            label: `${this.stockName} Price`,
            data: priceData,
            borderColor: '#ff9800',
            backgroundColor: 'transparent',
            borderWidth: 3,
            fill: false,
            tension: 0.1,
            pointBackgroundColor: '#ff9800',
            pointBorderColor: '#fff',
            pointBorderWidth: 2,
            pointRadius: 4,
            pointHoverRadius: 6
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: `${this.stockName} Price Range`,
            font: {
              size: 16,
              weight: 'bold'
            }
          },
          legend: {
            display: true,
            position: 'top'
          },
          tooltip: {
            mode: 'index',
            intersect: false,
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            titleColor: '#fff',
            bodyColor: '#fff',
            borderColor: '#ff9800',
            borderWidth: 1,
            callbacks: {
              label: context => {
                const value = context.parsed.y
                const label = context.dataset.label
                if (label?.includes('Price')) {
                  return `${label}: $${value.toFixed(2)}`
                } else if (label?.includes('High')) {
                  return `${label}: $${value.toFixed(2)}`
                } else if (label?.includes('Low')) {
                  return `${label}: $${value.toFixed(2)}`
                }
                return `$${value.toFixed(2)}`
              }
            }
          }
        },
        scales: {
          x: {
            display: true,
            title: {
              display: true,
              text: 'Time Period'
            },
            grid: {
              color: 'rgba(0, 0, 0, 0.1)'
            }
          },
          y: {
            display: true,
            title: {
              display: true,
              text: 'Price ($)'
            },
            grid: {
              color: 'rgba(0, 0, 0, 0.1)'
            },
            ticks: {
              callback: function (value) {
                return '$' + value
              }
            }
          }
        },
        interaction: {
          mode: 'nearest',
          axis: 'x',
          intersect: false
        }
      }
    }
  }

  private generateSampleData () {
    const data = []
    let basePrice = 100 + Math.random() * 200

    for (let i = 0; i < 30; i++) {
      // Simulate price movement
      const change = (Math.random() - 0.5) * 10
      basePrice = Math.max(10, basePrice + change)

      data.push({
        price: parseFloat(basePrice.toFixed(2)),
        volume: Math.floor(Math.random() * 1000000),
        timestamp: new Date(Date.now() - (29 - i) * 24 * 60 * 60 * 1000)
      })
    }

    return data
  }

  updateChartType (type: 'line' | 'candlestick' | 'volume') {
    this.chartType = type
    if (this.chart) {
      this.chart.destroy()
      this.createChart()
    }
  }
}
