# DOCKERFILE SIMPLES (sem boas práticas - propositalmente!)
# Este é um exemplo de como NÃO fazer um Dockerfile

FROM golang:1.22-alpine AS builder

# Instalar ffmpeg
RUN apk add --no-cache ffmpeg

# Criar diretório de trabalho
WORKDIR /app

# Copiar arquivos
COPY . .

# Instalar dependências
RUN go mod tidy

# Build estático do binário
RUN go build -o app main.go

# Criar diretórios necessários
RUN mkdir -p uploads outputs temp

# Imagem final enxuta
FROM alpine:3.20
RUN apk add --no-cache ffmpeg
WORKDIR /app
COPY --from=builder /app/app /app/app
COPY --from=builder /app/uploads /app/uploads
COPY --from=builder /app/outputs /app/outputs
COPY --from=builder /app/temp /app/temp

# Expor porta
EXPOSE 8080

# Executar aplicação
CMD ["/app/app"]