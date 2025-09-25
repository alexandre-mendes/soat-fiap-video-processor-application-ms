# Video Processor - Processamento de Vídeos via SQS/S3

Sistema de processamento de # 3. Testar via API
curl -X POST http://localhost:8080/api/process-message 
  -H 'Content-Type: application/json' 
  -d '{"fileId": "videos/test-video.mp4", "processId": "test-123"}'com suporte a upload HTTP e mensageria AWS (SQS/S3).

## Para rodar o projeto local

### Pré-requisitos
- Go 1.21 ou superior
- FFmpeg
- Docker (para LocalStack)
- AWS CLI (para testes)

### Instalação

1. **Instalar dependências Go:**
```bash
go mod tidy
```

2. **Configurar variáveis de ambiente:**
```bash
# Copiar arquivo de exemplo
cp .env.example .env

# Editar configurações conforme necessário
nano .env
```

3. **Instalar FFmpeg:**
```bash
sudo apt update
sudo apt install ffmpeg
```

4. **Instalar AWS CLI (para testes):**
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

### Executando o Sistema

#### Modo 1: Apenas Upload HTTP
```bash
go run main.go
```

#### Modo 2: Com SQS/S3 (LocalStack)
1. **Iniciar LocalStack:**
```bash
docker run --rm -it -p 4566:4566 -p 4510-4559:4510-4559 localstack/localstack
```

2. **Configurar ambiente de teste:**
```bash
chmod +x scripts/setup-test-env.sh
./scripts/setup-test-env.sh
```

3. **Iniciar aplicação:**
```bash
# As configurações serão lidas do arquivo .env automaticamente
go run main.go
```

#### Modo 3: Processador dedicado (apenas mensageria)
```bash
go run cmd/message-processor/main.go
```

## APIs Disponíveis

### Upload HTTP
- **POST** `/upload` - Upload de vídeo via formulário
- **GET** `/download/:filename` - Download do ZIP processado
- **GET** `/api/status` - Status dos arquivos processados

### Processamento via Mensageria
- **POST** `/api/process-message` - Processar vídeo do S3 diretamente
- **GET** `/api/message-processor/status` - Status do processador SQS

### Exemplos

**Processar vídeo do S3:**
```bash
curl -X POST http://localhost:8080/api/process-message \
  -H 'Content-Type: application/json' \
  -d '{"fileId": "videos/test-video.mp4", "processId": "test-process-123"}'
```

**Verificar status:**
```bash
curl http://localhost:8080/api/message-processor/status
```

## Configuração via .env

O projeto suporta arquivo `.env` para configuração, similar ao Node.js:

```bash
# Criar arquivo de configuração
cp .env.example .env

# Editar conforme necessário
nano .env
```

### Arquivo .env de exemplo:
```env
# LocalStack (desenvolvimento)
SQS_QUEUE_URL=http://localhost:4566/000000000000/video-processing-queue
LOCALSTACK_URL=http://localhost:4566
AWS_REGION=us-east-1
RESULTS_BUCKET=video-results
POLLING_INTERVAL_SECONDS=5
MAX_MESSAGES=10
PORT=8080

# Para AWS real (produção)
# SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789012/video-queue
# LOCALSTACK_URL=
# RESULTS_BUCKET=my-production-bucket
```

## Variáveis de Ambiente

### Principais:
- `SQS_QUEUE_URL`: URL da fila SQS para processamento
- `RESULTS_QUEUE_URL`: URL da fila SQS para resultados
- `LOCALSTACK_URL`: URL do LocalStack (deixe vazio para AWS real)  
- `AWS_REGION`: Região AWS (padrão: us-east-1)
- `SOURCE_BUCKET`: Bucket S3 de origem dos vídeos (padrão: video-bucket)
- `RESULTS_BUCKET`: Bucket S3 para ZIPs processados (padrão: video-results)

### Opcionais:
- `POLLING_INTERVAL_SECONDS`: Intervalo de polling SQS (padrão: 5)
- `MAX_MESSAGES`: Máximo de mensagens por lote (padrão: 10)
- `PORT`: Porta do servidor web (padrão: 8080)
- `DEBUG`: Modo debug (padrão: false)
- `LOG_LEVEL`: Nível de log (padrão: info)

## Estrutura do Projeto

```
├── cmd/message-processor/    # Processador dedicado
├── controllers/             # Handlers HTTP
├── services/               # Lógica de negócio
├── models/                 # Estruturas de dados
├── utils/                  # Utilitários
├── scripts/                # Scripts de setup
├── uploads/                # Vídeos temporários
├── outputs/                # ZIPs gerados
└── temp/                   # Frames temporários
```

## Fluxo de Processamento e Notificação

### 📥 **Mensagem de Entrada (SQS)**
```json
{
  "fileId": "videos/meu-video.mp4",
  "processId": "proc-2024-001"
}
```

### 📤 **Mensagem de Resultado (Fila de Resultados)**

**🔄 Início do Processamento:**
```json
{
  "processId": "proc-2024-001",
  "zipKey": "",
  "status": "IN_PROGRESS",
  "timestamp": "2024-08-27T19:42:30Z"
}
```

**✅ Processamento Concluído:**
```json
{
  "processId": "proc-2024-001",
  "zipKey": "processed/proc-2024-001_frames_20240827_194241.zip",
  "status": "COMPLETED",
  "timestamp": "2024-08-27T19:42:41Z"
}
```

**❌ Erro no Processamento:**
```json
{
  "processId": "proc-2024-001",
  "zipKey": "",
  "status": "FAILED",
  "timestamp": "2024-08-27T19:42:35Z"
}
```

### 🔄 **Estados Possíveis**
- **`IN_PROGRESS`**: Processamento iniciado
- **`COMPLETED`**: Processamento concluído com sucesso
- **`FAILED`**: Erro durante o processamento

### 🎯 **Integração com Microserviços**
O sistema enviará notificações para a fila `RESULTS_QUEUE_URL` sempre que:
- 🔄 Processamento for iniciado
- ✅ Processamento for concluído com sucesso
- ❌ Ocorrer erro durante download, processamento ou upload

### 📊 **Fluxo de Notificações:**

1. **🔄 Início**: `status: "IN_PROGRESS"` - Processamento iniciado
2. **✅ Sucesso**: `status: "COMPLETED"` - ZIP enviado para S3 com sucesso
3. **❌ Erro**: `status: "FAILED"` - Falha em qualquer etapa

**Exemplo de Consumo (outro microserviço):**
```bash
# Receber notificação de resultado
aws sqs receive-message \
  --queue-url http://localhost:4566/000000000000/video-results-queue \
  --endpoint-url http://localhost:4566
```

### 💡 **Exemplo de Implementação no Microserviço de Gerenciamento:**

```go
// Estrutura para tracking do status
type ProcessStatus struct {
    ProcessID string
    Status    string  // "IN_PROGRESS", "COMPLETED", "FAILED"
    ZipKey    string  // preenchido apenas no COMPLETED
    StartTime time.Time
    EndTime   *time.Time
}

// Handler para consumir resultados
func handleProcessingResult(result VideoProcessingResult) {
    switch result.Status {
    case "IN_PROGRESS":
        updateStatus(result.ProcessID, "EM_ANDAMENTO", "")
        
    case "COMPLETED":
        updateStatus(result.ProcessID, "CONCLUIDO", result.ZipKey)
        notifyUserSuccess(result.ProcessID, result.ZipKey)
        
    case "FAILED":
        updateStatus(result.ProcessID, "ERRO", "")
        notifyUserError(result.ProcessID)
    }
}
```