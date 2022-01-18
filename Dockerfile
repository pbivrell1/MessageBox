# syntax=docker/Dockerfile:1

FROM golang:1.17.6-alpine as base
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o messagebox-server .

FROM alpine:latest
WORKDIR /app
COPY --from=base /app/messagebox-server /app
CMD ["./messagebox-server"]
