FROM golang:1.23.2 AS builder

WORKDIR /build

COPY go.mod go.sum ./

COPY ./util/ ./util/

RUN go mod download

COPY ./cmd/ ./cmd/

WORKDIR /build/cmd

RUN CGO_ENABLED=0 go build -o ./main


FROM bash:devel-alpine3.21

WORKDIR /

COPY ./css/*.css ./css/

COPY ./views/*.html ./views/

COPY --from=builder /build/cmd/main ./cmd/main

COPY .env /

EXPOSE 8080

ENTRYPOINT ["./cmd/main"]
