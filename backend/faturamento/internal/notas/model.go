package notas

import "time"

// numeração é o próprio id, sem contador redundante
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

// descrição sempre resolvida no servidor, nunca confia no que o cliente manda
type ItemEntrada struct {
	Codigo     string `json:"codigo"`
	Quantidade int    `json:"quantidade"`
}
