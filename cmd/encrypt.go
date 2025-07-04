package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Для команды encrypt config не обязателен если есть -i
		return nil
	},
	RunE: runEncrypt,
}

func init() {
	rootCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Входной файл для шифрования (если не указан, используется --config)")
	encryptCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Выходной зашифрованный файл")
	encryptCmd.MarkFlagRequired("output")
}

var (
	inputFile  string
	outputFile string
)

func runEncrypt(cmd *cobra.Command, args []string) error {
	// Определяем входной файл: либо из -i, либо из глобального --config
	input := inputFile
	if input == "" {
		input = configFile
		if input == "" {
			return fmt.Errorf("необходимо указать входной файл через -i или --config")
		}
	}

	// Читаем входной файл
	data, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("не удалось прочитать входной файл %s: %w", input, err)
	}

	// Получаем ключ шифрования
	key, err := getEncryptionKeyHashed(keyFile)
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

	fmt.Printf("Файл успешно зашифрован: %s → %s\n", input, outputFile)
	return nil
}

// getEncryptionKeyHashed получает ключ шифрования и хеширует его до 32 байт
func getEncryptionKeyHashed(keyFile string) ([]byte, error) {
	var rawKey []byte
	var err error

	if keyFile != "" {
		// Читаем ключ из файла
		rawKey, err = os.ReadFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать файл ключа: %w", err)
		}
	} else {
		// Пытаемся получить ключ из системной переменной
		key := os.Getenv("PORT_KNOCKER_KEY")
		if key == "" {
			return nil, fmt.Errorf("ключ шифрования не найден ни в файле, ни в переменной PORT_KNOCKER_KEY")
		}
		rawKey = []byte(key)
	}

	// Хешируем ключ SHA256 чтобы получить всегда 32 байта
	hash := sha256.Sum256(rawKey)
	return hash[:], nil
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
