# microservices/order

Microsserviço **Order** — responsável por receber e persistir pedidos de compra em um sistema de comércio eletrônico. Implementado em **Go** com **arquitetura hexagonal** e comunicação via **gRPC**.

---

## Visão Geral

O serviço Order expõe um método gRPC `Create` que recebe os dados de um pedido (cliente + lista de itens), persiste no banco de dados MySQL e retorna o ID gerado. A arquitetura hexagonal garante que a lógica de negócio fique completamente isolada dos detalhes de infraestrutura (banco de dados, protocolo de comunicação).

---

## Arquitetura Hexagonal

```
                        ┌─────────────────────────────┐
          Driver Side   │                             │   Driven Side
                        │                             │
  ┌──────────┐          │  ┌──────────────────────┐   │          ┌──────────┐
  │  Cliente │──gRPC───▶│  │   Application Core   │   │──Save───▶│  MySQL   │
  │  gRPC    │          │  │                      │   │          │  (GORM)  │
  └──────────┘          │  │  domain/order.go     │   │          └──────────┘
                        │  │  api/api.go          │   │
  (futuro)              │  │                      │   │
  ┌──────────┐          │  └──────────────────────┘   │
  │  Cliente │──REST───▶│                             │
  │  HTTP    │          │       Ports (interfaces)    │
  └──────────┘          │                             │
                        └─────────────────────────────┘
```

A arquitetura hexagonal separa a aplicação em três zonas:

- **Application Core** — lógica de negócio pura, sem dependências externas
- **Ports** — interfaces Go que definem os contratos de entrada e saída
- **Adapters** — implementações concretas dos ports para tecnologias específicas

Trocar o banco de dados (ex: de MySQL para PostgreSQL) requer apenas escrever um novo adapter em `internal/adapters/db/`, sem tocar na lógica de negócio.

---

## Estrutura de Pastas

```
microservices/order/
├── cmd/
│   └── main.go                          # Ponto de entrada: conecta todos os componentes
├── config/
│   └── config.go                        # Leitura de variáveis de ambiente
├── internal/
│   ├── adapters/
│   │   ├── db/
│   │   │   └── db.go                    # Adapter MySQL via GORM (implementa DBPort)
│   │   ├── grpc/
│   │   │   └── server.go                # Adapter gRPC (implementa APIPort, expõe Create)
│   │   └── rest/
│   │       └── server.go                # Placeholder para adapter REST (futuro)
│   ├── application/
│   │   └── core/
│   │       ├── api/
│   │       │   └── api.go               # Lógica de negócio: PlaceOrder
│   │       └── domain/
│   │           └── order.go             # Entidades: Order, OrderItem, NewOrder
│   └── ports/
│       ├── api.go                        # Interface APIPort (PlaceOrder)
│       └── db.go                         # Interface DBPort (Get, Save)
└── go.mod                               # Módulo Go com dependências
```

---

## Detalhamento dos Arquivos

### `internal/application/core/domain/order.go`
Define as **entidades de domínio** do sistema:

- `OrderItem` — item individual: código do produto, preço unitário, quantidade
- `Order` — pedido completo: ID, ID do cliente, status, lista de itens, data de criação
- `NewOrder(customerId, orderItems)` — função construtora que inicializa um pedido com status `"Pending"` e timestamp atual

Estas structs são a **linguagem ubíqua** da aplicação. Todo o sistema fala em termos dessas entidades, independentemente do banco ou do protocolo usado.

### `internal/ports/api.go`
Define a interface **APIPort**:

```go
type APIPort interface {
    PlaceOrder(order domain.Order) (domain.Order, error)
}
```

Os **adapters de entrada** (gRPC, REST) usam esta interface para chamar a lógica de negócio sem conhecer sua implementação interna.

### `internal/ports/db.go`
Define a interface **DBPort**:

```go
type DBPort interface {
    Get(id string) (domain.Order, error)
    Save(*domain.Order) error
}
```

O **núcleo da aplicação** usa esta interface para persistir dados sem depender diretamente de MySQL ou qualquer outro banco.

### `internal/application/core/api/api.go`
Implementação da **lógica de negócio**:

- Recebe um `domain.Order` via `PlaceOrder`
- Chama `db.Save()` para persistir (usando a interface `DBPort`)
- Retorna o pedido com o ID preenchido pelo banco

