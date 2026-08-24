package estoqueclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// timeout curto pra não travar a impressão se o Estoque cair
type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

type Produto struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	Saldo     int    `json:"saldo"`
}

// falha de comunicação (timeout, conexão recusada), diferente de erro de negócio
type ErrIndisponivel struct{ causa error }

func (e *ErrIndisponivel) Error() string {
	return fmt.Sprintf("serviço de estoque indisponível: %v", e.causa)
}
func (e *ErrIndisponivel) Unwrap() error { return e.causa }

// erro de negócio do Estoque (produto não encontrado, saldo insuficiente etc.)
type ErrNegocio struct {
	Status   int
	Mensagem string
}

func (e *ErrNegocio) Error() string { return e.Mensagem }

// reaproveita o endpoint de listagem com filtro, não existe busca por código exato
func (c *Client) ObterProdutoPorCodigo(ctx context.Context, codigo string) (Produto, error) {
	url := fmt.Sprintf("%s/produtos?busca=%s", c.baseURL, codigo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Produto{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return Produto{}, &ErrIndisponivel{causa: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Produto{}, &ErrNegocio{Status: resp.StatusCode, Mensagem: "erro ao consultar produto no estoque"}
	}

	var produtos []Produto
	if err := json.NewDecoder(resp.Body).Decode(&produtos); err != nil {
		return Produto{}, fmt.Errorf("erro ao ler resposta do estoque: %w", err)
	}

	for _, p := range produtos {
		if p.Codigo == codigo {
			return p, nil
		}
	}

	return Produto{}, &ErrNegocio{Status: http.StatusNotFound, Mensagem: fmt.Sprintf("produto %s não encontrado no estoque", codigo)}
}

type ItemBaixa struct {
	Codigo     string `json:"codigo"`
	Quantidade int    `json:"quantidade"`
}

// timeout vira ErrIndisponivel, erro de negócio (404/409) vira ErrNegocio
func (c *Client) BaixarLote(ctx context.Context, itens []ItemBaixa) error {
	body, err := json.Marshal(map[string]any{"itens": itens})
	if err != nil {
		return err
	}

	url := c.baseURL + "/produtos/baixa"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return &ErrIndisponivel{causa: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	var payload struct {
		Erro string `json:"erro"`
	}
	json.NewDecoder(resp.Body).Decode(&payload)
	mensagem := payload.Erro
	if mensagem == "" {
		mensagem = "erro ao processar baixa de saldo no estoque"
	}

	return &ErrNegocio{Status: resp.StatusCode, Mensagem: mensagem}
}
