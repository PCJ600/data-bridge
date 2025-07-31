FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.24.5-alpine3.22 AS builder

# need for gcc
RUN apk add --no-cache git build-base

ENV GOPROXY=https://goproxy.cn,direct \
    GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# confluent-kafka-go recommends musl tags for alpine linux builds
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags musl -ldflags="-s -w" -o /app/cloud-agent ./cmd/main.go

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/alpine:3.22.0

COPY --from=builder /app/cloud-agent /usr/local/bin/

EXPOSE 8080

CMD ["cloud-agent"]
