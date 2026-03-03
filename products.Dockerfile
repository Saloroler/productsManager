FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/products ./cmd/products

FROM alpine:3.22

RUN adduser -D -g '' appuser
USER appuser
WORKDIR /app

COPY --from=build /out/products /app/products

EXPOSE 3000

ENTRYPOINT ["/app/products"]
