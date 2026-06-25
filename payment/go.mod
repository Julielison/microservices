module github.com/huseyinbabal/microservices/payment

go 1.24.5

require github.com/ruandg/microservices-proto/golang/payment v0.0.0-00010101000000-000000000000

require (
	github.com/sirupsen/logrus v1.9.0
	github.com/uptrace/opentelemetry-go-extra/otelgorm v0.3.2
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
	google.golang.org/grpc v1.74.2
	gorm.io/driver/mysql v1.3.2
	gorm.io/gorm v1.25.12
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.3.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/ruandg/microservices-proto/golang/payment => ../../microservices-proto/golang/payment
