# Korp_Teste_ThiagoRibeiro

Sistema de emissão de Notas Fiscais — teste técnico Korp ERP.

## Objetivo

Aplicação para cadastro de produtos, criação de notas fiscais e impressão
(fechamento) de notas, com baixa de saldo em estoque no momento da
impressão. O sistema é dividido em dois microsserviços em Go (Estoque e
Faturamento) e um frontend em Angular.

## Arquitetura

```text
frontend/                  Angular (SPA)
backend/
  estoque/                 Microsserviço de Estoque (Go)
  faturamento/              Microsserviço de Faturamento (Go)
```

- **Estoque** (porta padrão `8081`): cadastro de produtos, consulta e baixa
  de saldo.
- **Faturamento** (porta padrão `8082`): criação e consulta de notas fiscais,
  impressão/fechamento. Chama o Estoque via HTTP no momento da impressão
  para dar baixa no saldo dos produtos utilizados.
- **Frontend** (porta padrão `4200`): consome os dois microsserviços
  diretamente via HTTP a partir do navegador.

Cada microsserviço tem seu próprio banco de dados PostgreSQL
(`estoque_db` e `faturamento_db`), reforçando que cada serviço é dono dos
seus próprios dados — não há acesso cruzado a tabelas entre eles.

### Fluxo de impressão e tratamento de falha

1. O frontend chama `POST /notas/{numeracao}/imprimir` no Faturamento.
2. O Faturamento verifica se a nota está `Aberta`. Se não estiver, retorna
   erro (`409`) e nada é alterado.
3. O Faturamento chama o Estoque (`POST /produtos/baixa`) para reduzir o
   saldo de todos os itens da nota **antes** de fechar a nota.
4. Só depois que o Estoque confirma a baixa (sucesso), o Faturamento marca
   a nota como `Fechada`.
5. Se a chamada ao Estoque falhar (timeout, conexão recusada, ou erro de
   negócio como saldo insuficiente), o Faturamento **não altera a nota**:
   ela continua `Aberta`, e o erro é retornado ao frontend com uma mensagem
   clara. O usuário pode tentar novamente assim que o Estoque voltar.

Essa ordem (baixar saldo primeiro, fechar depois) é o que garante que a nota
nunca fique `Fechada` com uma baixa de saldo que não aconteceu, sem
precisar de mecanismos de compensação/saga.

## Banco de dados

**PostgreSQL**, com acesso via `pgx` (driver nativo, sem ORM — SQL direto).
Optamos por não usar ferramenta de migration (como `golang-migrate`): cada
serviço aplica seu schema (`CREATE TABLE IF NOT EXISTS ...`) automaticamente
no startup, o que é suficiente para o escopo do teste.

A numeração sequencial da nota fiscal é a própria chave primária (`id` da
tabela `notas`), evitando manter um contador redundante em paralelo.

## Pré-requisitos

- Go 1.22+
- Node.js 20+ e npm
- PostgreSQL 14+ (rodando localmente ou acessível via rede)

## Configuração das variáveis de ambiente

Cada microsserviço lê configuração de variáveis de ambiente, carregadas
automaticamente de um arquivo `.env` na raiz do serviço (não há nenhuma
credencial padrão no código-fonte).

**1. Copie os arquivos de exemplo:**

```bash
cp backend/estoque/.env.example backend/estoque/.env
cp backend/faturamento/.env.example backend/faturamento/.env
```

**2. Edite `backend/estoque/.env` e `backend/faturamento/.env`** com a
senha real do seu PostgreSQL local:

```env
# backend/estoque/.env
ESTOQUE_DB_URL=postgres://postgres:SUA_SENHA@localhost:5432/estoque_db?sslmode=disable
ESTOQUE_PORT=8081
```

```env
# backend/faturamento/.env
FATURAMENTO_DB_URL=postgres://postgres:SUA_SENHA@localhost:5432/faturamento_db?sslmode=disable
FATURAMENTO_PORT=8082
ESTOQUE_BASE_URL=http://localhost:8081
```

**3. Crie os dois bancos de dados:**

```bash
psql -U postgres -c "CREATE DATABASE estoque_db;"
psql -U postgres -c "CREATE DATABASE faturamento_db;"
```

As tabelas são criadas automaticamente quando cada serviço sobe pela
primeira vez.

## Como executar

**Serviço de Estoque** (terminal 1):

```bash
cd backend/estoque
go mod tidy
go run ./cmd/server
```

**Serviço de Faturamento** (terminal 2):

```bash
cd backend/faturamento
go mod tidy
go run ./cmd/server
```

**Frontend** (terminal 3):

```bash
cd frontend
npm install
npm start
```

Acesse `http://localhost:4200`.

## Portas utilizadas

| Serviço      | Porta padrão |
|--------------|--------------|
| Estoque      | 8081         |
| Faturamento  | 8082         |
| Frontend     | 4200         |

## Fluxo básico da aplicação

1. Cadastre produtos em **Produtos** (código, descrição, saldo).
2. Crie uma nota em **Notas Fiscais → Nova Nota Fiscal**, adicionando
   produtos e quantidades. A nota nasce com status `Aberta` e o saldo dos
   produtos **não** é alterado nesse momento.
3. Abra a nota e clique em **Imprimir Nota**. O saldo dos produtos é
   reduzido e a nota passa para `Fechada`.
4. Uma nota `Fechada` não pode ser impressa novamente nem ter seus itens
   alterados.

## Como demonstrar o cenário de falha do Estoque

1. Com o Faturamento e o frontend rodando, **pare o serviço de Estoque**
   (`Ctrl+C` no terminal dele).
