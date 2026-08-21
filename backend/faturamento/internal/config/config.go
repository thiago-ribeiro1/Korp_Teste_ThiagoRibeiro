package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL    string
	Port           string
	EstoqueBaseURL string
}

// Load lê variáveis de ambiente do processo e, se existir, de um arquivo
// .env na raiz do serviço. FATURAMENTO_DB_URL é obrigatória: não há senha
// padrão embutida no código.
func Load() Config {
	loadDotEnv(".env")

	dbURL := os.Getenv("FATURAMENTO_DB_URL")
	if dbURL == "" {
		log.Fatal("variável de ambiente FATURAMENTO_DB_URL não definida (configure um arquivo .env a partir do .env.example)")
	}

	return Config{
		DatabaseURL:    dbURL,
		Port:           getEnv("FATURAMENTO_PORT", "8082"),
		EstoqueBaseURL: getEnv("ESTOQUE_BASE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

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
