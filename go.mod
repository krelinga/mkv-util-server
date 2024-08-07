module github.com/krelinga/mkv-util-server

go 1.21.5

require (
	buf.build/gen/go/krelinga/proto/connectrpc/go v1.16.2-20240722193044-be7793236f15.1
	buf.build/gen/go/krelinga/proto/protocolbuffers/go v1.34.2-20240722200501-e4a8f516a168.2
	connectrpc.com/connect v1.16.2
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.6.0
	github.com/krelinga/kgo v0.1.0
	golang.org/x/net v0.23.0
	google.golang.org/protobuf v1.34.2
)

require golang.org/x/text v0.14.0 // indirect
