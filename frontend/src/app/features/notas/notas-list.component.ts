import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
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

  constructor(
    private notasService: NotasService,
    private router: Router,
  ) {}

  ngOnInit(): void {
    this.carregar();
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
