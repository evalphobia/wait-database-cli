FROM golang:1.21.7-bookworm as builder

WORKDIR /go/src/github.com/evalphobia/wait-mysql-cli
ENV GO111MODULE on
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o wait-mysql-cli

FROM gcr.io/distroless/base-debian12

COPY --from=builder /go/src/github.com/evalphobia/wait-mysql-cli/wait-mysql-cli /wait-mysql-cli

CMD ["/wait-mysql-cli"]
