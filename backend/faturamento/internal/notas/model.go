package notas

import "time"

// Numeracao é a própria chave primária da nota (id), evitando manter um
// contador sequencial redundante.
type Nota struct {
	Numeracao int        `json:"numeracao"`
	Status    string     `json:"status"`
	CriadoEm  time.Time  `json:"criado_em"`
	FechadoEm *time.Time `json:"fechado_em,omitempty"`
	Itens     []ItemNota `json:"itens,omitempty"`

	QuantidadeItens int `json:"quantidade_itens"`
	QuantidadeTotal int `json:"quantidade_total"`
}

type ItemNota struct {
	ID               int    `json:"id"`
	ProdutoCodigo    string `json:"produto_codigo"`
	ProdutoDescricao string `json:"produto_descricao"`
	Quantidade       int    `json:"quantidade"`
}

// ItemEntrada é o item informado pelo cliente ao criar ou editar uma nota;
// a descrição é sempre resolvida no servidor a partir do código.
type ItemEntrada struct {
	Codigo     string `json:"codigo"`
	Quantidade int    `json:"quantidade"`
}
