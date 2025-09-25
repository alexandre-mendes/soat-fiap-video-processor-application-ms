#!/bin/bash

# Script para testar o sistema de processamento via SQS/S3
# Este script configura LocalStack e testa o fluxo completo

echo "🚀 Configurando ambiente de teste SQS/S3"

# Configurar variáveis de ambiente
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
export SQS_QUEUE_URL=http://localhost:4566/000000000000/video-processing-queue
export LOCALSTACK_URL=http://localhost:4566

echo "📦 Criando bucket S3..."
aws --endpoint-url=http://localhost:4566 s3 mb s3://video-bucket
aws --endpoint-url=http://localhost:4566 s3 mb s3://video-results

echo "📬 Criando filas SQS..."
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name video-processing-queue
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name video-results-queue

echo "📤 Upload de vídeo de teste para S3..."
# Assumindo que você tem um vídeo de teste
# aws --endpoint-url=http://localhost:4566 s3 cp test-video.mp4 s3://video-bucket/videos/

echo "📨 Enviando mensagem de teste para SQS..."
aws --endpoint-url=http://localhost:4566 sqs send-message \
  --queue-url http://localhost:4566/000000000000/video-processing-queue \
  --message-body '{"fileId": "videos/test-video.mp4", "processId": "test-process-123"}'

echo "✅ Configuração concluída!"
echo "🔧 Para testar via API:"
echo "curl -X POST http://localhost:8080/api/process-message \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"fileId\": \"videos/test-video.mp4\", \"processId\": \"test-process-123\"}'"
