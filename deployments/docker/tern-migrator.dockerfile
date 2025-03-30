FROM golang:1.23-alpine AS builder

RUN go install github.com/jackc/tern/v2@latest

COPY ../../migrations /migrations

ENTRYPOINT ["tern"]