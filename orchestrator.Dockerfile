FROM golang:1.23.2 as builder

WORKDIR /build

COPY go.mod go.sum ./

COPY ./services/ ./services/

COPY ./util/ ./util/

RUN go mod download

WORKDIR /build/services/orchestration

RUN CGO_ENABLED=0 go build -o ./orchestrator


FROM bash:devel-alpine3.21

WORKDIR /

COPY --from=builder /build/services/orchestration/orchestrator ./services/orchestration/orchestrator

ENTRYPOINT ["./services/orchestration/orchestrator"]
