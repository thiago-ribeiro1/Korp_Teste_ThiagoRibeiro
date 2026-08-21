package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
	Port        string
}

// Load lê variáveis de ambiente do processo e, se existir, de um arquivo
// .env na raiz do serviço. Não há senha padrão embutida: ESTOQUE_DB_URL é
// obrigatória para evitar credenciais hardcoded no código-fonte.
func Load() Config {
	loadDotEnv(".env")

	dbURL := os.Getenv("ESTOQUE_DB_URL")
	if dbURL == "" {
		log.Fatal("variável de ambiente ESTOQUE_DB_URL não definida (configure um arquivo .env a partir do .env.example)")
	}

	return Config{
		DatabaseURL: dbURL,
		Port:        getEnv("ESTOQUE_PORT", "8081"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv aplica pares KEY=VALUE de um arquivo .env como variáveis de
// ambiente do processo, sem sobrescrever variáveis já definidas no shell.
// Implementação simples e sem dependências externas, suficiente para o
// escopo do teste.
func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}
