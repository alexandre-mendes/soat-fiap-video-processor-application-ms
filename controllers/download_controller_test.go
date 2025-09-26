package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleDownload_FileNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "filename", Value: "arquivo_inexistente.zip"}}

	HandleDownload(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Esperado status 404, obtido %d", w.Code)
	}
}

func TestHandleDownload_Sucesso(t *testing.T) {
	filename := "arquivo_teste.zip"
	filePath := "outputs/" + filename
	os.MkdirAll("outputs", 0755)
	os.WriteFile(filePath, []byte("conteudo"), 0644)
	defer os.Remove(filePath)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "filename", Value: filename}}
	req, _ := http.NewRequest("GET", "/download/"+filename, nil)
	c.Request = req

	HandleDownload(c)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if w.Header().Get("Content-Disposition") == "" {
		t.Error("Esperado Content-Disposition para download")
	}
	if w.Header().Get("Content-Type") != "application/zip" {
		t.Errorf("Esperado Content-Type 'application/zip', obtido '%s'", w.Header().Get("Content-Type"))
	}
	if w.Body.Len() == 0 {
		t.Error("Esperado corpo não vazio para download")
	}
}
