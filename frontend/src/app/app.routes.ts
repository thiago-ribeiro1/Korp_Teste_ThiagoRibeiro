import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'produtos' },
  {
    path: 'produtos',
    loadComponent: () =>
      import('./features/produtos/produtos-list.component').then((m) => m.ProdutosListComponent),
  },
  {
    path: 'notas-fiscais',
    loadComponent: () =>
      import('./features/notas/notas-list.component').then((m) => m.NotasListComponent),
  },
  {
    path: 'notas-fiscais/nova',
    loadComponent: () =>
      import('./features/notas/nota-form.component').then((m) => m.NotaFormComponent),
  },
  {
    path: 'notas-fiscais/:numeracao',
    loadComponent: () =>
      import('./features/notas/nota-detail.component').then((m) => m.NotaDetailComponent),
  },
  { path: '**', redirectTo: 'produtos' },
];
