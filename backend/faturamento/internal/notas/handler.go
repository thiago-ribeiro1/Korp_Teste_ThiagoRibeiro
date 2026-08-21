package notas

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"korp/faturamento/internal/estoqueclient"
)

type Handler struct {
	repo    *Repository
	estoque *estoqueclient.Client
}

func NewHandler(repo *Repository, estoque *estoqueclient.Client) *Handler {
	return &Handler{repo: repo, estoque: estoque}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /notas", h.criar)
	mux.HandleFunc("GET /notas", h.listar)
	mux.HandleFunc("GET /notas/{id}", h.obter)
	mux.HandleFunc("PUT /notas/{id}", h.atualizarItens)
	mux.HandleFunc("POST /notas/{id}/imprimir", h.imprimir)
}

type notaRequest struct {
	Itens []ItemEntrada `json:"itens"`
}

func (req notaRequest) validar() error {
	if len(req.Itens) == 0 {
		return errors.New("a nota precisa ter ao menos um produto")
	}
	for _, item := range req.Itens {
		if item.Codigo == "" {
			return errors.New("todo item precisa informar o código do produto")
		}
		if item.Quantidade <= 0 {
			return errors.New("a quantidade de cada item deve ser maior que zero")
		}
	}
	return nil
}

// resolverItens consulta o Estoque para confirmar a existência de cada
// produto e obter sua descrição atual antes de gravar a nota.
func (h *Handler) resolverItens(r *http.Request, entradas []ItemEntrada) ([]ItemNota, error) {
	itens := make([]ItemNota, 0, len(entradas))
	for _, entrada := range entradas {
		produto, err := h.estoque.ObterProdutoPorCodigo(r.Context(), entrada.Codigo)
		if err != nil {
			return nil, err
		}
		itens = append(itens, ItemNota{
			ProdutoCodigo:    produto.Codigo,
			ProdutoDescricao: produto.Descricao,
			Quantidade:       entrada.Quantidade,
		})
	}
	return itens, nil
}

func (h *Handler) criar(w http.ResponseWriter, r *http.Request) {
	var req notaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := req.validar(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	itens, err := h.resolverItens(r, req.Itens)
	if err != nil {
		respondEstoqueErro(w, err)
		return
	}

	nota, err := h.repo.Create(r.Context(), itens)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "erro ao criar nota fiscal")
		return
	}

	respondJSON(w, http.StatusCreated, nota)
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	busca := r.URL.Query().Get("busca")

	notas, err := h.repo.List(r.Context(), status, busca)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "erro ao listar notas fiscais")
		return
	}
	if notas == nil {
		notas = []Nota{}
	}

	respondJSON(w, http.StatusOK, notas)
}

func (h *Handler) obter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "numeração inválida")
		return
	}

	nota, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "erro ao buscar nota fiscal")
		return
	}

	respondJSON(w, http.StatusOK, nota)
}

func (h *Handler) atualizarItens(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "numeração inválida")
		return
	}

	var req notaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := req.validar(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	itens, err := h.resolverItens(r, req.Itens)
	if err != nil {
		respondEstoqueErro(w, err)
		return
	}

	nota, err := h.repo.UpdateItens(r.Context(), id, itens)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrNotaFechada):
			respondError(w, http.StatusConflict, "notas fechadas não podem ter seus itens alterados")
		default:
			respondError(w, http.StatusInternalServerError, "erro ao atualizar nota fiscal")
		}
		return
	}

	respondJSON(w, http.StatusOK, nota)
}

// imprimir só fecha a nota depois que o Estoque confirma a baixa de saldo.
// Se o Estoque falhar, nada é persistido aqui: a nota permanece Aberta.
func (h *Handler) imprimir(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "numeração inválida")
		return
	}

	nota, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "erro ao buscar nota fiscal")
		return
	}

	if nota.Status != "Aberta" {
		respondError(w, http.StatusConflict, "somente notas com status Aberta podem ser impressas")
		return
	}

	itensBaixa := make([]estoqueclient.ItemBaixa, 0, len(nota.Itens))
	for _, item := range nota.Itens {
		itensBaixa = append(itensBaixa, estoqueclient.ItemBaixa{
			Codigo:     item.ProdutoCodigo,
			Quantidade: item.Quantidade,
		})
	}

	if err := h.estoque.BaixarLote(r.Context(), itensBaixa); err != nil {
		respondEstoqueErro(w, err)
		return
	}

	notaFechada, err := h.repo.FecharSeAberta(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotaFechada) {
			respondError(w, http.StatusConflict, "a nota já foi fechada por outra operação")
			return
		}
		respondError(w, http.StatusInternalServerError, "saldo baixado no estoque, mas houve erro ao fechar a nota")
		return
	}

	respondJSON(w, http.StatusOK, notaFechada)
}

// respondEstoqueErro traduz falhas do cliente do Estoque para respostas
// HTTP apropriadas, distinguindo indisponibilidade de erro de negócio.
func respondEstoqueErro(w http.ResponseWriter, err error) {
	var indisponivel *estoqueclient.ErrIndisponivel
	if errors.As(err, &indisponivel) {
		respondError(w, http.StatusServiceUnavailable,
			"não foi possível concluir a operação: o serviço de estoque não respondeu. "+
				"A nota permanece com status Aberta e nenhum saldo foi alterado. Tente novamente em instantes.")
		return
	}

	var negocio *estoqueclient.ErrNegocio
	if errors.As(err, &negocio) {
		status := negocio.Status
		if status < 400 || status > 599 {
			status = http.StatusBadRequest
		}
		respondError(w, status, negocio.Mensagem)
		return
	}

	respondError(w, http.StatusInternalServerError, "erro inesperado ao comunicar com o estoque")
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, mensagem string) {
	respondJSON(w, status, map[string]string{"erro": mensagem})
}
