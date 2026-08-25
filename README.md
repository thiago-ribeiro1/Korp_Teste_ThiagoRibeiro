# Sistema de emissão de Notas Fiscais

## Sobre o projeto

Aplicação para cadastro de produtos, criação de notas fiscais e impressão
(fechamento) de notas, com baixa de saldo em estoque no momento da
impressão. O sistema é dividido em dois microsserviços em Go (Estoque e
Faturamento) e um frontend em Angular.

## Tecnologias utilizadas

- **Backend**: Go 1.22+, `net/http` (biblioteca padrão, sem framework), `pgx`
  (driver PostgreSQL, sem ORM)
- **Frontend**: Angular (standalone components), RxJS, HttpClient
- **Banco de dados**: PostgreSQL (um banco por serviço)

## Arquitetura

```text
frontend/                  Angular (SPA)
backend/
  estoque/                 Microsserviço de Estoque (Go)
  faturamento/             Microsserviço de Faturamento (Go)
```

- **Estoque** (porta `8081`): cadastro de produtos, consulta e baixa de saldo.
- **Faturamento** (porta `8082`): criação e consulta de notas fiscais,
  impressão/fechamento. Chama o Estoque via HTTP no momento da impressão
  para dar baixa no saldo dos produtos utilizados.
- **Frontend** (porta `4200`): consome os dois microsserviços diretamente
  via HTTP a partir do navegador.

Cada microsserviço tem seu próprio banco (`estoque_db` e `faturamento_db`) -
cada serviço é dono dos seus dados, sem acesso cruzado a tabelas entre eles.

### Fluxo de impressão de uma nota

1. O frontend chama `POST /notas/{id}/imprimir` no Faturamento.
2. O Faturamento verifica se a nota está `Aberta` (senão retorna `409`).
3. O Faturamento chama o Estoque (`POST /produtos/baixa`) para reduzir o
   saldo dos itens da nota **antes** de fechá-la.
4. Só depois que o Estoque confirma a baixa, a nota é marcada como `Fechada`.
5. Se a chamada ao Estoque falhar (timeout, saldo insuficiente etc.), a nota
   **continua `Aberta`** e o erro é exibido ao usuário, que pode tentar
   novamente.

Essa ordem (baixar saldo primeiro, fechar depois) garante que a nota nunca
fique `Fechada` com uma baixa que não aconteceu.

## Endpoints principais

**Faturamento** (`:8082`)

| Método | Rota                     | Descrição                         |
|--------|--------------------------|------------------------------------|
| POST   | `/notas`                 | Cria uma nota fiscal               |
| GET    | `/notas`                 | Lista as notas fiscais             |
| GET    | `/notas/{id}`            | Consulta uma nota                  |
| PUT    | `/notas/{id}`            | Atualiza os itens de uma nota      |
| POST   | `/notas/{id}/imprimir`   | Imprime/fecha a nota               |
| GET    | `/health`                | Healthcheck                        |

**Estoque** (`:8081`)

| Método | Rota                | Descrição                     |
|--------|---------------------|--------------------------------|
| POST   | `/produtos`         | Cadastra um produto            |
| GET    | `/produtos`         | Lista os produtos              |
| GET    | `/produtos/{id}`    | Consulta um produto            |
| PUT    | `/produtos/{id}`    | Atualiza um produto            |
| DELETE | `/produtos/{id}`    | Remove um produto              |
| POST   | `/produtos/baixa`   | Baixa em lote o saldo de itens |
| GET    | `/health`           | Healthcheck                    |

## Pré-requisitos

- Go 1.22+
- Node.js 20+ e npm
- PostgreSQL 14+ (local ou acessível via rede)

**1. Copie os arquivos de exemplo:**

```bash
cp backend/estoque/.env.example backend/estoque/.env
cp backend/faturamento/.env.example backend/faturamento/.env
```

**2. Edite os dois `.env`** com a senha do seu PostgreSQL local:

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

As tabelas são criadas automaticamente na primeira execução de cada serviço.

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

## Principais funcionalidades

1. Cadastro de produtos (**Produtos**: código, descrição, saldo).
2. Criação de notas fiscais (**Notas Fiscais → Nova Nota Fiscal**), com
   itens e quantidades. A nota nasce `Aberta` e o saldo dos produtos
   **não** é alterado nesse momento.
3. Impressão da nota (**Imprimir Nota**): reduz o saldo dos produtos e
   fecha a nota (`Fechada`).
4. Uma nota `Fechada` não pode ser reimpressa nem ter seus itens alterados.

### Testando o cenário de falha do Estoque

1. Com Faturamento e frontend rodando, pare o serviço de Estoque (`Ctrl+C`).
2. Abra uma nota `Aberta` e clique em **Imprimir Nota**.
3. A aplicação mostra o erro, a nota continua `Aberta` e nenhum saldo é
   alterado (o painel de "Serviços" também indica o Estoque indisponível).
4. Suba o Estoque novamente e clique em **Tentar novamente** - a impressão
   é concluída normalmente.
