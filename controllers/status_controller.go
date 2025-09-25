package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleStatus(c *gin.Context) {
	files, err := filepath.Glob(filepath.Join("outputs", "*.zip"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar arquivos"})
		return
	}

	var results []map[string]interface{}
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"filename":     filepath.Base(file),
			"size":         info.Size(),
			"created_at":   info.ModTime().Format("2006-01-02 15:04:05"),
			"download_url": "/download/" + filepath.Base(file),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"files": results,
		"total": len(results),
	})
}

// HandleHealth retorna o status de saúde da aplicação
func HandleHealth(c *gin.Context) {
	// Verificar se os diretórios essenciais existem
	dirs := []string{"uploads", "outputs", "temp"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "unhealthy",
				"message":   "Required directory missing: " + dir,
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"message":   "Service is running",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}
