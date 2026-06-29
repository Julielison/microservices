FROM golang:1.25 AS builder
WORKDIR /usr/src/app

# Copia os módulos proto locais (necessários pelo replace)
COPY microservices-proto/ /microservices-proto/

COPY microservices/shipping/ .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shipping ./cmd/main.go

FROM scratch
COPY --from=builder /usr/src/app/shipping ./shipping
CMD ["./shipping"]
