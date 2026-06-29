# microservices/order

Microsserviço **Order** — responsável por receber pedidos de compra, validar o estoque, orquestrar o pagamento e o envio. Implementado em **Go** com **arquitetura hexagonal** e comunicação via **gRPC**.

> **Parte Final:** o serviço Order agora também valida se os produtos existem no estoque e chama o microsserviço **Shipping** para agendar o envio após o pagamento.

---

## Arquitetura Hexagonal

```
                        ┌──────────────────────────────────────┐
          Driver Side   │                                      │   Driven Side
                        │                                      │
  ┌──────────┐          │  ┌──────────────────────────────┐   │          ┌──────────┐
  │  Cliente │──gRPC───▶│  │      Application Core        │   │──Save───▶│  MySQL   │
  │  gRPC    │          │  │                              │   │          │ (GORM)   │
  └──────────┘          │  │  domain/order.go             │   │          └──────────┘
                        │  │  api/api.go                  │   │
                        │  │                              │   │──Charge─▶┌──────────┐
                        │  └──────────────────────────────┘   │          │ Payment  │
                        │                                      │          │ (gRPC)   │
                        │       Ports (interfaces)             │          └──────────┘
                        │                                      │
                        │                                      │──Ship───▶┌──────────┐
                        │                                      │          │ Shipping │
                        └──────────────────────────────────────┘          │ (gRPC)   │
                                                                          └──────────┘
```

---

## Estrutura de Pastas

```
microservices/order/
├── cmd/
│   └── main.go                           # Ponto de entrada
├── config/
│   └── config.go                         # Variáveis de ambiente
├── internal/
│   ├── adapters/
│   │   ├── db/
│   │   │   └── db.go                     # Adapter MySQL — inclui tabela stock_items
│   │   ├── grpc/
│   │   │   └── server.go                 # Adapter gRPC de entrada
│   │   ├── payment/
│   │   │   └── payment.go                # Adapter gRPC → Payment
│   │   ├── shipping/
│   │   │   └── shipping.go               # (Parte Final) Adapter gRPC → Shipping
│   │   └── rest/
│   │       └── server.go                 # Placeholder REST
│   ├── application/
│   │   └── core/
│   │       ├── api/
│   │       │   └── api.go                # Lógica de negócio: PlaceOrder
│   │       └── domain/
│   │           └── order.go              # Entidades de domínio
│   └── ports/
│       ├── api.go                         # Interface APIPort
│       ├── db.go                          # Interface DBPort (inclui ProductExists)
│       ├── payment.go                     # Interface PaymentPort
│       └── shipping.go                    # (Parte Final) Interface ShippingPort
├── Dockerfile
└── go.mod
```

---

## Fluxo de uma Requisição (Parte Final)

```
Cliente gRPC
    │  CreateOrderRequest{costumer_id, order_items}
    ▼
adapters/grpc/server.go
    │  Converte para domain.Order → chama api.PlaceOrder(order)
    ▼
application/core/api/api.go
    │  1. Valida quantidade total ≤ 50
    │  2. Verifica se cada product_code existe em stock_items  ──▶ MySQL
    │  3. db.Save(&order) com status "Pending"                ──▶ MySQL
    │  4. payment.Charge(&order)                              ──▶ Payment gRPC
    │  5. shipping.Ship(&order)  (só se pagamento OK)         ──▶ Shipping gRPC
    │  6. db.Save com status "Paid" + delivery_deadline       ──▶ MySQL
    ▼
adapters/grpc/server.go
    │  Retorna CreateOrderResponse{order_id}
    ▼
Cliente gRPC
```

---

## Validações

| Regra | Código gRPC | Momento |
|---|---|---|
| Produto não existe no estoque | `NOT_FOUND` | antes de salvar |
| Total de itens > 50 | `INVALID_ARGUMENT` | antes de salvar |
| Valor total > R$ 1000 | `INVALID_ARGUMENT` | durante Payment |
| Timeout (2s) | `DEADLINE_EXCEEDED` | Payment / Shipping |
| Serviço indisponível | `UNAVAILABLE` | até 5 retentativas |

---

## Variáveis de Ambiente

| Variável              | Descrição                                   |
|-----------------------|---------------------------------------------|
| `ENV`                 | Ambiente (`development` / `production`)     |
| `DATA_SOURCE_URL`     | DSN do MySQL                                |
| `APPLICATION_PORT`    | Porta do servidor gRPC                      |
| `PAYMENT_SERVICE_URL` | Endereço `host:porta` do Payment            |
| `SHIPPING_SERVICE_URL`| Endereço `host:porta` do Shipping           |

---

## Tabela de Estoque (`stock_items`)

A tabela é criada automaticamente pelo GORM na primeira inicialização. Para popular produtos, execute o script `seed-stock.sql` disponível na raiz do repositório `microservices/`:

```bash
# Local
mysql -h 127.0.0.1 -u root -ps3cr3t < ../seed-stock.sql

# Docker Compose
docker compose exec -T mysql mysql -u root -ps3cr3t < seed-stock.sql

# Kubernetes
kubectl apply -f k8s/seed-stock.yaml
```

Produtos disponíveis por padrão após o seed:

| product_code | description | unit_price |
|---|---|---|
| `ABC123`  | Produto A | R$ 10,50 |
| `XYZ789`  | Produto B | R$ 20,00 |
| `PROD001` | Produto C | R$  5,00 |
| `PROD002` | Produto D | R$ 15,75 |
| `PROD003` | Produto E | R$ 99,99 |

---

## Dependências

| Pacote | Versão | Uso |
|---|---|---|
| `google.golang.org/grpc` | v1.81.1 | Framework gRPC |
| `github.com/grpc-ecosystem/go-grpc-middleware` | v1.4.0 | Retry com backoff linear |
| `gorm.io/gorm` | v1.25.9 | ORM |
| `gorm.io/driver/mysql` | v1.5.6 | Driver MySQL |
| `github.com/Julielison/microservices-proto/golang/order` | local | Stubs Order |
| `github.com/Julielison/microservices-proto/golang/payment` | local | Stubs Payment |
| `github.com/Julielison/microservices-proto/golang/shipping` | local | Stubs Shipping |

---

## Comunicação Resiliente

- **Deadline:** 2 segundos por chamada (Payment e Shipping)
- **Retry:** até 5 tentativas com backoff linear de 1 s nos erros `UNAVAILABLE` e `RESOURCE_EXHAUSTED`

```go
grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
    grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
    grpc_retry.WithMax(5),
    grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
))
```
