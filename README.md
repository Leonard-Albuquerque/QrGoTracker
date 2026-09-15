# QrGoTracker

> This documentation is available in two languages:
> - [Portugues (Brasil)](#portugues-brasil)
> - [English](#english)

---

<a name="portugues-brasil"></a>

# Portugues (Brasil)

Encurtador de URLs leve com gerador de QR Code, construido em Go. A partir de uma URL longa, o servico gera um link curto de redirecionamento e uma imagem QR correspondente, registrando quantas vezes cada link foi acessado.

---

## Sumario

- [Visao Geral](#visao-geral)
- [Arquitetura](#arquitetura)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Referencia da API](#referencia-da-api)
- [Configuracao](#configuracao)
- [Executando Localmente](#executando-localmente)
- [Docker](#docker-pt)
- [Tecnologias Utilizadas](#tecnologias-utilizadas)

---

## Visao Geral

O QrGoTracker expoe uma API JSON e uma interface web renderizada no servidor. O fluxo principal e:

1. O cliente envia um POST com a URL de destino para a API.
2. O servico gera (ou aceita) um codigo alfanumerico curto e persiste o link no SQLite.
3. A URL curta (`/r/{code}`) emite um redirecionamento HTTP 302 e incrementa o contador de cliques.
4. O endpoint de QR Code (`/qr/{code}.png`) retorna uma imagem PNG codificando a URL curta.
5. As estatisticas podem ser consultadas via API ou visualizadas no navegador em `/stats/{code}`.

---

## Arquitetura

A aplicacao segue uma arquitetura em camadas com separacao clara de responsabilidades:

```
Requisicao HTTP
    |
    v
[ Middleware ]        Request ID, Logging, Security Headers, Rate Limiter
    |
    v
[ Handler ]          Decodificacao/codificacao HTTP, validacao de entrada, mapeamento de erros
    |
    v
[ Service ]          Logica de negocio: geracao de codigo, validacao de URL, rastreamento de cliques
    |
    v
[ Repository ]       Consultas SQL contra o SQLite
    |
    v
[ Database ]         Arquivo SQLite via modernc/sqlite (Go puro, sem CGO)
```

A camada de interface web (`internal/web`) e independente dos handlers da API e serve templates HTML renderizados no servidor, embutidos no binario.

---

## Estrutura do Projeto

```
.
├── cmd/
│   └── server/
│       └── main.go          # Ponto de entrada: conecta dependencias, configura o router, inicia o servidor
├── internal/
│   ├── config/              # Carrega configuracoes a partir de variaveis de ambiente
│   ├── database/            # Conexao SQLite e migracao automatica
│   ├── handler/             # Handlers HTTP para endpoints da API e redirecionamento
│   ├── middleware/          # Request ID, logging estruturado, cabecalhos de seguranca, rate limiting
│   ├── models/              # Modelo de dominio (Link)
│   ├── repository/          # Repositorio SQLite implementando a persistencia
│   ├── service/             # Logica de negocio (criacao de links, rastreamento de cliques, construcao de URLs)
│   ├── utils/               # Validacao de URL, geracao e validacao de codigos curtos
│   └── web/                 # Templates HTML e assets estaticos
├── migrations/
│   └── 001_init.sql         # Schema inicial: tabela de links
├── Dockerfile
├── .env.example
└── requests.http            # Requisicoes HTTP de exemplo para testes manuais
```

---

## Referencia da API

Todas as respostas da API utilizam `Content-Type: application/json`.

### Health Check

```
GET /healthz
```

Resposta `200 OK`:
```json
{ "status": "ok" }
```

---

### Criar um Link Curto

```
POST /api/links
Content-Type: application/json
```

Corpo da requisicao:

| Campo        | Tipo   | Obrigatorio | Descricao                                                        |
|--------------|--------|-------------|------------------------------------------------------------------|
| `target_url` | string | Sim         | URL de destino (deve ser http ou https)                          |
| `name`       | string | Nao         | Codigo curto personalizado (3-40 caracteres, letras/digitos/`_`/`-`). Gerado automaticamente se omitido. |

Resposta `201 Created`:
```json
{
  "code":        "n13c6DG",
  "short_url":   "http://localhost:8080/r/n13c6DG",
  "qr_url":      "http://localhost:8080/qr/n13c6DG.png",
  "target_url":  "https://example.com/page",
  "click_count": 0,
  "created_at":  "2026-09-14T23:00:00Z",
  "is_active":   true
}
```

Codigos de erro: `invalid_payload`, `invalid_url`, `invalid_name`, `name_unavailable`.

---

### Obter Detalhes do Link

```
GET /api/links/{code}
```

Resposta `200 OK`: mesmo formato da resposta de criacao.

---

### Obter Estatisticas do Link

```
GET /api/links/{code}/stats
```

Resposta `200 OK`:
```json
{
  "code":        "n13c6DG",
  "target_url":  "https://example.com/page",
  "click_count": 42,
  "created_at":  "2026-09-14T23:00:00Z",
  "is_active":   true
}
```

---

### Redirecionamento

```
GET /r/{code}
```

Emite um redirecionamento HTTP `302 Found` para a URL de destino e incrementa o contador de cliques. Retorna `410 Gone` se o link estiver inativo.

---

### Imagem do QR Code

```
GET /qr/{code}.png
```

Retorna uma imagem PNG `256x256` codificando a URL curta com nivel de correcao de erros medio.

---

### Interface Web e WebSocket

| Rota                | Descricao                                              |
|---------------------|--------------------------------------------------------|
| `GET /`             | Pagina inicial (formulario de criacao)                 |
| `GET /stats/{code}` | Pagina de estatisticas no navegador                    |
| `GET /ws/{code}`    | Conexao WebSocket para atualizacoes de clique ao vivo  |

---

## Configuracao

Copie `.env.example` para `.env` e ajuste os valores antes de executar.

| Variavel       | Padrao                             | Descricao                                                        |
|----------------|------------------------------------|------------------------------------------------------------------|
| `PORT`         | `8080`                             | Porta TCP em que o servidor escuta                               |
| `BASE_URL`     | `http://localhost:8080`            | URL base publica usada para construir URLs curtas e de QR Code   |
| `DB_PATH`      | `./data/qr.db`                     | Caminho para o arquivo do banco de dados SQLite                  |
| `HASH_SALT`    | `change-me-but-set-a-secure-value` | Salt para hashing opcional de IP. Defina um valor forte em producao. |
| `RL_ENABLED`   | `true`                             | Habilitar ou desabilitar o rate limiter                          |
| `RL_RPS`       | `5`                                | Requisicoes sustentadas por segundo por IP                       |
| `RL_BURST`     | `10`                               | Tamanho maximo de rajada para o rate limiter                     |
| `ALLOW_ORIGINS`| _(vazio)_                          | Origens CORS separadas por virgula. Vazio desabilita o CORS.     |
| `LOG_LEVEL`    | `info`                             | Verbosidade do log (apenas informacional, JSON estruturado)      |

---

## Executando Localmente

**Pre-requisitos:** Go 1.24+

```bash
# Clonar o repositorio
git clone https://github.com/your-username/QrGoTracker.git
cd QrGoTracker

# Copiar e editar as variaveis de ambiente
cp .env.example .env

# Iniciar o servidor
go run ./cmd/server
```

O servidor inicia na porta definida em `.env` (padrao: `8080`).

O banco de dados SQLite e o schema sao criados automaticamente na primeira execucao — nenhuma etapa de migracao manual e necessaria.

---

<a name="docker-pt"></a>

## Docker

```bash
# Construir a imagem
docker build -t qr-go-tracker .

# Executar o container
docker run -p 8080:8085 \
  -e BASE_URL=http://localhost:8080 \
  -e DB_PATH=/data/qr.db \
  -v $(pwd)/data:/data \
  qr-go-tracker
```

> O Dockerfile compila o binario dentro da imagem oficial `golang:1.24-alpine` e expoe a porta `8085`. Mapeie para a porta do host de sua preferencia.

---

## Tecnologias Utilizadas

| Componente       | Escolha                                                         |
|------------------|-----------------------------------------------------------------|
| Linguagem        | Go 1.24                                                         |
| Router HTTP      | [chi v5](https://github.com/go-chi/chi)                         |
| Banco de Dados   | SQLite via [modernc/sqlite](https://pkg.go.dev/modernc.org/sqlite) (Go puro, sem CGO) |
| Geracao de QR    | [skip2/go-qrcode](https://github.com/skip2/go-qrcode)           |
| Logging          | `log/slog` (JSON estruturado, stdlib)                           |
| Real-time WS     | [gorilla/websocket](https://github.com/gorilla/websocket)      |
| IDs Unicos       | `crypto/rand` + `encoding/hex`                                  |
| Containerizacao  | Docker (Alpine)                                                 |

---

<a name="english"></a>

# English

A lightweight URL shortener and QR code generator built with Go. Given a long URL, the service produces a short redirect link and a corresponding QR code image, and tracks how many times each link has been accessed.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Running Locally](#running-locally)
- [Docker](#docker)
- [Tech Stack](#tech-stack)

---

## Overview

QrGoTracker exposes a small JSON API and a minimal server-rendered web interface. The core flow is:

1. A client POSTs a target URL to the API.
2. The service generates (or accepts) a short alphanumeric code and persists the link in SQLite.
3. The short URL (`/r/{code}`) issues an HTTP 302 redirect and increments a click counter.
4. The QR code endpoint (`/qr/{code}.png`) returns a PNG image encoding the short URL.
5. Statistics can be queried via the API or viewed in a browser at `/stats/{code}`.

---

## Architecture

The application follows a layered architecture with clear separation of concerns:

```
HTTP Request
    |
    v
[ Middleware ]        Request ID, Logging, Security Headers, Rate Limiter
    |
    v
[ Handler ]          HTTP decoding/encoding, input validation, error mapping
    |
    v
[ Service ]          Business logic: code generation, URL validation, click tracking
    |
    v
[ Repository ]       SQL queries against SQLite
    |
    v
[ Database ]         SQLite file via modernc/sqlite (pure Go, no CGO)
```

The web UI layer (`internal/web`) is independent from the API handlers and serves server-side rendered HTML templates embedded in the binary.

---

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go          # Entry point: wires dependencies, configures router, starts server
├── internal/
│   ├── config/              # Loads configuration from environment variables
│   ├── database/            # SQLite connection and auto-migration
│   ├── handler/             # HTTP handlers for API and redirect endpoints
│   ├── middleware/          # Request ID, structured logging, security headers, rate limiting
│   ├── models/              # Domain model (Link)
│   ├── repository/          # SQLite repository implementing persistence
│   ├── service/             # Business logic (link creation, click tracking, URL building)
│   ├── utils/               # URL validation, short code generation and validation
│   └── web/                 # Server-side HTML templates and static assets
├── migrations/
│   └── 001_init.sql         # Initial schema: links table
├── Dockerfile
├── .env.example
└── requests.http            # Sample HTTP requests for manual testing
```

---

## API Reference

All API responses use `Content-Type: application/json`.

### Health Check

```
GET /healthz
```

Response `200 OK`:
```json
{ "status": "ok" }
```

---

### Create a Short Link

```
POST /api/links
Content-Type: application/json
```

Request body:

| Field        | Type   | Required | Description                                      |
|--------------|--------|----------|--------------------------------------------------|
| `target_url` | string | Yes      | The destination URL (must be http or https)      |
| `name`       | string | No       | Custom short code (3-40 chars, letters/digits/`_`/`-`). Auto-generated if omitted. |

Response `201 Created`:
```json
{
  "code":        "n13c6DG",
  "short_url":   "http://localhost:8080/r/n13c6DG",
  "qr_url":      "http://localhost:8080/qr/n13c6DG.png",
  "target_url":  "https://example.com/page",
  "click_count": 0,
  "created_at":  "2026-09-14T23:00:00Z",
  "is_active":   true
}
```

Error codes: `invalid_payload`, `invalid_url`, `invalid_name`, `name_unavailable`.

---

### Get Link Details

```
GET /api/links/{code}
```

Response `200 OK`: same shape as the Create response.

---

### Get Link Statistics

```
GET /api/links/{code}/stats
```

Response `200 OK`:
```json
{
  "code":        "n13c6DG",
  "target_url":  "https://example.com/page",
  "click_count": 42,
  "created_at":  "2026-09-14T23:00:00Z",
  "is_active":   true
}
```

---

### Redirect

```
GET /r/{code}
```

Issues an HTTP `302 Found` to the target URL and increments the click counter. Returns `410 Gone` if the link is inactive.

---

### QR Code Image

```
GET /qr/{code}.png
```

Returns a `256x256` PNG image encoding the short URL at medium error-correction level.

---

### Web UI and WebSocket

| Route               | Description                             |
|---------------------|-----------------------------------------|
| `GET /`             | Index page (link creation form)         |
| `GET /stats/{code}` | Browser-friendly statistics page        |
| `GET /ws/{code}`    | WebSocket connection for live clicks    |

---

## Configuration

Copy `.env.example` to `.env` and adjust the values before running.

| Variable       | Default                            | Description                                               |
|----------------|------------------------------------|-----------------------------------------------------------|
| `PORT`         | `8080`                             | TCP port the server listens on                            |
| `BASE_URL`     | `http://localhost:8080`            | Public base URL used to build short and QR URLs           |
| `DB_PATH`      | `./data/qr.db`                     | Path to the SQLite database file                          |
| `HASH_SALT`    | `change-me-but-set-a-secure-value` | Salt for optional IP hashing. Set a strong value in production. |
| `RL_ENABLED`   | `true`                             | Enable or disable the rate limiter                        |
| `RL_RPS`       | `5`                                | Sustained requests per second per IP                      |
| `RL_BURST`     | `10`                               | Maximum burst size for the rate limiter                   |
| `ALLOW_ORIGINS`| _(empty)_                          | Comma-separated CORS origins. Leave empty to disable CORS.|
| `LOG_LEVEL`    | `info`                             | Log verbosity (informational only, structured JSON)       |

---

## Running Locally

**Prerequisites:** Go 1.24+

```bash
# Clone the repository
git clone https://github.com/your-username/QrGoTracker.git
cd QrGoTracker

# Copy and edit environment variables
cp .env.example .env

# Run the server
go run ./cmd/server
```

The server starts on the port defined in `.env` (default: `8080`).

The SQLite database and schema are created automatically on first run — no manual migration step is required.

---

## Docker

```bash
# Build the image
docker build -t qr-go-tracker .

# Run the container
docker run -p 8080:8085 \
  -e BASE_URL=http://localhost:8080 \
  -e DB_PATH=/data/qr.db \
  -v $(pwd)/data:/data \
  qr-go-tracker
```

> The Dockerfile compiles the binary inside the official `golang:1.24-alpine` image and exposes port `8085`. Map it to whichever host port you prefer.

---

## Tech Stack

| Component        | Choice                                                  |
|------------------|---------------------------------------------------------|
| Language         | Go 1.24                                                 |
| HTTP Router      | [chi v5](https://github.com/go-chi/chi)                 |
| Database         | SQLite via [modernc/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO) |
| QR Generation    | [skip2/go-qrcode](https://github.com/skip2/go-qrcode)  |
| Logging          | `log/slog` (structured JSON, stdlib)                    |
| Real-time WS     | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Unique IDs       | `crypto/rand` + `encoding/hex`                          |
| Containerization | Docker (Alpine-based)                                   |
