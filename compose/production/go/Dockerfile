FROM golang:1.20.5 as base

FROM base as dev

WORKDIR /usr/src/app

RUN go install github.com/cosmtrek/air@latest

COPY . .
RUN go mod tidy