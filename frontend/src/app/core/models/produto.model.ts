export interface Produto {
  id: number;
  codigo: string;
  descricao: string;
  saldo: number;
  criado_em: string;
}

export interface ProdutoInput {
  codigo: string;
  descricao: string;
  saldo: number;
}
