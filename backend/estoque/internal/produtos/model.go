package produtos

import "time"

type Produto struct {
	ID        int       `json:"id"`
	Codigo    string    `json:"codigo"`
	Descricao string    `json:"descricao"`
	Saldo     int       `json:"saldo"`
	CriadoEm  time.Time `json:"criado_em"`
}

// usado na baixa em lote pedida pelo Faturamento
type ItemBaixa struct {
	Codigo     string `json:"codigo"`
	Quantidade int    `json:"quantidade"`
}
