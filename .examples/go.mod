module github.com/miyamo2/nrdeco/examples

go 1.24

require (
	github.com/google/wire v0.6.0
	github.com/joho/godotenv v1.5.1
	github.com/newrelic/go-agent/v3 v3.39.0
)

require (
	github.com/google/subcommands v1.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/miyamo2/nrdeco v0.0.0-00010101000000-000000000000 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/miyamo2/nrdeco => ../

tool (
	github.com/google/wire/cmd/wire
	github.com/miyamo2/nrdeco/cmd/nrdeco
)
