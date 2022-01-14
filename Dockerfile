# syntax=docker/Dockerfile:1

FROM golang:1.17.6-alpine

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY internal/*.go ./

RUN go build -o /messagebox-server

CMD ["/messagebox-server"]
