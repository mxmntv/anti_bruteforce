FROM golang:1.19-alpine3.17 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

FROM golang:1.19-alpine3.17 as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN go env -w CGO_ENABLED=0

CMD ["go", "test", "-v", "./integration-test/..."]