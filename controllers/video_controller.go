package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"video-processor/models"
	"video-processor/services"

	"github.com/gin-gonic/gin"
)

func HandleVideoUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ProcessingResult{
			Success: false,
			Message: "Erro ao receber arquivo: " + err.Error(),
		})
		return
	}
	defer file.Close()

	if !IsValidVideoFile(header.Filename) {
		c.JSON(http.StatusBadRequest, models.ProcessingResult{
			Success: false,
			Message: "Formato de arquivo não suportado. Use: mp4, avi, mov, mkv",
		})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s", timestamp, header.Filename)
	videoPath := filepath.Join("uploads", filename)

	out, err := os.Create(videoPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ProcessingResult{
			Success: false,
			Message: "Erro ao salvar arquivo: " + err.Error(),
		})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ProcessingResult{
			Success: false,
			Message: "Erro ao salvar arquivo: " + err.Error(),
		})
		return
	}

	result := services.ProcessVideo(videoPath, timestamp)

	if result.Success {
		os.Remove(videoPath)
	}

	c.JSON(http.StatusOK, result)
}

func IsValidVideoFile(filename string) bool {
	ext := filepath.Ext(filename)
	ext = stringLower(ext)
	validExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}
	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

func stringLower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
