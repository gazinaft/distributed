FROM golang:1.23.2 as builder

WORKDIR /build

COPY go.mod go.sum ./

COPY ./services/ ./services/

COPY ./util/ ./util/

RUN go mod download

WORKDIR /build/services/sync

RUN CGO_ENABLED=0 go build -o ./image-service


FROM bash:devel-alpine3.21

WORKDIR /

COPY --from=builder /build/services/sync/image-service ./services/sync/image-service

EXPOSE 8081

ENTRYPOINT ["./services/sync/image-service"]
