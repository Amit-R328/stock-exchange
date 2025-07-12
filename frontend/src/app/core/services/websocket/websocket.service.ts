import { Injectable, NgZone } from '@angular/core'
import { Observable, Subject, BehaviorSubject } from 'rxjs'
import { environment } from '../../../../environments/environment'

@Injectable({
  providedIn: 'root'
})
export class WebSocketService {
  private socket!: WebSocket
  private subject = new Subject<any>()
  private connectionStatus$ = new BehaviorSubject<string>('disconnected')
  private reconnectInterval = 3000
  private reconnectTimer: any
  private shouldReconnect = true

  constructor (private ngZone: NgZone) {}

  connect (): Observable<any> {
    this.shouldReconnect = true
    this.createConnection()
    return this.subject.asObservable()
  }

  private createConnection (): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      return
    }

    console.log('🔌 Creating WebSocket connection...')
    this.socket = new WebSocket(environment.wsUrl)

    this.socket.onopen = () => {
      this.ngZone.run(() => {
        console.log('✅ WebSocket connected')
        this.connectionStatus$.next('connected')

        // Clear any pending reconnection
        if (this.reconnectTimer) {
          clearTimeout(this.reconnectTimer)
          this.reconnectTimer = null
        }

        // Notify about reconnection
        this.subject.next({ type: 'connection', status: 'connected' })
      })
    }

    this.socket.onmessage = event => {
      this.ngZone.run(() => {
        this.subject.next(JSON.parse(event.data))
      })
    }

    this.socket.onerror = error => {
      this.ngZone.run(() => {
        console.error('❌ WebSocket error:', error)
        this.connectionStatus$.next('error')
      })
    }

    this.socket.onclose = event => {
      this.ngZone.run(() => {
        console.log('🔌 WebSocket closed:', event.code)
        this.connectionStatus$.next('disconnected')

        // Only reconnect if I should and aren't already trying
        if (this.shouldReconnect && !this.reconnectTimer) {
          console.log(`🔄 Reconnecting in ${this.reconnectInterval}ms...`)
          this.reconnectTimer = setTimeout(() => {
            this.reconnectTimer = null
            if (this.shouldReconnect) {
              this.createConnection()
            }
          }, this.reconnectInterval)
        }
      })
    }
  }

  disconnect (): void {
    console.log('🛑 Disconnecting WebSocket')
    this.shouldReconnect = false

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    if (this.socket) {
      this.socket.close()
    }

    this.connectionStatus$.next('disconnected')
  }

  getConnectionStatus (): Observable<string> {
    return this.connectionStatus$.asObservable()
  }
}
