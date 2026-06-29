FROM golang:1.25 AS builder
WORKDIR /workspace

# Copia o código do serviço
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o order ./cmd/main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/order /order
ENTRYPOINT ["/order"]