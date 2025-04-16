From golang:1.24-alpine As builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o product-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/product-service .
COPY --from=builder /app/.env .

EXPOSE 8000

CMD ["./product-service"]
