FROM golang:1.21.3 AS builder

WORKDIR /build

COPY go.mod go.sum .

RUN go get ./...

RUN BIN="/usr/local/bin" && \
	VERSION="1.27.1" && \
	curl -sSL \
	"https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
	-o "${BIN}/buf" && \
	chmod +x "${BIN}/buf" && \
	apt-get update && \
	apt-get install -y ca-certificates curl gnupg && \
	mkdir -p /etc/apt/keyrings && \
	curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg && \
	echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_18.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list && \
	apt-get update && \
	apt-get install nodejs -y && \
	go install \
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
	google.golang.org/protobuf/cmd/protoc-gen-go \
	google.golang.org/grpc/cmd/protoc-gen-go-grpc

COPY . .

RUN go generate .

RUN CGO_ENABLED=0 go build ./cmd/proxy

RUN CGO_ENABLED=0 go build ./cmd/grpc

FROM debian:bullseye-slim AS master

WORKDIR /app

COPY --from=builder /build/proxy .

COPY --from=builder /build/grpc .

COPY docker/entrypoint.sh /docker-entrypoint.d/entrypoint.sh

ENTRYPOINT [ "/docker-entrypoint.d/entrypoint.sh" ]

CMD [ "proxy" ]
