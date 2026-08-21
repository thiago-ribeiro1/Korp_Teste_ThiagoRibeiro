package produtos

import "time"

// Produto representa um item cadastrado no serviço de Estoque.
type Produto struct {
	ID        int       `json:"id"`
	Codigo    string    `json:"codigo"`
	Descricao string    `json:"descricao"`
	Saldo     int       `json:"saldo"`
	CriadoEm  time.Time `json:"criado_em"`
}

// ItemBaixa representa a quantidade a ser reduzida do saldo de um produto,
// usado na baixa em lote solicitada pelo serviço de Faturamento.
type ItemBaixa struct {
	Codigo     string `json:"codigo"`
	Quantidade int    `json:"quantidade"`
}