module github.com/tristan-club/wizard

go 1.16

require (
	github.com/bwmarrin/discordgo v0.25.1-0.20220804185119-c0803d021f34
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/kr/pretty v0.3.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/rs/zerolog v1.25.0
	github.com/shopspring/decimal v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/tristan-club/kit v0.0.0-20220727082038-26f086812c90
	github.com/trustwallet/go-primitives v0.0.35
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/net v0.0.0-20220418201149-a630d4f3e7a2
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	google.golang.org/genproto v0.0.0-20220414192740-2d67ff6cf2b4 // indirect
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/ethereum/go-ethereum v1.10.4 => github.com/ethereum/go-ethereum v1.9.21

//replace github.com/tristan-club/kit v0.0.1 => ../kit
