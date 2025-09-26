package controllers

import (
 "net/http"
 "net/http/httptest"
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

