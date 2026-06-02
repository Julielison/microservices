module github.com/ruandg/microservices/order

go 1.22

require (
	github.com/ruandg/microservices-proto/golang/order v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.63.2
	gorm.io/driver/mysql v1.5.6
	gorm.io/gorm v1.25.9
)

require (
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/ruandg/microservices-proto/golang/order => ../../microservices-proto/golang/order
