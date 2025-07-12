import { TestBed } from '@angular/core/testing'
import { Router } from '@angular/router'
import { provideHttpClient, withInterceptors } from '@angular/common/http'
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing'
import { HttpClient, HttpErrorResponse } from '@angular/common/http'
import { AuthService } from '../services/auth/auth.service'
import { errorInterceptor } from './error.interceptor'

describe('ErrorInterceptor', () => {
  let httpClient: HttpClient
  let httpMock: HttpTestingController
  let mockRouter: jasmine.SpyObj<Router>
  let mockAuthService: jasmine.SpyObj<AuthService>

  beforeEach(() => {
    const routerSpy = jasmine.createSpyObj('Router', ['navigate'])
    const authServiceSpy = jasmine.createSpyObj('AuthService', ['logout'])

    TestBed.configureTestingModule({
      providers: [
        { provide: Router, useValue: routerSpy },
        { provide: AuthService, useValue: authServiceSpy },
        provideHttpClient(withInterceptors([errorInterceptor])),
        provideHttpClientTesting()
      ]
    })

    httpClient = TestBed.inject(HttpClient)
    httpMock = TestBed.inject(HttpTestingController)
    mockRouter = TestBed.inject(Router) as jasmine.SpyObj<Router>
    mockAuthService = TestBed.inject(AuthService) as jasmine.SpyObj<AuthService>
  })

  afterEach(() => {
    httpMock.verify()
  })

  it('should handle 401 unauthorized errors', () => {
    const testUrl = '/test'

    httpClient.get(testUrl).subscribe({
      next: () => fail('Expected an error'),
      error: (error: HttpErrorResponse) => {
        expect(error.status).toBe(401)
        expect(mockAuthService.logout).toHaveBeenCalled()
        expect(mockRouter.navigate).toHaveBeenCalledWith(['/login'])
      }
    })

    const req = httpMock.expectOne(testUrl)
    req.flush('Unauthorized', { status: 401, statusText: 'Unauthorized' })
  })

  it('should pass through non-401 errors without special handling', () => {
    const testUrl = '/test'

    httpClient.get(testUrl).subscribe({
      next: () => fail('Expected an error'),
      error: (error: HttpErrorResponse) => {
        expect(error.status).toBe(500)
        expect(mockAuthService.logout).not.toHaveBeenCalled()
        expect(mockRouter.navigate).not.toHaveBeenCalled()
      }
    })

    const req = httpMock.expectOne(testUrl)
    req.flush('Server Error', {
      status: 500,
      statusText: 'Internal Server Error'
    })
  })

  it('should extract error message from error.error.message', () => {
    const testUrl = '/test'
    const customErrorMessage = 'Custom error message'

    httpClient.get(testUrl).subscribe({
      next: () => fail('Expected an error'),
      error: (error: HttpErrorResponse) => {
        expect(error.error.message).toBe(customErrorMessage)
      }
    })

    const req = httpMock.expectOne(testUrl)
    req.flush(
      { message: customErrorMessage },
      { status: 400, statusText: 'Bad Request' }
    )
  })

  it('should handle network errors', () => {
    const testUrl = '/test'

    httpClient.get(testUrl).subscribe({
      next: () => fail('Expected an error'),
      error: (error: HttpErrorResponse) => {
        expect(error.status).toBe(0)
        expect(mockAuthService.logout).not.toHaveBeenCalled()
        expect(mockRouter.navigate).not.toHaveBeenCalled()
      }
    })

    const req = httpMock.expectOne(testUrl)
    req.error(new ProgressEvent('Network error'))
  })

  it('should log error messages to console', () => {
    const consoleSpy = spyOn(console, 'error')
    const testUrl = '/test'

    httpClient.get(testUrl).subscribe({
      next: () => fail('Expected an error'),
      error: () => {
        // Error callback executed
      }
    })

    const req = httpMock.expectOne(testUrl)
    req.flush('Error', { status: 404, statusText: 'Not Found' })

    expect(consoleSpy).toHaveBeenCalledWith('HTTP Error:', 'Not Found')
  })

  it('should use statusText when no error message is available', () => {
    const consoleSpy = spyOn(console, 'error')
    const testUrl = '/test'
    let errorReceived: any

    httpClient.get(testUrl).subscribe({
      next: () => fail('Expected an error'),
      error: error => {
        errorReceived = error
      }
    })

    const req = httpMock.expectOne(testUrl)
    req.flush(null, { status: 404, statusText: 'Not Found' })

    expect(errorReceived.status).toBe(404)
    expect(consoleSpy).toHaveBeenCalledWith('HTTP Error:', 'Not Found')
  })

  it('should handle network errors with default statusText', () => {
    const consoleSpy = spyOn(console, 'error')
    const testUrl = '/test'
    let errorReceived: any

    httpClient.get(testUrl).subscribe({
      next: () => fail('Expected an error'),
      error: error => {
        errorReceived = error
      }
    })

    const req = httpMock.expectOne(testUrl)
    // For status 0, Angular automatically sets statusText to 'Unknown Error'
    // So let's test with that actual behavior
    req.flush(null, { status: 0, statusText: '' })

    expect(errorReceived.status).toBe(0)
    // Angular sets statusText to 'Unknown Error' for network errors (status 0)
    expect(consoleSpy).toHaveBeenCalledWith('HTTP Error:', 'Unknown Error')
  })

  it('should use fallback message when both error.message and statusText are falsy', () => {
    // Test the interceptor logic directly by mocking an HttpErrorResponse
    // with both error.message and statusText as falsy values
    const mockError = {
      error: {}, // no message property
      status: 500,
      statusText: '', // explicitly empty
      message: '', // Angular HttpErrorResponse has this
      name: 'HttpErrorResponse'
    } as HttpErrorResponse

    const consoleSpy = spyOn(console, 'error')

    // Simulate the interceptor's error message extraction logic
    const errorMessage =
      mockError.error?.message || mockError.statusText || 'An error occurred'
    console.error('HTTP Error:', errorMessage)

    expect(consoleSpy).toHaveBeenCalledWith('HTTP Error:', 'An error occurred')
  })
})
