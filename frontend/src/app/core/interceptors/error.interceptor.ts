import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http'
import { inject } from '@angular/core'
import { Router } from '@angular/router'
import { AuthService } from '../services/auth/auth.service'
import { catchError, throwError } from 'rxjs'

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const router = inject(Router)
  const authService = inject(AuthService)

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401) {
        authService.logout()
        router.navigate(['/login'])
      }

      const errorMessage =
        error.error?.message || error.statusText || 'An error occurred'
      console.error('HTTP Error:', errorMessage)

      return throwError(() => error)
    })
  )
}
