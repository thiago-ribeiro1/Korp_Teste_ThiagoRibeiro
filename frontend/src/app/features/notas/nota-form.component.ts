import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { ProdutosService } from '../../core/services/produtos.service';
import { NotasService } from '../../core/services/notas.service';
import { Produto } from '../../core/models/produto.model';
import { ItemEntrada } from '../../core/models/nota.model';

interface ItemRascunho extends ItemEntrada {
  descricao: string;
  saldoAtual: number;
}

@Component({
  selector: 'app-nota-form',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './nota-form.component.html',
})
export class NotaFormComponent implements OnInit {
  produtos: Produto[] = [];
  itens: ItemRascunho[] = [];

  produtoSelecionadoId: number | null = null;
  quantidadeSelecionada = 1;

  salvando = false;
  erro = '';

  constructor(
    private produtosService: ProdutosService,
    private notasService: NotasService,
    private router: Router,
  ) {}

  ngOnInit(): void {
    this.produtosService.listar().subscribe({
      next: (produtos) => (this.produtos = produtos),
      error: () => (this.erro = 'Não foi possível carregar os produtos do estoque.'),
    });
  }

  get quantidadeTotal(): number {
    return this.itens.reduce((soma, item) => soma + item.quantidade, 0);
  }

  adicionarItem(): void {
    this.erro = '';
    const produto = this.produtos.find((p) => p.id === this.produtoSelecionadoId);
    if (!produto) {
      this.erro = 'Selecione um produto.';
      return;
    }
    if (!this.quantidadeSelecionada || this.quantidadeSelecionada <= 0) {
      this.erro = 'Informe uma quantidade maior que zero.';
      return;
    }

    const existente = this.itens.find((i) => i.codigo === produto.codigo);
    if (existente) {
      existente.quantidade += this.quantidadeSelecionada;
    } else {
      this.itens.push({
        codigo: produto.codigo,
        descricao: produto.descricao,
        saldoAtual: produto.saldo,
        quantidade: this.quantidadeSelecionada,
      });
    }

    this.produtoSelecionadoId = null;
    this.quantidadeSelecionada = 1;
  }

  removerItem(codigo: string): void {
    this.itens = this.itens.filter((i) => i.codigo !== codigo);
  }

  salvar(): void {
    if (this.itens.length === 0) {
      this.erro = 'Adicione ao menos um produto à nota.';
      return;
    }

    this.salvando = true;
    this.erro = '';

    const payload: ItemEntrada[] = this.itens.map((i) => ({ codigo: i.codigo, quantidade: i.quantidade }));

    this.notasService.criar(payload).subscribe({
      next: (nota) => {
        this.salvando = false;
        this.router.navigate(['/notas-fiscais', nota.numeracao]);
      },
      error: (err) => {
        this.salvando = false;
        this.erro = err?.error?.erro || 'Erro ao salvar a nota fiscal.';
      },
    });
  }
}
