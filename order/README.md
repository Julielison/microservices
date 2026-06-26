# microservices/order

Microsserviço **Order** — responsável por receber e persistir pedidos de compra em um sistema de comércio eletrônico. Implementado em **Go** com **arquitetura hexagonal** e comunicação via **gRPC**.

> **Parte 2:** o serviço Order agora atua também como **cliente gRPC**, integrando-se ao microsserviço **Payment** para realizar a cobrança do cliente após o registro do pedido.

---

## Visão Geral

O serviço Order expõe um método gRPC `Create` que recebe os dados de um pedido (cliente + lista de itens), persiste no banco de dados MySQL, solicita a cobrança ao microsserviço **Payment** e retorna o ID do pedido gerado. A arquitetura hexagonal garante que a lógica de negócio fique completamente isolada dos detalhes de infraestrutura (banco de dados, protocolo de comunicação, serviços externos).

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
  (futuro)              │  │                      │   │──Charge─▶┌──────────┐
  ┌──────────┐          │  └──────────────────────┘   │          │ Payment  │
  │  Cliente │──REST───▶│                             │          │ (gRPC)   │
  │  HTTP    │          │       Ports (interfaces)    │          └──────────┘
  └──────────┘          │                             │
                        └─────────────────────────────┘
```

A arquitetura hexagonal separa a aplicação em três zonas:

- **Application Core** — lógica de negócio pura, sem dependências externas
- **Ports** — interfaces Go que definem os contratos de entrada e saída
- **Adapters** — implementações concretas dos ports para tecnologias específicas

Trocar o banco de dados (ex: de MySQL para PostgreSQL) requer apenas escrever um novo adapter em `internal/adapters/db/`, sem tocar na lógica de negócio. Da mesma forma, o serviço Order não conhece detalhes de como o Payment realiza a cobrança — ele apenas chama `PaymentPort.Charge`.

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
│   │   │   └── server.go                # Adapter gRPC de entrada (implementa APIPort, expõe Create)
│   │   ├── payment/
│   │   │   └── payment.go               # Adapter gRPC de saída (implementa PaymentPort, chama o serviço Payment)
│   │   └── rest/
│   │       └── server.go                # Placeholder para adapter REST (futuro)
│   ├── application/
│   │   └── core/
│   │       ├── api/
│   │       │   └── api.go               # Lógica de negócio: PlaceOrder
│   │       └── domain/
│   │           └── order.go             # Entidades: Order, OrderItem, NewOrder, TotalPrice
│   └── ports/
│       ├── api.go                        # Interface APIPort (PlaceOrder)
│       ├── db.go                         # Interface DBPort (Get, Save)
│       └── payment.go                    # Interface PaymentPort (Charge)
└── go.mod                               # Módulo Go com dependências
```

---

## Detalhamento dos Arquivos

### `internal/application/core/domain/order.go`
Define as **entidades de domínio** do sistema:

- `OrderItem` — item individual: código do produto, preço unitário, quantidade
- `Order` — pedido completo: ID, ID do cliente, status, lista de itens, data de criação
- `NewOrder(customerId, orderItems)` — função construtora que inicializa um pedido com status `"Pending"` e timestamp atual
- `TotalPrice()` — **(Parte 2)** método que calcula o valor total do pedido, somando `UnitPrice * Quantity` de cada item. É usado para informar o valor a ser cobrado ao microsserviço Payment

Estas structs são a **linguagem ubíqua** da aplicação. Todo o sistema fala em termos dessas entidades, independentemente do banco, do protocolo ou do serviço externo usado.

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

### `internal/ports/payment.go` (Parte 2)
Define a interface **PaymentPort**:

```go
type PaymentPort interface {
    Charge(order *domain.Order) error
}
```

O **núcleo da aplicação** usa esta interface para solicitar a cobrança de um pedido, sem depender diretamente do protocolo gRPC ou da implementação do microsserviço Payment.

