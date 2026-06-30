# microservices

Sistema de comércio eletrônico baseado em **microsserviços** com comunicação via **gRPC** e arquitetura **hexagonal**.

![Arquitetura gRPC](docs/architecture.png)

> **Parte Final:** adicionado o microsserviço **Shipping**, validação de estoque no Order e deploy completo com **Docker Compose** e **Kubernetes**.

---

## Arquitetura

```
Cliente gRPC
    │
    ▼
┌─────────┐    gRPC    ┌─────────┐
│  Order  │──────────▶│ Payment │
│ :3000   │           │ :8081   │
│         │    gRPC    ├─────────┤
│         │──────────▶│Shipping │
└─────────┘           │ :8082   │
    │                 └─────────┘
    ▼
 MySQL
(order / payment / shipping)
```

**Fluxo de um pedido:**
1. Cliente envia `Order.Create` para o microsserviço Order
2. Order valida quantidade máxima de itens (≤ 50)
3. Order verifica se todos os `product_code` existem no estoque (tabela `stock_items`)
4. Order salva o pedido com status `"Pending"`
5. Order chama `Payment.Create` — se falhar, status → `"Canceled"`
6. Order chama `Shipping.Create` — se falhar, status → `"Canceled"`
7. Order atualiza status → `"Paid"` e salva o prazo de entrega
8. Retorna o `order_id` ao cliente

---

## Microsserviços

| Serviço  | Porta | Banco        | Responsabilidade                         |
|----------|-------|--------------|------------------------------------------|
| Order    | 3000  | `order`      | Recebe pedidos, orquestra Payment e Shipping |
| Payment  | 8081  | `payment`    | Realiza a cobrança do pedido             |
| Shipping | 8082  | `shipping`   | Agenda o envio e calcula o prazo         |

**Prazo de entrega (Shipping):** mínimo 1 dia + 1 dia a cada 5 unidades totais.

---

## Pré-requisitos

