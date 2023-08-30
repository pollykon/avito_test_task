FROM golang:alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod verify

COPY . .
RUN go build -o service ./cmd/service
RUN go build -o crons ./cmd/crons/data_deleter
