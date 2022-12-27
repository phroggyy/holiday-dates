FROM golang:alpine as build

WORKDIR /app

ADD go.mod go.mod
ADD go.sum go.sum

RUN go mod download

ADD . .

RUN go build -o app main.go

FROM alpine:latest

WORKDIR /app
COPY --from=build /app/app /app/app

ENTRYPOINT ["/app/app"]