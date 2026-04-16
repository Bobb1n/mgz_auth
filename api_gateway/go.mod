module api_gateway

go 1.25.3

require (
	auth_service v0.0.0
	github.com/S1FFFkA/user-mgz v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/joho/godotenv v1.5.1
	github.com/labstack/echo/v4 v4.15.1
	github.com/prometheus/client_golang v1.23.2
	gitlab.com/siffka/chat-message-mgz v0.0.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
	swipe-mgz v0.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
)

replace (
	auth_service => ../auth_service
	github.com/S1FFFkA/user-mgz => ../../user-mgz
	gitlab.com/siffka/chat-message-mgz => ../../chat-message-mgz
	swipe-mgz => ../../swipe-mgz
)
