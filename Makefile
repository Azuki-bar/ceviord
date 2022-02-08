build:
	go build -o ceviord cmd/main.go

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