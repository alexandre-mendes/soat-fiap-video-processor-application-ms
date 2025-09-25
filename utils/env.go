package utils

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadEnv carrega variáveis de ambiente do arquivo .env
func LoadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Ignorar linhas vazias e comentários
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Dividir key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remover aspas se existirem
		value = strings.Trim(value, "\"'")
		
		// Definir variável de ambiente se não existir
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	
	return scanner.Err()
}

// GetEnv retorna o valor da variável de ambiente ou o valor padrão
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvBool retorna o valor booleano da variável de ambiente
func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// GetEnvInt retorna o valor inteiro da variável de ambiente
func GetEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}
	
	return defaultValue
}

// GetEnvDuration retorna a duração da variável de ambiente
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	// Tentar parsing como segundos primeiro
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	
	// Tentar parsing como duração (ex: "5s", "2m", "1h")
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	
	return defaultValue
}
