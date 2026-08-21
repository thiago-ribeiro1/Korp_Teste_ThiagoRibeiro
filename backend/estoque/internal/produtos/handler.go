package produtos

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /produtos", h.criar)
	mux.HandleFunc("GET /produtos", h.listar)
	mux.HandleFunc("GET /produtos/{id}", h.obter)
	mux.HandleFunc("PUT /produtos/{id}", h.atualizar)
	mux.HandleFunc("DELETE /produtos/{id}", h.excluir)
	mux.HandleFunc("POST /produtos/baixa", h.baixarLote)
}

type produtoRequest struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	Saldo     *int   `json:"saldo"`
}

func (req produtoRequest) validar() error {
	if req.Codigo == "" {
		return errors.New("campo 'codigo' é obrigatório")
	}
	if req.Descricao == "" {
		return errors.New("campo 'descricao' é obrigatório")
	}
	if req.Saldo == nil {
		return errors.New("campo 'saldo' é obrigatório")
	}
	if *req.Saldo < 0 {
		return errors.New("campo 'saldo' não pode ser negativo")
	}
	return nil
}

func (h *Handler) criar(w http.ResponseWriter, r *http.Request) {
	var req produtoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := req.validar(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	produto, err := h.repo.Create(r.Context(), req.Codigo, req.Descricao, *req.Saldo)
	if err != nil {
		if errors.Is(err, ErrCodigoDuplicado) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "erro ao criar produto")
		return
	}

	respondJSON(w, http.StatusCreated, produto)
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	busca := r.URL.Query().Get("busca")

	produtos, err := h.repo.List(r.Context(), busca)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "erro ao listar produtos")
		return
	}
	if produtos == nil {
		produtos = []Produto{}
	}

	respondJSON(w, http.StatusOK, produtos)
}

func (h *Handler) obter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "id inválido")
		return
	}

	produto, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "erro ao buscar produto")
		return
	}

	respondJSON(w, http.StatusOK, produto)
}

func (h *Handler) atualizar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req produtoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := req.validar(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	produto, err := h.repo.Update(r.Context(), id, req.Codigo, req.Descricao, *req.Saldo)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrCodigoDuplicado):
			respondError(w, http.StatusConflict, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "erro ao atualizar produto")
		}
		return
	}

	respondJSON(w, http.StatusOK, produto)
}

func (h *Handler) excluir(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "id inválido")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "erro ao excluir produto")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type baixaLoteRequest struct {
	Itens []ItemBaixa `json:"itens"`
}

// baixarLote é o endpoint interno chamado pelo serviço de Faturamento no
// momento da impressão de uma nota, para dar baixa no saldo de todos os
// produtos da nota de uma só vez, de forma atômica.
func (h *Handler) baixarLote(w http.ResponseWriter, r *http.Request) {
	var req baixaLoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if len(req.Itens) == 0 {
		respondError(w, http.StatusBadRequest, "lista de itens não pode ser vazia")
		return
	}

	err := h.repo.BaixarSaldoLote(r.Context(), req.Itens)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrSaldoInsuficiente):
			respondError(w, http.StatusConflict, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "erro ao baixar saldo dos produtos")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, mensagem string) {
	respondJSON(w, status, map[string]string{"erro": mensagem})
}
