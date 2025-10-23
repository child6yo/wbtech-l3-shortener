FROM golang:1.24.5-alpine 

RUN apk add --no-cache git

WORKDIR /shortener

COPY go.mod go.sum ./
COPY ./ ./

RUN go mod tidy

RUN go build -o shortener ./cmd/main.go

CMD ["./shortener"]