.PHONY: proto
proto:
	protoc \
		--go_out=. \
		--go_opt=module=github.com/kidrury/grpc-blog \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/kidrury/grpc-blog \
		proto/blog/blog.proto

.PHONY: certs
certs:
	openssl genrsa -out server.key 2048
	openssl req -new -x509 -key server.key -out server.crt -days 365 \
		-subj "/CN=localhost" \
		-addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

.PHONY: run-server
	go run server/main.go

# .PHONY: run-client
# 	go run 