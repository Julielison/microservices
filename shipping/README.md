# microservices/shipping

Microsserviço **Shipping** — responsável por agendar o envio de pedidos e calcular o prazo de entrega. Implementado em **Go** com **arquitetura hexagonal** e comunicação via **gRPC**.

---

## Cálculo do Prazo de Entrega

O prazo é calculado com base na **quantidade total de unidades** (soma de `quantity` de todos os itens):

```
deadline = 1 + floor(totalQuantidade / 5)
```

| Total de unidades | Prazo |
|---|---|
| 1 – 4   | 1 dia  |
| 5 – 9   | 2 dias |
| 10 – 14 | 3 dias |
| ...      | ...    |

---

## Arquitetura Hexagonal

```
                        ┌─────────────────────────────┐
          Driver Side   │                             │   Driven Side
                        │                             │
  ┌──────────┐          │  ┌──────────────────────┐   │          ┌──────────┐
  │  Order   │──gRPC───▶│  │   Application Core   │   │──Save───▶│  MySQL   │
  │ (client) │          │  │                      │   │          │ (GORM)   │
  └──────────┘          │  │  domain/shipping.go  │   │          └──────────┘
                        │  │  api/api.go          │   │
                        │  └──────────────────────┘   │
                        │                             │
                        └─────────────────────────────┘
```

---

## Estrutura de Pastas

```
microservices/shipping/
├── cmd/
│   └── main.go
├── config/
│   └── config.go
├── internal/
│   ├── adapters/
│   │   ├── db/
│   │   │   └── db.go          # Adapter MySQL via GORM
│   │   └── grpc/
│   │       └── server.go      # Adapter gRPC de entrada
│   ├── application/
│   │   └── core/
│   │       ├── api/
│   │       │   └── api.go     # Lógica de negócio: Ship
│   │       └── domain/
│   │           └── shipping.go # Entidades + CalculateDeadline
│   └── ports/
│       ├── api.go              # Interface APIPort
│       └── db.go               # Interface DBPort
├── Dockerfile
└── go.mod
```

---

## Variáveis de Ambiente

| Variável           | Descrição                                   |
|--------------------|---------------------------------------------|
| `ENV`              | Ambiente (`development` / `production`)     |
| `DATA_SOURCE_URL`  | DSN do MySQL (banco `shipping`)             |
| `APPLICATION_PORT` | Porta do servidor gRPC (padrão `8082`)      |

---

## Exemplo de Requisição (grpcurl)

```bash
grpcurl -plaintext \
  -d '{
    "order_id": 1,
    "shipping_items": [
      {"product_code": "ABC123", "quantity": 3},
      {"product_code": "XYZ789", "quantity": 4}
    ]
  }' \
  localhost:8082 Shipping/Create
```

Resposta esperada (7 unidades → 2 dias):

```json
{
  "shippingId": "1",
  "deliveryDeadline": 2
}
```

---

## Dependências

| Pacote | Versão | Uso |
|---|---|---|
| `google.golang.org/grpc` | v1.81.1 | Framework gRPC |
| `gorm.io/gorm` | v1.25.9 | ORM |
| `gorm.io/driver/mysql` | v1.5.6 | Driver MySQL |
| `github.com/Julielison/microservices-proto/golang/shipping` | local | Stubs gerados pelo protobuf |