### `internal/application/core/api/api.go`
Implementação da **lógica de negócio**:

- Recebe um `domain.Order` via `PlaceOrder`
- Chama `db.Save()` para persistir o pedido (usando a interface `DBPort`)
- **(Parte 2)** Após salvar com sucesso, chama `payment.Charge()` para solicitar a cobrança (usando a interface `PaymentPort`)
- Retorna o pedido com o ID preenchido pelo banco, ou o erro ocorrido em qualquer uma das etapas

A struct `Application` só conhece `ports.DBPort` e `ports.PaymentPort` — nunca o GORM ou o cliente gRPC do Payment diretamente.

### `internal/adapters/grpc/server.go`
**Adapter de entrada** gRPC:

- Incorpora `order.UnimplementedOrderServer` (gerado pelo protobuf) para satisfazer a interface gRPC
- Implementa o método `Create(ctx, *order.CreateOrderRequest)` definido no `.proto`
- Converte `order.CreateOrderRequest` → `domain.Order` → chama `api.PlaceOrder()`
- Converte o resultado de volta para `*order.CreateOrderResponse`
- A função `Run()` inicializa o servidor TCP gRPC na porta configurada
- Em modo `development`, habilita **server reflection** (permite uso do `grpcurl`)

### `internal/adapters/payment/payment.go` (Parte 2)
**Adapter de saída** gRPC — implementa `PaymentPort` fazendo o papel de **cliente** do microsserviço Payment:

- `NewAdapter(paymentServiceUrl)` abre uma conexão gRPC com o serviço Payment (sem TLS, usando `insecure.NewCredentials()`) e inicializa o stub `payment.PaymentClient`
- `Charge(order *domain.Order)` monta um `CreatePaymentRequest` com `UserId`, `OrderId` e `TotalPrice` (calculado por `order.TotalPrice()`) e invoca o método `Create` do serviço Payment
- **(Parte 4)** A conexão é configurada com interceptor de retry automático; cada chamada possui deadline individual de 2 segundos

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
| `PAYMENT_SERVICE_URL` | **(Parte 2)** Endereço (`host:porta`) do microsserviço Payment |

### `cmd/main.go`
**Ponto de entrada** — conecta todos os componentes na ordem correta:

```
dbAdapter      = db.NewAdapter(dsn)                            ← implementa DBPort
paymentAdapter = payment_adapter.NewAdapter(paymentServiceUrl) ← implementa PaymentPort
application    = api.NewApplication(db, payment)               ← recebe DBPort e PaymentPort, implementa APIPort
grpcAdapter    = grpc.NewAdapter(app, port)                    ← recebe APIPort
grpcAdapter.Run()                                              ← inicia servidor
```

---

## Fluxo de uma Requisição (Parte 2)

```
Cliente gRPC
    │
    │  CreateOrderRequest{costumer_id, order_items}
    ▼
grpc/server.go (Adapter de entrada)
    │  Converte para domain.OrderItem[]
    │  Cria domain.Order via domain.NewOrder()
    │  Chama api.PlaceOrder(order)
    ▼
application/core/api/api.go (Lógica de Negócio)
    │  Chama db.Save(&order)              ──▶ adapters/db/db.go ──▶ INSERT no MySQL (Order)
    │  Chama payment.Charge(&order)       ──▶ adapters/payment/payment.go
    │                                          │  Monta CreatePaymentRequest{UserId, OrderId, TotalPrice}
    │                                          ▼
    │                                     microsserviço Payment (gRPC)
    │                                          │  Registra a cobrança
    │                                          ▼
    │                                     CreatePaymentResponse{payment_id, bill_id}
    ▼
application/core/api/api.go
    │  Retorna order com ID preenchido (ou erro, se DB ou Payment falharem)
    ▼
grpc/server.go (Adapter de entrada)
    │  Converte para CreateOrderResponse{order_id}
    ▼
Cliente gRPC
    Recebe: CreateOrderResponse{order_id: 42}
```

