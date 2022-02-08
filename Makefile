proto:
	protoc \
	--go_out=spec/ \
	--go_opt=Mproto/textToSpeech.proto=. \
	 --go-grpc_out=. \
	  --go-grpc_opt=Mproto/textToSpeech.proto \
	  -I./ proto/textToSpeech.proto
