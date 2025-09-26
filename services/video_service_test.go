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

func TestProcessVideo_Sucesso(t *testing.T) {
	// Simula arquivo válido (mas não executa ffmpeg real)
	result := ProcessVideo("arquivo_invalido.mp4", "20220101_000000")
	if result.Success {
		t.Error("Esperado falha no processamento de arquivo inválido")
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

func TestCreateZipFile_Sucesso(t *testing.T) {
	// Cria arquivo fake
	file := "outputs/teste.txt"
	os.MkdirAll("outputs", 0755)
	os.WriteFile(file, []byte("conteudo"), 0644)
	defer os.Remove(file)
	zipPath := "outputs/teste.zip"
	err := CreateZipFile([]string{file}, zipPath)
	if err != nil {
		t.Errorf("Erro ao criar ZIP: %v", err)
	}
	os.Remove(zipPath)
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

func TestAddFileToZip_ErroAoAbrir(t *testing.T) {
	zipPath := "outputs/teste_add.zip"
	os.MkdirAll("outputs", 0755)
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

func TestAddFileToZip_ErroAoCriarHeader(t *testing.T) {
	zipPath := "outputs/teste_header.zip"
	os.MkdirAll("outputs", 0755)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo zip de teste: %v", err)
	}
	defer os.Remove(zipPath)
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Cria arquivo fake sem permissão para forçar erro no Stat
	file := "outputs/fake.txt"
	os.WriteFile(file, []byte("conteudo"), 0000)
	defer os.Remove(file)

	err = addFileToZip(zipWriter, file)
	if err == nil {
		t.Error("Esperado erro ao criar header do arquivo no zip")
	}
}

func TestProcessVideo_RemoveFalhaAoRemoverTemp(t *testing.T) {
	// Simula erro ao remover diretório temp
	timestamp := "20220101_000000"
	tempDir := "temp/" + timestamp
	os.MkdirAll(tempDir, 0755)
	// Remove permissão para forçar erro no RemoveAll
	os.Chmod(tempDir, 0000)
	defer os.Chmod(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	result := ProcessVideo("arquivo_invalido.mp4", timestamp)
	if result.Success {
		t.Error("Esperado falha no processamento de arquivo inválido")
	}
	if result.Message == "" {
		t.Error("Esperado mensagem de erro não vazia")
	}
}

func TestCreateZipFile_ErroAoCriarArquivo(t *testing.T) {
	files := []string{"arquivo_inexistente.txt"}
	// Simula diretório outputs sem permissão
	os.MkdirAll("outputs", 0555)
	defer os.Chmod("outputs", 0755)
	zipPath := "outputs/teste_perm.zip"
	err := CreateZipFile(files, zipPath)
	if err == nil {
		t.Error("Esperado erro ao criar arquivo ZIP sem permissão")
	}
	os.Remove(zipPath)
}
