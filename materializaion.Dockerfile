FROM golang:1.23.2 AS builder

WORKDIR /build

COPY go.mod go.sum ./

COPY ./services/ ./services/

COPY ./util/ ./util/

RUN go mod download

WORKDIR /build/services/materialization

RUN CGO_ENABLED=0 go build -o ./image-service-async


FROM bash:devel-alpine3.21

WORKDIR /

COPY --from=builder /build/services/materialization/materialization ./services/materialization/materialization

COPY .env /

ENTRYPOINT ["./services/materialization/materialization"]
