import { Routes } from '@angular/router'
import { authGuard } from './core/guards/auth.guard'

export const routes: Routes = [
  { path: '', redirectTo: '/stocks', pathMatch: 'full' },
  {
    path: 'login',
    loadComponent: () =>
      import('./features/login/login').then(m => m.LoginComponent)
  },
  {
    path: 'stocks',
    canActivate: [authGuard],
    children: [
      {
        path: '',
        loadComponent: () =>
          import('./features/stocks/stock-list/stock-list').then(
            m => m.StockListComponent
          )
      },
      {
        path: ':id',
        loadComponent: () =>
          import('./features/stocks/stock-detail/stock-detail').then(
            m => m.StockDetailComponent
          )
      }
    ]
  },
  {
    path: 'traders',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/traders/traders').then(m => m.TradersComponent)
  },
  {
    path: 'personal',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/personal/personal').then(m => m.PersonalComponent)
  }
]
