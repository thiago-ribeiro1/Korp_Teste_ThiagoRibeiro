import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { NotasService } from '../../core/services/notas.service';
import { HealthService, StatusServicos } from '../../core/services/health.service';
import { Nota } from '../../core/models/nota.model';

@Component({
  selector: 'app-nota-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './nota-detail.component.html',
})
export class NotaDetailComponent implements OnInit {
  nota: Nota | null = null;
  carregando = true;
  erroCarregamento = '';

  imprimindo = false;
  erroImpressao = '';
  mensagemSucesso = '';
  status: StatusServicos = { estoque: true, faturamento: true };

  constructor(
    private route: ActivatedRoute,
    private notasService: NotasService,
    private health: HealthService,
  ) {}

  ngOnInit(): void {
    this.carregar();
  }

  private get numeracao(): number {
    return Number(this.route.snapshot.paramMap.get('numeracao'));
  }

  carregar(): void {
    this.carregando = true;
    this.erroCarregamento = '';
    this.notasService.obter(this.numeracao).subscribe({
      next: (nota) => {
        this.nota = nota;
        this.carregando = false;
      },
      error: () => {
        this.erroCarregamento = 'Nota fiscal não encontrada.';
        this.carregando = false;
      },
    });
  }

  numeracaoFormatada(numeracao: number): string {
    return numeracao.toString().padStart(6, '0');
  }

  // Ao imprimir, o indicador de processamento fica visível durante a
  // chamada; se o estoque falhar, a nota permanece Aberta e a UI oferece
  // "tentar novamente" sem exigir recarregar a página.
  imprimir(): void {
    if (!this.nota || this.nota.status !== 'Aberta') return;

    this.imprimindo = true;
    this.erroImpressao = '';
    this.mensagemSucesso = '';

    this.notasService.imprimir(this.nota.numeracao).subscribe({
      next: (notaAtualizada) => {
        this.imprimindo = false;
        this.mensagemSucesso = `Impressão concluída. A nota ${this.numeracaoFormatada(notaAtualizada.numeracao)} foi fechada e o saldo dos produtos utilizados foi atualizado.`;
        this.carregar();
      },
      error: (err: HttpErrorResponse) => {
        this.imprimindo = false;
        this.erroImpressao = err?.error?.erro || 'Não foi possível concluir a impressão.';
        this.health.verificar().subscribe((status) => (this.status = status));
      },
    });
  }
}