2. Abra uma nota `Aberta` e clique em **Imprimir Nota**.
3. A aplicação exibe uma mensagem de erro informando que o serviço de
   estoque não respondeu, a nota continua `Aberta` e nenhum saldo foi
   alterado. O painel de "Serviços" na barra lateral também mostra o
   Estoque como indisponível.
4. Suba o serviço de Estoque novamente e clique em **Tentar novamente** —
   a impressão é concluída normalmente.

## Principais decisões técnicas

- **Sem ORM no Go**: SQL direto via `pgx`, mantendo o código explícito e
  fácil de explicar em entrevista.
- **Sem framework HTTP em Go**: `net/http` da biblioteca padrão (Go 1.22+
  já suporta roteamento por método + path), evitando dependência
  desnecessária para um projeto deste tamanho.
- **Numeração da nota = chave primária**: evita um segundo contador
  sequencial que precisaria ser mantido em sincronia.
- **`UPDATE ... WHERE status = 'Aberta'` no fechamento da nota**: fecha a
  nota de forma atômica em uma única instrução SQL, prevenindo reimpressão
  mesmo sob concorrência, sem precisar de lock explícito ou implementar o
  requisito opcional de concorrência.
- **`.env` sem dependência externa**: um parser simples de `KEY=VALUE` foi
  escrito nos dois serviços, evitando adicionar uma biblioteca só para
  isso.
- **CORS liberado de forma ampla**: como o teste não pede autenticação nem
  múltiplos ambientes, o `Access-Control-Allow-Origin` é `*` nos dois
  serviços.

---

## Detalhamento técnico (para o vídeo de apresentação)

### Angular

- **Standalone Components**: toda a aplicação usa componentes standalone
  (sem `NgModule`), com lazy loading das telas via `loadComponent()` nas
  rotas (`app.routes.ts`).
- **Ciclos de vida utilizados**: `ngOnInit`, usado nos componentes de
  listagem e detalhe para disparar a carga inicial de dados (produtos,
  notas, verificação de status dos serviços) assim que o componente é
  montado. Não foram usados outros lifecycle hooks (`ngOnChanges`,
  `ngOnDestroy` etc.) porque não havia necessidade real no escopo do
  projeto — os componentes não recebem `@Input()` dinâmicos nem precisam
  de limpeza manual de recursos.
- **RxJS**: usado através dos `Observable` retornados pelo `HttpClient`
  (toda chamada HTTP). Em `HealthService`, `forkJoin` combina as duas
  checagens de saúde (Estoque e Faturamento) em uma única resposta, e
  `catchError` traduz falha de rede em `false` (serviço indisponível) em
  vez de propagar erro. Não foram usados operadores mais avançados
  (`switchMap`, `debounceTime` etc.) porque a aplicação não tem cenários
  como busca reativa com digitação contínua — os filtros são aplicados sob
  clique/enter explícito do usuário.
- **Outras bibliotecas**: nenhuma além do próprio Angular
  (`@angular/common`, `@angular/forms` para `ngModel`, `@angular/router`).
- **Componentes visuais**: CSS próprio (SCSS/CSS puro com variáveis
  globais em `styles.scss`), sem biblioteca de UI (Angular Material,
  PrimeNG etc.). Optado por não usar biblioteca de componentes porque a
  interface do protótipo é composta de elementos simples (tabelas, modais,
  badges, formulários) que não justificam a dependência adicional.
- **Comunicação HTTP com os microsserviços**: `HttpClient` (via
  `provideHttpClient()` em `app.config.ts`), com dois serviços Angular
  (`ProdutosService`, `NotasService`) apontando cada um para a URL base do
  seu microsserviço correspondente, configuradas em `environment.ts`.

### Go

- **Gerenciamento de dependências**: Go Modules (`go.mod`/`go.sum`) em
  cada microsserviço, de forma independente — cada serviço é seu próprio
  módulo Go.
- **Frameworks utilizados**: nenhum framework HTTP externo. Os dois
  serviços usam `net/http` da biblioteca padrão, aproveitando o roteamento
  por método + path introduzido no Go 1.22 (`mux.HandleFunc("GET /produtos", ...)`).
  A única dependência externa é o driver de banco `github.com/jackc/pgx/v5`.
- **Tratamento de erros e exceções**: Go não tem exceções — erros são
  valores retornados explicitamente. Cada camada (repository → handler)
  propaga erros com `fmt.Errorf("...: %w", err)` para preservar contexto,
  e o handler HTTP traduz erros de domínio (`ErrNotFound`,
  `ErrNotaFechada`, `ErrSaldoInsuficiente` etc.) para o status HTTP
  apropriado (`404`, `409`, `400`) usando `errors.Is`/`errors.As`. Erros
  inesperados (falha de banco, etc.) retornam `500` com mensagem genérica,
  sem vazar detalhes internos para o cliente.
- **C# / LINQ**: não se aplica — a implementação usa Go, não C#.

### Comunicação entre os microsserviços

O Faturamento se comunica com o Estoque via HTTP/REST simples
(`internal/estoqueclient`), usando `http.Client` com timeout de 5 segundos.
Dois tipos de erro são diferenciados no cliente:

- `ErrIndisponivel`: falha de transporte (timeout, conexão recusada) —
  vira `503` para o frontend, com a mensagem de que o estoque não
  respondeu.
- `ErrNegocio`: erro retornado pelo próprio Estoque com um status HTTP
  (ex: `404` produto não encontrado, `409` saldo insuficiente) — propagado
  com o mesmo status e mensagem.

Essa distinção permite que a nota permaneça `Aberta` tanto em caso de
indisponibilidade quanto de erro de negócio, sempre sem alterar saldo
parcialmente (a baixa em lote no Estoque roda em uma única transação
SQL — se qualquer item falhar, nada é persistido).
