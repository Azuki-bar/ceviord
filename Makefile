dockerDebug:
	docker compose -f ./docker-compose.yml -f ./docker-compose.override.yml up --build

run:
	go run cmd/main.go

build:
	go build -ldflags '-X main.Version=$(VERSION)' -o ceviord cmd/main.go

test:
	go test -v ./...

pb: ttsProto
ttsProto:
	protoc \
	--go_out=spec/ \
	--go_opt=Mproto/textToSpeech.proto=. \
	--go-grpc_out=. \
	--go-grpc_opt=Mproto/textToSpeech.proto \
	-I./ proto/textToSpeech.proto

air: 
	air -c ./.air.toml