> A resposta enviada ao cliente não muda em relação à Parte 1 — a chamada ao serviço Payment é interna ao microsserviço Order. Se o `order_id` for retornado com sucesso, a comunicação entre os dois microsserviços ocorreu corretamente.

---

## Como Executar

### 1. Criar o `init.sql`

Para que o container MySQL já crie os bancos `order` e `payment` na inicialização, crie um arquivo `init.sql` (veja [`../init.sql`](../init.sql)):

```sql
CREATE DATABASE IF NOT EXISTS `order`;
CREATE DATABASE IF NOT EXISTS `payment`;
```

### 2. Subir o MySQL com Docker

```bash
docker run -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=minhasenha \
  -v "$(pwd)/init.sql:/docker-entrypoint-initdb.d/init.sql" \
  mysql
```

### 3. Baixar e executar o microsserviço Payment

Baixe o projeto do microsserviço Payment e coloque-o na pasta `microservices`, no mesmo nível da pasta `order`. Em seguida, execute-o:

```bash
DB_DRIVER=mysql \
DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/payment" \
APPLICATION_PORT=3001 \
ENV=development \
go run cmd/main.go
```

### 4. Executar o serviço Order

```bash
DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/order" \
APPLICATION_PORT=3000 \
PAYMENT_SERVICE_URL=localhost:3001 \
ENV=development \
go run cmd/main.go
```

### 5. Testar com grpcurl

```bash
grpcurl \
  -d '{"user_id": 123, "order_items": [{"product_code": "prod", "quantity": 4, "unit_price": 12}]}' \
  -plaintext \
  localhost:3000 \
  Order/Create
```

Se o `order_id` for retornado na resposta, a comunicação entre os microsserviços **Order** e **Payment** ocorreu com sucesso.

---

## Dependências

| Pacote | Versão | Uso |
|---|---|---|
| `google.golang.org/grpc` | v1.81.1 | Framework gRPC |
| `github.com/grpc-ecosystem/go-grpc-middleware` | v1.4.0 | **(Parte 4)** Interceptor de retry com backoff linear |
| `gorm.io/gorm` | v1.25.9 | ORM para acesso ao banco |
| `gorm.io/driver/mysql` | v1.5.6 | Driver MySQL para o GORM |
| `github.com/ruandg/microservices-proto/golang/order` | local | Stubs gerados pelo protobuf (Order) |
| `github.com/ruandg/microservices-proto/golang/payment` | local | **(Parte 2)** Stubs gerados pelo protobuf (Payment) |

---

## Tratamento de Erros (Parte 3)

### Validação de quantidade máxima de itens (Order)

O microsserviço Order passou a validar se o número total de itens do pedido (somando as quantidades de cada `OrderItem`) é maior que **50**. Caso seja, a requisição é rejeitada antes de qualquer acesso ao banco, retornando:

- **Status gRPC:** `INVALID_ARGUMENT`
- **Mensagem:** `"Order cannot have more than 50 items in total."`

### Validação de valor máximo (Payment)

O microsserviço Payment rejeita cobranças com `total_price > 1000`, retornando:

- **Status gRPC:** `INVALID_ARGUMENT`
- **Mensagem:** `"Payment over 1000 is not allowed."`

O adapter gRPC do Payment também distingue erros internos: qualquer erro que não seja `INVALID_ARGUMENT` é encapsulado em um erro com código `INTERNAL`.

### Atualização de status do pedido

Após o fluxo de `PlaceOrder`, o campo `Status` do pedido é atualizado no banco de dados:

| Situação | Status salvo |
|---|---|
| Cobrança efetuada com sucesso | `"Paid"` |
| Erro em qualquer etapa (Payment ou interno) | `"Canceled"` |

O status inicial do pedido ao ser criado continua sendo `"Pending"` (definido em `domain.NewOrder`).

### Exemplos de teste com grpcurl (Parte 3)

