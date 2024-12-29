FROM golang:1.23.2 AS builder

WORKDIR /build

COPY go.mod go.sum ./

COPY ./services/ ./services/

COPY ./util/ ./util/

RUN go mod download

WORKDIR /build/services/async

RUN CGO_ENABLED=0 go build -o ./image-service-async


FROM bash:devel-alpine3.21

WORKDIR /

COPY --from=builder /build/services/async/image-service-async ./services/async/image-service-async

COPY .env /

ENTRYPOINT ["./services/async/image-service-async"]
