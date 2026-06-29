module github.com/ruandg/microservices/order

go 1.25.0

require (
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/ruandg/microservices-proto/golang/order v0.0.0-00010101000000-000000000000
	github.com/ruandg/microservices-proto/golang/payment v0.0.0-00010101000000-000000000000
	github.com/ruandg/microservices-proto/golang/shipping v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.81.1
	gorm.io/driver/mysql v1.5.6
	gorm.io/gorm v1.25.9
)

require (
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/ruandg/microservices-proto/golang/order => ../../microservices-proto/golang/order
replace github.com/ruandg/microservices-proto/golang/payment => ../../microservices-proto/golang/payment
replace github.com/ruandg/microservices-proto/golang/shipping => ../../microservices-proto/golang/shipping
