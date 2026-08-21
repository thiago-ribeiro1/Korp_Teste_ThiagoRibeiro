package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"korp/estoque/internal/config"
	"korp/estoque/internal/db"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// Conecta ao banco de dados antes de subir o servidor HTTP. Se a
	// conexão falhar, o serviço não deve subir "quebrado" — preferimos
	// falhar rápido (log.Fatalf) e deixar claro o motivo no terminal.
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("falha ao conectar ao banco de dados: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("falha ao aplicar schema: %v", err)
	}

	mux := http.NewServeMux()

	// Endpoint de verificação de saúde do serviço. Usado nesta etapa para
	// confirmar que o servidor subiu e está conectado ao banco. Nas
	// próximas etapas, o frontend também vai usar um endpoint parecido
	// para detectar quando o serviço de Estoque está indisponível.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "erro",
				"detalhe": err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"servico": "estoque",
		})
	})

	log.Printf("serviço de estoque rodando na porta %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}