```bash
# Teste valor > 1000 (erro do Payment)
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"PROD1","quantity":1,"unit_price":1500}]}' \
  localhost:3000 Order/Create

# Teste mais de 50 itens (erro do Order)
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"PROD1","quantity":51,"unit_price":5}]}' \
  localhost:3000 Order/Create

# Teste pedido válido (deve retornar order_id e status "Paid")
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"PROD1","quantity":2,"unit_price":10}]}' \
  localhost:3000 Order/Create
```

---

## Comunicação Resiliente (Parte 4)

### Timeout por chamada

Cada chamada ao microsserviço Payment possui um **deadline individual de 2 segundos**, criado dentro da função `Charge` do adapter de pagamento:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
```

Se o serviço Payment não responder dentro desse prazo, a chamada é cancelada automaticamente pelo runtime do gRPC e um erro com código `DeadlineExceeded` é retornado. O adapter trata esse caso imprimindo uma mensagem de log específica:

```go
if status.Code(err) == codes.DeadlineExceeded {
    log.Printf("[Payment] Timeout (DeadlineExceeded) ao cobrar o pedido %d do cliente %d: %v",
        order.ID, order.CustomerID, err)
}
```

> **Nota sobre contextos encadeados:** quando um mesmo contexto é reutilizado em chamadas subsequentes, o deadline é compartilhado entre todas elas. Neste projeto optamos por criar um contexto novo a cada chamada, garantindo 2 segundos exclusivos para cada requisição ao Payment.

### Retransmissão automática (retry com backoff)

A conexão com o serviço Payment é configurada com um **interceptor de retry** via `go-grpc-middleware`. O interceptor é adicionado em `NewAdapter`, antes de `grpc.Dial`:

```go
grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
    grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
    grpc_retry.WithMax(5),
    grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
))
```

| Parâmetro | Valor | Significado |
|---|---|---|
| `WithCodes` | `Unavailable`, `ResourceExhausted` | Códigos de erro que disparam nova tentativa |
| `WithMax` | `5` | Número máximo de tentativas (incluindo a primeira) |
| `WithBackoff` | `BackoffLinear(1s)` | Janela de espera cresce 1 segundo a cada falha |

O backoff linear dessincroniza chamadas concorrentes de múltiplos clientes, evitando que todos retentem ao mesmo tempo e agravem a sobrecarga do servidor.

### Instalando a nova dependência

```bash
cd microservices/order
go get github.com/grpc-ecosystem/go-grpc-middleware@v1.4.0
go mod tidy
```

### Tabela consolidada de erros (Partes 3 e 4)

| Situação | Código gRPC | Status do pedido |
|---|---|---|
| Total de itens > 50 | `INVALID_ARGUMENT` | não salvo (rejeição antecipada) |
| Valor > 1000 (Payment) | `INVALID_ARGUMENT` | `"Canceled"` |
| Timeout na chamada ao Payment | `DEADLINE_EXCEEDED` | `"Canceled"` + log de aviso |
| Serviço Payment indisponível | `UNAVAILABLE` | até 5 retentativas; se esgotar, `"Canceled"` |
| Erro interno | `INTERNAL` | `"Canceled"` |
| Sucesso | — | `"Paid"` |

### Simulando falhas (Parte 4)

```bash
# Simular timeout: pare o serviço Payment e faça uma requisição normal.
# O Order aguardará 2 segundos, tentará até 5 vezes (Unavailable) e retornará erro.
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"PROD1","quantity":2,"unit_price":10}]}' \
  localhost:3000 Order/Create
```

---

## Tecnologias

- [Go](https://golang.org/) 1.22+
- [gRPC](https://grpc.io/)
- [Protocol Buffers v3](https://protobuf.dev/)
- [GORM](https://gorm.io/)
- [MySQL](https://www.mysql.com/)
- [Docker](https://www.docker.com/) (para os bancos de dados em desenvolvimento)
- [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware) **(Parte 4)**