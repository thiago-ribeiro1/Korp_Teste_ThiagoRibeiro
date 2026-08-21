import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { ProdutosService } from '../../core/services/produtos.service';
import { Produto, ProdutoInput } from '../../core/models/produto.model';

@Component({
  selector: 'app-produtos-list',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './produtos-list.component.html',
})
export class ProdutosListComponent implements OnInit {
  produtos: Produto[] = [];
  busca = '';
  carregando = false;
  erro = '';
  mensagemSucesso = '';

  modalAberto = false;
  produtoEmEdicao: Produto | null = null;
  form: ProdutoInput = { codigo: '', descricao: '', saldo: 0 };
  errosForm: Record<string, string> = {};
  salvando = false;
  erroModal = '';

  produtoParaExcluir: Produto | null = null;

  constructor(private produtosService: ProdutosService) {}

  ngOnInit(): void {
    this.carregar();
  }

  carregar(): void {
    this.carregando = true;
    this.erro = '';
    this.produtosService.listar(this.busca || undefined).subscribe({
      next: (produtos) => {
        this.produtos = produtos;
        this.carregando = false;
      },
      error: () => {
        this.erro = 'Não foi possível carregar os produtos. Verifique se o serviço de estoque está disponível.';
        this.carregando = false;
      },
    });
  }

  abrirNovo(): void {
    this.produtoEmEdicao = null;
    this.form = { codigo: '', descricao: '', saldo: 0 };
    this.errosForm = {};
    this.erroModal = '';
    this.modalAberto = true;
  }

  abrirEdicao(produto: Produto): void {
    this.produtoEmEdicao = produto;
    this.form = { codigo: produto.codigo, descricao: produto.descricao, saldo: produto.saldo };
    this.errosForm = {};
    this.erroModal = '';
    this.modalAberto = true;
  }

  fecharModal(): void {
    this.modalAberto = false;
  }

  private validarForm(): boolean {
    this.errosForm = {};
    if (!this.form.codigo?.trim()) this.errosForm['codigo'] = 'Campo obrigatório.';
    if (!this.form.descricao?.trim()) this.errosForm['descricao'] = 'Campo obrigatório.';
    if (this.form.saldo === null || this.form.saldo === undefined || this.form.saldo < 0) {
      this.errosForm['saldo'] = 'Informe um saldo válido.';
    }
    return Object.keys(this.errosForm).length === 0;
  }

  salvar(): void {
    if (!this.validarForm()) return;

    this.salvando = true;
    this.erroModal = '';

    const request = this.produtoEmEdicao
      ? this.produtosService.atualizar(this.produtoEmEdicao.id, this.form)
      : this.produtosService.criar(this.form);

    request.subscribe({
      next: () => {
        this.salvando = false;
        this.modalAberto = false;
        this.mensagemSucesso = this.produtoEmEdicao
          ? `Produto ${this.form.codigo} atualizado com sucesso.`
          : `Produto ${this.form.codigo} cadastrado com sucesso.`;
        this.carregar();
      },
      error: (err) => {
        this.salvando = false;
        this.erroModal = err?.error?.erro || 'Erro ao salvar produto.';
      },
    });
  }

  confirmarExclusao(produto: Produto): void {
    this.produtoParaExcluir = produto;
  }

  cancelarExclusao(): void {
    this.produtoParaExcluir = null;
  }

  excluir(): void {
    if (!this.produtoParaExcluir) return;
    const codigo = this.produtoParaExcluir.codigo;
    this.produtosService.excluir(this.produtoParaExcluir.id).subscribe({
      next: () => {
        this.produtoParaExcluir = null;
        this.mensagemSucesso = `Produto ${codigo} excluído com sucesso.`;
        this.carregar();
      },
      error: (err) => {
        this.produtoParaExcluir = null;
        this.erro = err?.error?.erro || 'Erro ao excluir produto.';
      },
    });
  }
}
