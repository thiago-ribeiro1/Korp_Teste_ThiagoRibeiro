package notas

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound    = errors.New("nota fiscal não encontrada")
	ErrNotaFechada = errors.New("nota fiscal está fechada")
	ErrItensVazios = errors.New("a nota precisa ter ao menos um produto")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// nota e itens na mesma transação: se um item falhar, não cria nada
func (r *Repository) Create(ctx context.Context, itens []ItemNota) (Nota, error) {
	if len(itens) == 0 {
		return Nota{}, ErrItensVazios
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Nota{}, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	var nota Nota
	err = tx.QueryRow(ctx,
		`INSERT INTO notas DEFAULT VALUES RETURNING id, status, criado_em, fechado_em`,
	).Scan(&nota.Numeracao, &nota.Status, &nota.CriadoEm, &nota.FechadoEm)
	if err != nil {
		return Nota{}, fmt.Errorf("erro ao criar nota: %w", err)
	}

	for _, item := range itens {
		if _, err := tx.Exec(ctx,
			`INSERT INTO itens_nota (nota_id, produto_codigo, produto_descricao, quantidade)
             VALUES ($1, $2, $3, $4)`,
			nota.Numeracao, item.ProdutoCodigo, item.ProdutoDescricao, item.Quantidade,
		); err != nil {
			return Nota{}, fmt.Errorf("erro ao inserir item da nota: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Nota{}, fmt.Errorf("erro ao confirmar criação da nota: %w", err)
	}

	nota.Itens = itens
	nota.QuantidadeItens = len(itens)
	for _, item := range itens {
		nota.QuantidadeTotal += item.Quantidade
	}

	return nota, nil
}

// traz só os contadores de itens/quantidade, sem carregar os itens completos
func (r *Repository) List(ctx context.Context, status, busca string) ([]Nota, error) {
	query := `
        SELECT n.id, n.status, n.criado_em, n.fechado_em,
               COUNT(i.id) AS qtd_itens,
               COALESCE(SUM(i.quantidade), 0) AS qtd_total
        FROM notas n
        LEFT JOIN itens_nota i ON i.nota_id = n.id
        WHERE ($1 = '' OR n.status = $1)
          AND ($2 = '' OR LPAD(n.id::text, 6, '0') ILIKE '%' || $2 || '%')
        GROUP BY n.id
        ORDER BY n.id DESC`

	rows, err := r.pool.Query(ctx, query, status, busca)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar notas: %w", err)
	}
	defer rows.Close()

	var lista []Nota
	for rows.Next() {
		var n Nota
		if err := rows.Scan(&n.Numeracao, &n.Status, &n.CriadoEm, &n.FechadoEm, &n.QuantidadeItens, &n.QuantidadeTotal); err != nil {
			return nil, fmt.Errorf("erro ao ler nota: %w", err)
		}
		lista = append(lista, n)
	}
	return lista, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id int) (Nota, error) {
	var n Nota
	err := r.pool.QueryRow(ctx,
		`SELECT id, status, criado_em, fechado_em FROM notas WHERE id = $1`, id,
	).Scan(&n.Numeracao, &n.Status, &n.CriadoEm, &n.FechadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Nota{}, ErrNotFound
		}
		return Nota{}, fmt.Errorf("erro ao buscar nota: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, produto_codigo, produto_descricao, quantidade FROM itens_nota WHERE nota_id = $1 ORDER BY id`, id,
	)
	if err != nil {
		return Nota{}, fmt.Errorf("erro ao buscar itens da nota: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item ItemNota
		if err := rows.Scan(&item.ID, &item.ProdutoCodigo, &item.ProdutoDescricao, &item.Quantidade); err != nil {
			return Nota{}, fmt.Errorf("erro ao ler item da nota: %w", err)
		}
		n.Itens = append(n.Itens, item)
		n.QuantidadeTotal += item.Quantidade
	}
	n.QuantidadeItens = len(n.Itens)

	return n, rows.Err()
}

// só permite trocar os itens se a nota ainda está Aberta
func (r *Repository) UpdateItens(ctx context.Context, id int, itens []ItemNota) (Nota, error) {
	if len(itens) == 0 {
		return Nota{}, ErrItensVazios
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Nota{}, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM notas WHERE id = $1 FOR UPDATE`, id).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Nota{}, ErrNotFound
		}
		return Nota{}, fmt.Errorf("erro ao verificar status da nota: %w", err)
	}
	if status != "Aberta" {
		return Nota{}, ErrNotaFechada
	}

	if _, err := tx.Exec(ctx, `DELETE FROM itens_nota WHERE nota_id = $1`, id); err != nil {
		return Nota{}, fmt.Errorf("erro ao remover itens antigos: %w", err)
	}

	for _, item := range itens {
		if _, err := tx.Exec(ctx,
			`INSERT INTO itens_nota (nota_id, produto_codigo, produto_descricao, quantidade)
             VALUES ($1, $2, $3, $4)`,
			id, item.ProdutoCodigo, item.ProdutoDescricao, item.Quantidade,
		); err != nil {
			return Nota{}, fmt.Errorf("erro ao inserir item da nota: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Nota{}, fmt.Errorf("erro ao confirmar atualização da nota: %w", err)
	}

	return r.GetByID(ctx, id)
}

// condição de status no próprio UPDATE evita fechar a nota duas vezes em impressões concorrentes
func (r *Repository) FecharSeAberta(ctx context.Context, id int) (Nota, error) {
	var n Nota
	err := r.pool.QueryRow(ctx,
		`UPDATE notas SET status = 'Fechada', fechado_em = now()
         WHERE id = $1 AND status = 'Aberta'
         RETURNING id, status, criado_em, fechado_em`,
		id,
	).Scan(&n.Numeracao, &n.Status, &n.CriadoEm, &n.FechadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Nota{}, ErrNotaFechada
		}
		return Nota{}, fmt.Errorf("erro ao fechar nota: %w", err)
	}
	return n, nil
}
