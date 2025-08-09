FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /yankrun .

FROM alpine:3.20
RUN adduser -D -u 10001 appuser
USER appuser
COPY --from=builder /yankrun /usr/local/bin/yankrun
ENTRYPOINT ["yankrun"]
CMD ["--help"]


