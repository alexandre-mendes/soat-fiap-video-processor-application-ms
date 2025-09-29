
# 🚀 SOAT-FIAP Video Processor Application Microservice

## Visão Geral
Microserviço Go para processamento de vídeos, arquitetura modular, observabilidade e pronto para produção em Docker/Kubernetes. Métricas expostas para Prometheus/Grafana.

---

## 🧩 Arquitetura

- **Controllers**: Endpoints HTTP REST (upload, download, status, health, métricas, HTML)
- **Services**: Lógica de negócio (processamento de vídeo, integração direta com AWS SQS/S3)
- **Models**: Estruturas de dados (requests, resultados, entidades de vídeo)
- **Utils**: Funções utilitárias (.env, helpers)
- **Testes**: Cobertura unitária/integrada com Go (arquivos *_test.go)
- **Observabilidade**: Métricas HTTP/sistema para Prometheus/Grafana
- **Deploy**: Docker/Kubernetes, CI/CD

---

## ✨ Funcionalidades

- Upload de vídeos: Recebe e armazena arquivos
- Processamento: Extrai frames, gera ZIP, integra com SQS/S3
- Download: Disponibiliza arquivos processados
- Métricas: Requisições HTTP, latência, status, `/metrics` para Prometheus
- Health Check: `/health` para disponibilidade
- Testes automatizados: Cobertura alta (>80%)
- Configuração via `.env`

---

## 🔗 Principais Endpoints

- `POST /upload` — Upload de vídeo
- `GET /download/:filename` — Download de arquivo
- `POST /api/process-message` — Processamento via SQS
- `GET /api/message-processor/status` — Status do processador
- `GET /metrics` — Métricas Prometheus
- `GET /health` — Health check
- `GET /` — Página HTML de upload

---

## 📊 Observabilidade

- Métricas HTTP: total, latência, status
- Métricas de sistema: CPU, memória
- Pronto para Prometheus/Grafana

---

## 🧪 Testes

- Executar: `go test ./...`
- Cobertura: `go test -coverprofile=coverage.out ./...`

---

## 🚢 Deploy

- Dockerfile para build local/produção
- Arquivos Kubernetes (`k8s/`) para produção

---

## ▶️ Como rodar

1. Instale dependências: `go mod tidy`
2. Configure variáveis em `.env`
3. Suba o ambiente: `docker-compose up` ou `go run main.go`
4. Acesse endpoints conforme documentação

---
