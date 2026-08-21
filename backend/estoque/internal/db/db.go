package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// schema define a estrutura mínima necessária para o serviço de Estoque.
//
// Optamos por aplicar o schema diretamente no startup do serviço (via
// CREATE TABLE IF NOT EXISTS), em vez de usar uma ferramenta de migration
// (como golang-migrate), porque o escopo do teste é pequeno e uma única
// tabela não justifica essa complexidade adicional.
const schema = `
CREATE TABLE IF NOT EXISTS produtos (
    id SERIAL PRIMARY KEY,
    codigo TEXT NOT NULL UNIQUE,
    descricao TEXT NOT NULL,
    saldo INTEGER NOT NULL DEFAULT 0,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

// Connect abre um pool de conexões com o PostgreSQL usando pgxpool
// (em vez de uma conexão única) para suportar múltiplas requisições
// HTTP concorrentes de forma segura, e valida a conectividade com um
// Ping antes de retornar.
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

// Migrate garante que as tabelas necessárias existam, aplicando o schema
// definido acima. É idempotente: pode ser chamado toda vez que o serviço
// sobe, sem causar erro caso a tabela já exista.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("erro ao aplicar schema: %w", err)
	}
	return nil
}