A struct `Application` só conhece `ports.DBPort` — nunca o GORM diretamente.

### `internal/adapters/grpc/server.go`
**Adapter de entrada** gRPC:

- Incorpora `order.UnimplementedOrderServer` (gerado pelo protobuf) para satisfazer a interface gRPC
- Implementa o método `Create(ctx, *order.CreateOrderRequest)` definido no `.proto`
- Converte `order.CreateOrderRequest` → `domain.Order` → chama `api.PlaceOrder()`
- Converte o resultado de volta para `*order.CreateOrderResponse`
- A função `Run()` inicializa o servidor TCP gRPC na porta configurada
- Em modo `development`, habilita **server reflection** (permite uso do `grpcurl`)

### `internal/adapters/db/db.go`
**Adapter de saída** MySQL com GORM:

- Define structs locais `Order` e `OrderItem` (com `gorm.Model` para ID/timestamps automáticos)
- `NewAdapter(dsn)` abre conexão e executa `AutoMigrate` (cria tabelas automaticamente)
- `Save(*domain.Order)` — converte `domain.Order` para structs GORM e persiste; preenche `order.ID` após criação
- `Get(id)` — busca pedido no banco e converte de volta para `domain.Order`

### `config/config.go`
Lê variáveis de ambiente obrigatórias:

| Variável | Uso |
|---|---|
| `ENV` | Ambiente (`development`, `production`) |
| `DATA_SOURCE_URL` | DSN do MySQL |
| `APPLICATION_PORT` | Porta TCP do servidor gRPC |

### `cmd/main.go`
**Ponto de entrada** — conecta todos os componentes na ordem correta:

```
dbAdapter    = db.NewAdapter(dsn)         ← implementa DBPort
application  = api.NewApplication(db)    ← recebe DBPort, implementa APIPort
grpcAdapter  = grpc.NewAdapter(app, port) ← recebe APIPort
grpcAdapter.Run()                         ← inicia servidor
```

---

## Fluxo de uma Requisição

```
Cliente gRPC
    │
    │  CreateOrderRequest{costumer_id, order_items}
    ▼
grpc/server.go (Adapter)
    │  Converte para domain.OrderItem[]
    │  Cria domain.Order via domain.NewOrder()
    │  Chama api.PlaceOrder(order)
    ▼
application/core/api/api.go (Lógica de Negócio)
    │  Chama db.Save(&order)
    ▼
adapters/db/db.go (Adapter)
    │  Converte para structs GORM
    │  Executa INSERT no MySQL
    │  Preenche order.ID
    ▼
application/core/api/api.go
    │  Retorna order com ID preenchido
    ▼
grpc/server.go (Adapter)
    │  Converte para CreateOrderResponse{order_id}
    ▼
Cliente gRPC
    Recebe: CreateOrderResponse{order_id: 42}
```

---

## Como Executar

### 1. Subir o MySQL com Docker

```bash
docker run -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=minhasenha \
  -e MYSQL_DATABASE=order \
  mysql
```

### 2. Executar o serviço

```bash
DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/order" \
APPLICATION_PORT=3000 \
ENV=development \
go run cmd/main.go
```

### 3. Testar com grpcurl

```bash
grpcurl \
  -d '{"user_id": 123, "order_items": [{"product_code": "prod", "quantity": 4, "unit_price": 12}]}' \
  -plaintext \
  localhost:3000 \
  Order/Create
```

---

## Dependências

| Pacote | Versão | Uso |
|---|---|---|
| `google.golang.org/grpc` | v1.63.2 | Framework gRPC |
| `gorm.io/gorm` | v1.25.9 | ORM para acesso ao banco |
| `gorm.io/driver/mysql` | v1.5.6 | Driver MySQL para o GORM |
| `github.com/ruandg/microservices-proto/golang/order` | local | Stubs gerados pelo protobuf |

---

## Tecnologias

- [Go](https://golang.org/) 1.22+
- [gRPC](https://grpc.io/)
- [Protocol Buffers v3](https://protobuf.dev/)
- [GORM](https://gorm.io/)
- [MySQL](https://www.mysql.com/)
- [Docker](https://www.docker.com/) (para o banco de dados em desenvolvimento)
