import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { ItemEntrada, Nota } from '../models/nota.model';

@Injectable({ providedIn: 'root' })
export class NotasService {
  private readonly baseUrl = `${environment.faturamentoApiUrl}/notas`;

  constructor(private http: HttpClient) {}

  listar(status?: string, busca?: string): Observable<Nota[]> {
    const params = new URLSearchParams();
    if (status) params.set('status', status);
    if (busca) params.set('busca', busca);
    const query = params.toString();
    return this.http.get<Nota[]>(query ? `${this.baseUrl}?${query}` : this.baseUrl);
  }

  obter(numeracao: number): Observable<Nota> {
    return this.http.get<Nota>(`${this.baseUrl}/${numeracao}`);
  }

  criar(itens: ItemEntrada[]): Observable<Nota> {
    return this.http.post<Nota>(this.baseUrl, { itens });
  }

  atualizarItens(numeracao: number, itens: ItemEntrada[]): Observable<Nota> {
    return this.http.put<Nota>(`${this.baseUrl}/${numeracao}`, { itens });
  }

  imprimir(numeracao: number): Observable<Nota> {
    return this.http.post<Nota>(`${this.baseUrl}/${numeracao}/imprimir`, {});
  }
}
