package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"korp/faturamento/internal/config"
	"korp/faturamento/internal/db"
	"korp/faturamento/internal/estoqueclient"
	"korp/faturamento/internal/notas"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("falha ao conectar ao banco de dados: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("falha ao aplicar schema: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "erro", "detalhe": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "servico": "faturamento"})
	})

	estoqueClient := estoqueclient.New(cfg.EstoqueBaseURL)
	notasRepo := notas.NewRepository(pool)
	notasHandler := notas.NewHandler(notasRepo, estoqueClient)
	notasHandler.RegisterRoutes(mux)

	log.Printf("serviço de faturamento rodando na porta %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, withCORS(mux)); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
