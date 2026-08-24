package produtos

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("produto não encontrado")
	ErrCodigoDuplicado   = errors.New("código de produto já cadastrado")
	ErrSaldoInsuficiente = errors.New("saldo insuficiente")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, codigo, descricao string, saldo int) (Produto, error) {
	var p Produto
	err := r.pool.QueryRow(ctx,
		`INSERT INTO produtos (codigo, descricao, saldo) VALUES ($1, $2, $3)
         RETURNING id, codigo, descricao, saldo, criado_em`,
		codigo, descricao, saldo,
	).Scan(&p.ID, &p.Codigo, &p.Descricao, &p.Saldo, &p.CriadoEm)
	if err != nil {
		if strings.Contains(err.Error(), "produtos_codigo_key") {
			return Produto{}, ErrCodigoDuplicado
		}
		return Produto{}, fmt.Errorf("erro ao inserir produto: %w", err)
	}
	return p, nil
}

// filtra por código ou descrição, case-insensitive
func (r *Repository) List(ctx context.Context, busca string) ([]Produto, error) {
	query := `SELECT id, codigo, descricao, saldo, criado_em FROM produtos`
	args := []any{}
	if busca != "" {
		query += ` WHERE codigo ILIKE $1 OR descricao ILIKE $1`
		args = append(args, "%"+busca+"%")
	}
	query += ` ORDER BY codigo`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar produtos: %w", err)
	}
	defer rows.Close()

	var produtos []Produto
	for rows.Next() {
		var p Produto
		if err := rows.Scan(&p.ID, &p.Codigo, &p.Descricao, &p.Saldo, &p.CriadoEm); err != nil {
			return nil, fmt.Errorf("erro ao ler produto: %w", err)
		}
		produtos = append(produtos, p)
	}
	return produtos, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id int) (Produto, error) {
	var p Produto
	err := r.pool.QueryRow(ctx,
		`SELECT id, codigo, descricao, saldo, criado_em FROM produtos WHERE id = $1`, id,
	).Scan(&p.ID, &p.Codigo, &p.Descricao, &p.Saldo, &p.CriadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Produto{}, ErrNotFound
		}
		return Produto{}, fmt.Errorf("erro ao buscar produto: %w", err)
	}
	return p, nil
}

func (r *Repository) Update(ctx context.Context, id int, codigo, descricao string, saldo int) (Produto, error) {
	var p Produto
	err := r.pool.QueryRow(ctx,
		`UPDATE produtos SET codigo = $1, descricao = $2, saldo = $3 WHERE id = $4
         RETURNING id, codigo, descricao, saldo, criado_em`,
		codigo, descricao, saldo, id,
	).Scan(&p.ID, &p.Codigo, &p.Descricao, &p.Saldo, &p.CriadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Produto{}, ErrNotFound
		}
		if strings.Contains(err.Error(), "produtos_codigo_key") {
			return Produto{}, ErrCodigoDuplicado
		}
		return Produto{}, fmt.Errorf("erro ao atualizar produto: %w", err)
	}
	return p, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM produtos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("erro ao excluir produto: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// tudo numa transação: se faltar saldo de um item, reverte a baixa inteira
func (r *Repository) BaixarSaldoLote(ctx context.Context, itens []ItemBaixa) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx) // no-op se o commit já tiver ocorrido

	for _, item := range itens {
		var saldoAtual int
		err := tx.QueryRow(ctx,
			`SELECT saldo FROM produtos WHERE codigo = $1 FOR UPDATE`, item.Codigo,
		).Scan(&saldoAtual)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: produto %s", ErrNotFound, item.Codigo)
			}
			return fmt.Errorf("erro ao consultar saldo do produto %s: %w", item.Codigo, err)
		}

		if saldoAtual < item.Quantidade {
			return fmt.Errorf("%w: produto %s (saldo atual: %d, solicitado: %d)",
				ErrSaldoInsuficiente, item.Codigo, saldoAtual, item.Quantidade)
		}

		if _, err := tx.Exec(ctx,
			`UPDATE produtos SET saldo = saldo - $1 WHERE codigo = $2`,
			item.Quantidade, item.Codigo,
		); err != nil {
			return fmt.Errorf("erro ao atualizar saldo do produto %s: %w", item.Codigo, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}
	return nil
}
