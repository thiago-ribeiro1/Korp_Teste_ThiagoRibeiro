export type StatusNota = 'Aberta' | 'Fechada';

export interface ItemNota {
  id: number;
  produto_codigo: string;
  produto_descricao: string;
  quantidade: number;
}

export interface Nota {
  numeracao: number;
  status: StatusNota;
  criado_em: string;
  fechado_em?: string;
  itens?: ItemNota[];
  quantidade_itens: number;
  quantidade_total: number;
}

export interface ItemEntrada {
  codigo: string;
  quantidade: number;
}
