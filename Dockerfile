# Build stage
FROM golang:1.25-alpine AS builder

# Instalar dependências necessárias para build
RUN apk add --no-cache git

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Download de dependências
RUN go mod download

# Copiar o código fonte (incluindo docs já gerados)
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Runtime stage
FROM alpine:latest

# Instalar ca-certificates para HTTPS e timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar o binário do build stage
COPY --from=builder /app/main .

# Expor porta da API
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]
