FROM golang:1.23.2 AS builder

WORKDIR /build

COPY go.mod go.sum ./

COPY ./services/ ./services/

COPY ./util/ ./util/

RUN go mod download

WORKDIR /build/services/event_store

RUN CGO_ENABLED=0 go build -o ./event_store


FROM bash:devel-alpine3.21

WORKDIR /

COPY --from=builder /build/services/event_store/event_store ./services/event_store/event_store

COPY .env /

ENTRYPOINT ["./services/event_store/event_store"]