### Para execução local (sem container)
- Go 1.22+
- MySQL 8.0+
- [grpcurl](https://github.com/fullstorydev/grpcurl) (opcional, para testes)

### Para Docker Compose
- Docker 24+
- Docker Compose v2

### Para Kubernetes
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- Cluster local: [minikube](https://minikube.sigs.k8s.io/) ou [kind](https://kind.sigs.k8s.io/)

---

## Opção 1 — Execução local

### 1. Subir o MySQL

```bash
docker run -d --name mysql-local \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=s3cr3t \
  -v "$(pwd)/init.sql:/docker-entrypoint-initdb.d/init.sql" \
  mysql:8.0
```

### 2. Popular o estoque

```bash
# Aguarde o MySQL inicializar (~20s) e então execute:
mysql -h 127.0.0.1 -u root -ps3cr3t < seed-stock.sql
```

### 3. Iniciar o Payment

```bash
cd payment
DATA_SOURCE_URL="root:s3cr3t@tcp(127.0.0.1:3306)/payment?charset=utf8mb4&parseTime=True&loc=Local" \
APPLICATION_PORT=8081 \
ENV=development \
go run cmd/main.go
```

### 4. Iniciar o Shipping

```bash
cd shipping
DATA_SOURCE_URL="root:s3cr3t@tcp(127.0.0.1:3306)/shipping?charset=utf8mb4&parseTime=True&loc=Local" \
APPLICATION_PORT=8082 \
ENV=development \
go run cmd/main.go
```

### 5. Iniciar o Order

```bash
cd order
DATA_SOURCE_URL="root:s3cr3t@tcp(127.0.0.1:3306)/order?charset=utf8mb4&parseTime=True&loc=Local" \
APPLICATION_PORT=3000 \
PAYMENT_SERVICE_URL=localhost:8081 \
SHIPPING_SERVICE_URL=localhost:8082 \
ENV=development \
go run cmd/main.go
```

---

## Opção 2 — Docker Compose

> Execute todos os comandos a partir da raiz do repositório `microservices/`.

```bash
# Construir e subir todos os serviços
docker compose up --build

# (em outro terminal) Popular o estoque após os containers subirem
docker compose exec -T mysql mysql -u root -ps3cr3t < seed-stock.sql

# (Opcional) Verificar se o estoque foi populado corretamente
# Entre no MySQL do container:
docker compose exec mysql mysql -u root -ps3cr3t order
# Em seguida, no prompt do MySQL, execute:
# mysql> SELECT * FROM stock_items;
# mysql> exit

# Parar tudo
docker compose down -v
```

---

## Opção 3 — Kubernetes (com minikube)

### 1. Instalar e iniciar o minikube

```bash
# macOS
brew install minikube
# Linux (amd64)
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

minikube start --driver=docker
```

### 2. Configurar o Docker para usar o registry do minikube

```bash
eval $(minikube docker-env)
```

### 3. Construir as imagens dentro do minikube

> Execute a partir da raiz do repositório `microservices/`.

```bash
# Payment
docker build -t payment:latest ./payment

# Shipping
docker build -t shipping:latest ./shipping

# Order
docker build -t order:latest ./order
```

### 4. Aplicar os manifests

```bash
kubectl apply -f microservices/k8s/namespace.yaml
kubectl apply -f microservices/k8s/mysql.yaml
# Aguardar o MySQL ficar pronto
kubectl wait --namespace microservices \
  --for=condition=ready pod \
  --selector=app=mysql \
  --timeout=120s

kubectl apply -f microservices/k8s/payment.yaml
kubectl apply -f microservices/k8s/shipping.yaml
kubectl apply -f microservices/k8s/order.yaml
```

### 5. Popular o estoque

```bash
kubectl apply -f microservices/k8s/seed-stock.yaml
# Acompanhar o Job
kubectl logs -n microservices job/seed-stock
```

### 6. Verificar os pods

```bash
kubectl get pods -n microservices
# Todos devem estar Running/Completed
```

### 7. Testar via port-forward

```bash
kubectl port-forward -n microservices svc/order 3000:3000
```

Em outro terminal:

```bash
# Pedido válido
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"ABC123","quantity":2,"unit_price":10.50}]}' \
  localhost:3000 Order/Create

# Produto inexistente → NOT_FOUND
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"INVALIDO","quantity":1,"unit_price":5}]}' \
  localhost:3000 Order/Create

# Valor > 1000 → INVALID_ARGUMENT (Payment)
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"ABC123","quantity":1,"unit_price":1500}]}' \
  localhost:3000 Order/Create

# Mais de 50 itens → INVALID_ARGUMENT (Order)
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"ABC123","quantity":51,"unit_price":5}]}' \
  localhost:3000 Order/Create
```

### 8. Remover o cluster

```bash
kubectl delete namespace microservices
minikube stop
```

---

## Testes com grpcurl (resumo)

| Cenário | Comando | Resultado esperado |
|---|---|---|
| Pedido válido | `product_code` existente, qtd ≤ 50, total ≤ 1000 | `order_id` retornado; status `Paid` |
| Produto inexistente | `product_code` não cadastrado | `NOT_FOUND` |
| Muitos itens | quantidade total > 50 | `INVALID_ARGUMENT` |
| Valor alto | `unit_price * qty > 1000` | `INVALID_ARGUMENT` (Payment) |
| Payment offline | Payment não disponível | até 5 retentativas; `UNAVAILABLE` |

---

## Tratamento de Erros

| Situação | Código gRPC | Status do pedido |
|---|---|---|
| Produto não encontrado no estoque | `NOT_FOUND` | não salvo |
| Total de itens > 50 | `INVALID_ARGUMENT` | não salvo |
| Valor > R$ 1000 (Payment) | `INVALID_ARGUMENT` | `Canceled` |
| Timeout (Payment ou Shipping) | `DEADLINE_EXCEEDED` | `Canceled` + log |
| Serviço indisponível | `UNAVAILABLE` | até 5 retentativas; `Canceled` |
| Erro interno | `INTERNAL` | `Canceled` |
| Sucesso | — | `Paid` |

---

## Comunicação Resiliente

- **Deadline:** 2 segundos por chamada a Payment e Shipping
- **Retry automático:** até 5 tentativas nos códigos `UNAVAILABLE` e `RESOURCE_EXHAUSTED`, com backoff linear de 1 segundo

---

## Tecnologias

- [Go](https://golang.org/) 1.22+
- [gRPC](https://grpc.io/) + [Protocol Buffers v3](https://protobuf.dev/)
- [GORM](https://gorm.io/) + MySQL 8.0
- [Docker](https://www.docker.com/) + Docker Compose v2
- [Kubernetes](https://kubernetes.io/) + [minikube](https://minikube.sigs.k8s.io/)
- [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)
