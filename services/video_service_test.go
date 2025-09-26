package services

import (
	"testing"
)

func TestProcessVideo_InvalidFile(t *testing.T) {
	result := ProcessVideo("arquivo_invalido.mp4", "20220101_000000")
	if result.Success {
		t.Error("Esperado falha no processamento de arquivo inválido")
	}
	if result.Message == "" {
		t.Error("Esperado mensagem de erro não vazia")
	}
}

func TestCreateZipFile_Empty(t *testing.T) {
	zipPath := "outputs/test_empty.zip"
	files := []string{}
	err := CreateZipFile(files, zipPath)
	if err == nil {
		t.Error("Esperado erro ao criar ZIP vazio")
	}
}
