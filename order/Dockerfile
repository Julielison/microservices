FROM golang:1.25 AS builder
WORKDIR /usr/src/app

COPY microservices-proto/ /microservices-proto/
COPY microservices/order/ .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o order ./cmd/main.go

FROM scratch
COPY --from=builder /usr/src/app/order ./order
CMD ["./order"]
