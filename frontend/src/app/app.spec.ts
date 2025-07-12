import { TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { of, BehaviorSubject } from 'rxjs';
import { App } from './app';
import { AuthService } from './core/services/auth/auth.service';

describe('App', () => {
  let mockAuthService: any;
  let authSubject: BehaviorSubject<string | null>;

  beforeEach(async () => {
    authSubject = new BehaviorSubject<string | null>(null);
    
    mockAuthService = {
      logout: jasmine.createSpy('logout'),
      currentTraderId$: authSubject.asObservable()
    };

    await TestBed.configureTestingModule({
      imports: [
        App,
        RouterTestingModule,
        BrowserAnimationsModule
      ],
      providers: [
        { provide: AuthService, useValue: mockAuthService }
      ]
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(App);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });

  it('should render app container', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector('.app-container')).toBeTruthy();
  });

  it('should show toolbar when user is logged in', () => {
    // Set logged in state before creating component
    authSubject.next('trader123');
    
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    
    expect(compiled.querySelector('mat-toolbar')).toBeTruthy();
    expect(compiled.querySelector('mat-toolbar span')?.textContent).toContain('Stock Exchange');
  });

  it('should hide toolbar when user is not logged in', () => {
    // Set logged out state before creating component
    authSubject.next(null);
    
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    
    expect(compiled.querySelector('mat-toolbar')).toBeFalsy();
  });

  it('should have router outlet', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector('router-outlet')).toBeTruthy();
  });
});
