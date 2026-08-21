package config

import "os"

// Config concentra as variáveis de ambiente usadas pelo serviço de Estoque.
// Mantemos a leitura de configuração centralizada aqui para não espalhar
// chamadas a os.Getenv pelo resto do código.
type Config struct {
	// DatabaseURL é a string de conexão completa do PostgreSQL, no formato:
	// postgres://usuario:senha@host:porta/nome_do_banco?sslmode=disable
	DatabaseURL string

	// Port é a porta HTTP em que o serviço de Estoque vai escutar.
	Port string
}

// Load lê as variáveis de ambiente ESTOQUE_DB_URL e ESTOQUE_PORT.
// Caso não estejam definidas, aplica valores padrão adequados para
// desenvolvimento local, para que o serviço rode com "go run" sem
// exigir configuração adicional.
func Load() Config {
	return Config{
		DatabaseURL: getEnv(
			"ESTOQUE_DB_URL",
			"postgres://postgres:REDACTED@localhost:5432/estoque_db?sslmode=disable",
		),
		Port: getEnv("ESTOQUE_PORT", "8081"),
	}
}

// getEnv retorna o valor da variável de ambiente "key" se ela estiver
// definida e não vazia; caso contrário, retorna "fallback".
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}