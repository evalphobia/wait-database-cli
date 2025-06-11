FROM golang:1.24.4-bookworm AS builder

WORKDIR /go/src/github.com/evalphobia/wait-database-cli
ENV GO111MODULE=on
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o wait-database-cli

FROM gcr.io/distroless/base-debian12

COPY --from=builder /go/src/github.com/evalphobia/wait-database-cli/wait-database-cli /wait-database-cli

CMD ["/wait-database-cli"]
