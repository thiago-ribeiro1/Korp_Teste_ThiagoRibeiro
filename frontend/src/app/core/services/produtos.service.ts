import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Produto, ProdutoInput } from '../models/produto.model';

@Injectable({ providedIn: 'root' })
export class ProdutosService {
  private readonly baseUrl = `${environment.estoqueApiUrl}/produtos`;

  constructor(private http: HttpClient) {}

  listar(busca?: string): Observable<Produto[]> {
    const url = busca ? `${this.baseUrl}?busca=${encodeURIComponent(busca)}` : this.baseUrl;
    return this.http.get<Produto[]>(url);
  }

  obter(id: number): Observable<Produto> {
    return this.http.get<Produto>(`${this.baseUrl}/${id}`);
  }

  criar(produto: ProdutoInput): Observable<Produto> {
    return this.http.post<Produto>(this.baseUrl, produto);
  }

  atualizar(id: number, produto: ProdutoInput): Observable<Produto> {
    return this.http.put<Produto>(`${this.baseUrl}/${id}`, produto);
  }

  excluir(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
