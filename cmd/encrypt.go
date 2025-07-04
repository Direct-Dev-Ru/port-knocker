package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Зашифровать конфигурационный файл",
	Long:  `Зашифровывает YAML конфигурационный файл с помощью AES-GCM шифрования`,
	RunE:  runEncrypt,
}

func init() {
	rootCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Входной файл для шифрования")
	encryptCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Выходной зашифрованный файл")
	encryptCmd.MarkFlagRequired("input")
	encryptCmd.MarkFlagRequired("output")
	
	// Убираем обязательность флага config для команды encrypt
	encryptCmd.Flags().MarkHidden("config")
}

var (
	inputFile  string
	outputFile string
)

func runEncrypt(cmd *cobra.Command, args []string) error {
	// Читаем входной файл
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("не удалось прочитать входной файл: %w", err)
	}

	// Получаем ключ шифрования
	key, err := getEncryptionKey(keyFile)
	if err != nil {
		return fmt.Errorf("не удалось получить ключ шифрования: %w", err)
	}

	// Шифруем данные
	encryptedData, err := encrypt(data, key)
	if err != nil {
		return fmt.Errorf("не удалось зашифровать данные: %w", err)
	}

	// Записываем зашифрованный файл с префиксом "ENCRYPTED:"
	output := "ENCRYPTED:" + encryptedData
	if err := os.WriteFile(outputFile, []byte(output), 0600); err != nil {
		return fmt.Errorf("не удалось записать зашифрованный файл: %w", err)
	}

	fmt.Printf("Файл успешно зашифрован: %s\n", outputFile)
	return nil
}

// getEncryptionKey получает ключ шифрования из файла или системной переменной
func getEncryptionKey(keyFile string) ([]byte, error) {
	if keyFile != "" {
		// Читаем ключ из файла
		key, err := os.ReadFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать файл ключа: %w", err)
		}
		return key, nil
	}

	// Пытаемся получить ключ из системной переменной
	key := os.Getenv("PORT_KNOCKER_KEY")
	if key == "" {
		return nil, fmt.Errorf("ключ шифрования не найден ни в файле, ни в переменной PORT_KNOCKER_KEY")
	}

	return []byte(key), nil
}

// encrypt шифрует данные с помощью AES-GCM
func encrypt(plaintext []byte, key []byte) (string, error) {
	// Создаем AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("не удалось создать AES cipher: %w", err)
	}

	// Создаем GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("не удалось создать GCM: %w", err)
	}

	// Создаем nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("не удалось создать nonce: %w", err)
	}

	// Шифруем
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Кодируем в base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
