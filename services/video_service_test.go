package services

import (
	"archive/zip"
	"os"
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

func TestAddFileToZip_InvalidFile(t *testing.T) {
	os.MkdirAll("outputs", 0755)
	zipPath := "outputs/test_addfile.zip"
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo zip de teste: %v", err)
	}
	defer os.Remove(zipPath)
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = addFileToZip(zipWriter, "arquivo_inexistente.txt")
	if err == nil {
		t.Error("Esperado erro ao adicionar arquivo inexistente ao zip")
	}
}
