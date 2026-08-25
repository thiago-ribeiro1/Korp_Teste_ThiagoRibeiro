import { CommonModule } from '@angular/common';
import { Component, DestroyRef, OnInit, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { Subject, debounceTime, distinctUntilChanged, switchMap } from 'rxjs';
import { NotasService } from '../../core/services/notas.service';
import { Nota } from '../../core/models/nota.model';

@Component({
  selector: 'app-notas-list',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './notas-list.component.html',
})
export class NotasListComponent implements OnInit {
  notas: Nota[] = [];
  filtroStatus: '' | 'Aberta' | 'Fechada' = '';
  busca = '';
  carregando = false;
  erro = '';

  private destroyRef = inject(DestroyRef);
  private buscaChange = new Subject<string>();

  constructor(
    private notasService: NotasService,
    private router: Router,
  ) {}

  ngOnInit(): void {
    this.carregar();

    // Busca em tempo real, com debounce e cancelamento da requisição
    // anterior via switchMap, seguindo o mesmo padrão da tela de produtos.
    this.buscaChange
      .pipe(
        debounceTime(300),
        distinctUntilChanged(),
        switchMap((busca) => {
          this.carregando = true;
          this.erro = '';
          return this.notasService.listar(this.filtroStatus || undefined, busca || undefined);
        }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: (notas) => {
          this.notas = notas;
          this.carregando = false;
        },
        error: () => {
          this.erro = 'Não foi possível carregar as notas fiscais. Verifique se o serviço de faturamento está disponível.';
          this.carregando = false;
        },
      });
  }

  onBuscaChange(valor: string): void {
    this.busca = valor;
    this.buscaChange.next(valor);
  }

  carregar(): void {
    this.carregando = true;
    this.erro = '';
    this.notasService.listar(this.filtroStatus || undefined, this.busca || undefined).subscribe({
      next: (notas) => {
        this.notas = notas;
        this.carregando = false;
      },
      error: () => {
        this.erro = 'Não foi possível carregar as notas fiscais. Verifique se o serviço de faturamento está disponível.';
        this.carregando = false;
      },
    });
  }

  filtrarPorStatus(status: '' | 'Aberta' | 'Fechada'): void {
    this.filtroStatus = status;
    this.carregar();
  }

  abrir(numeracao: number): void {
    this.router.navigate(['/notas-fiscais', numeracao]);
  }

  numeracaoFormatada(numeracao: number): string {
    return numeracao.toString().padStart(6, '0');
  }
}
