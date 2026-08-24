import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { catchError, forkJoin, map, of } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface StatusServicos {
  estoque: boolean;
  faturamento: boolean;
}

// chama os dois /health direto do front, não tem endpoint agregador no backend
@Injectable({ providedIn: 'root' })
export class HealthService {
  constructor(private http: HttpClient) {}

  verificar() {
    const estoque$ = this.http.get(`${environment.estoqueApiUrl}/health`).pipe(
      map(() => true),
      catchError(() => of(false)),
    );
    const faturamento$ = this.http.get(`${environment.faturamentoApiUrl}/health`).pipe(
      map(() => true),
      catchError(() => of(false)),
    );

    return forkJoin({ estoque: estoque$, faturamento: faturamento$ });
  }
}
