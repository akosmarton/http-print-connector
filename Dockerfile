FROM golang:1.25.4 as builder

WORKDIR /app
COPY . .
RUN --mount=type=cache,target=$HOME/.cache/go-build CGO_ENABLED=0 go build -v

FROM alpine

COPY --from=builder /app/http-print-connector /app/http-print-connector

ENV API_URL= API_KEY= PRINTER_URL=

ENTRYPOINT [ "/app/http-print-connector" ]
