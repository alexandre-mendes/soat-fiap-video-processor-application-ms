package main

import (
	"log"
	"os"
	"video-processor/controllers"
	"video-processor/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Carregar variáveis do arquivo .env
	if err := utils.LoadEnv(".env"); err != nil {
		log.Printf("⚠️ Aviso: Não foi possível carregar .env: %v", err)
		log.Println("📄 Usando apenas variáveis de ambiente do sistema")
	} else {
		log.Println("✅ Arquivo .env carregado com sucesso")
	}

	createDirs()

	// Inicializar processador de mensagens
	if err := controllers.InitMessageProcessor(); err != nil {
		log.Printf("⚠️  Aviso: Não foi possível inicializar processador de mensagens: %v", err)
		log.Println("🔄 Continuando apenas com upload HTTP...")
	} else {
		log.Println("✅ Processador de mensagens SQS/S3 inicializado")
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Static("/uploads", "./uploads")
	r.Static("/outputs", "./outputs")

	r.GET("/", controllers.HandleHTML)
	r.GET("/health", controllers.HandleHealth)
	r.POST("/upload", controllers.HandleVideoUpload)
	r.GET("/download/:filename", controllers.HandleDownload)
	r.GET("/api/status", controllers.HandleStatus)

	// Rotas para processamento via mensageria
	r.POST("/api/process-message", controllers.HandleProcessMessage)
	r.GET("/api/message-processor/status", controllers.HandleMessageProcessorStatus)

	// Endpoint para métricas Prometheus
	r.GET("/metrics", controllers.HandleMetrics)

	port := ":" + utils.GetEnv("PORT", "8080")
	log.Printf("🎬 Servidor iniciado na porta %s", port[1:])
	log.Printf("📂 Acesse: http://localhost%s", port)
	log.Fatal(r.Run(port))
}

func createDirs() {
	dirs := []string{"uploads", "outputs", "temp"}
	for _, dir := range dirs {
		// Usando a função padrão do pacote os
		// Se quiser, pode mover para utils depois
		_ = os.MkdirAll(dir, 0755)
	}
}
