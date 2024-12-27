module github.com/gazinaft/distributed

go 1.23.2

require (
	github.com/crazy3lf/colorconv v1.2.0
	github.com/google/uuid v1.6.0
	github.com/labstack/echo/v4 v4.13.3
)

require github.com/rabbitmq/amqp091-go v1.10.0 // indirect

require (
	github.com/gazinaft/distributed/util v0.0.0
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.8.0 // indirect
)

replace github.com/gazinaft/distributed/util => ./util
