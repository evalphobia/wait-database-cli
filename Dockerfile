FROM golang:1.21.7-bookworm

WORKDIR /go/src/github.com/evalphobia/wait-mysql-cli
ENV GO111MODULE on
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o wait-mysql-cli

CMD ["/wait-mysql-cli"]
