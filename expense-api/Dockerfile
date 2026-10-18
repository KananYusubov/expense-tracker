# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Pure Go binary (don't need CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o expense-api ./cmd/api

# Runtime stage
FROM alpine:3.19

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
RUN mkdir -p /app/data

COPY --from=builder /app/expense-api .

ENV PORT=8080
ENV DB_PATH=/app/data/expenses.db

EXPOSE 8080

CMD ["./expense-api"]