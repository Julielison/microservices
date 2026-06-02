module github.com/ruandg/microservices/order

go 1.22

require (
	github.com/ruandg/microservices-proto/golang/order v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.63.2
	gorm.io/driver/mysql v1.5.6
	gorm.io/gorm v1.25.9
)

replace github.com/ruandg/microservices-proto/golang/order => ../../microservices-proto/golang/order
