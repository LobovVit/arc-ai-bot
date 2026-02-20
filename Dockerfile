FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 go build -o /out/bot ./cmd/bot
RUN CGO_ENABLED=0 go build -o /out/ingest ./cmd/ingest
RUN CGO_ENABLED=0 go build -o /out/query ./cmd/query

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /out/bot /app/bin/bot
COPY --from=build /out/ingest /app/bin/ingest
COPY --from=build /out/query /app/bin/query
CMD ["/app/bin/bot"]