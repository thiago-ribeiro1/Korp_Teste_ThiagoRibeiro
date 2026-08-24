package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// id da nota já serve de numeração, sem contador redundante
const schema = `
CREATE TABLE IF NOT EXISTS notas (
    id SERIAL PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'Aberta' CHECK (status IN ('Aberta', 'Fechada')),
    criado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
    fechado_em TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS itens_nota (
    id SERIAL PRIMARY KEY,
    nota_id INTEGER NOT NULL REFERENCES notas(id) ON DELETE CASCADE,
    produto_codigo TEXT NOT NULL,
    produto_descricao TEXT NOT NULL,
    quantidade INTEGER NOT NULL CHECK (quantidade > 0)
);
`

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar pool de conexão: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("erro ao conectar ao banco: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("erro ao aplicar schema: %w", err)
	}
	return nil
}
