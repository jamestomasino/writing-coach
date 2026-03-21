FROM golang:1.25.0 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

RUN GOBIN=/out go install github.com/errata-ai/vale/v3/cmd/vale@v3.7.1

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/writing-coach ./cmd/writing-coach

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /out/writing-coach /usr/local/bin/writing-coach
COPY --from=build /out/vale /usr/local/bin/vale
COPY .vale.ini /app/.vale.ini
COPY styles /app/styles
COPY migrations /app/migrations

EXPOSE 8080

CMD ["writing-coach", "serve"]
