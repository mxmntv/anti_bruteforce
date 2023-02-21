FROM golang:1.19-alpine3.17 as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

FROM golang:1.19-alpine3.17 as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o /bin/app ./cmd/app
#RUN CGO_ENABLED=0 go build -o /bin/app ./cmd/cli

FROM scratch
COPY --from=builder /app/config /config
COPY --from=builder /bin/app /app
CMD ["/app"]