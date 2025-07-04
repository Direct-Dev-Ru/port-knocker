package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"

	"github.com/spf13/cobra"
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Расшифровать зашифрованный конфиг в открытый YAML",
	Long:  `Расшифровывает зашифрованный конфигурационный файл (ENCRYPTED:...) в обычный YAML-файл`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Для команды decrypt config не обязателен если есть -i
		return nil
	},
	RunE: runDecrypt,
}

var (
	decryptInputFile  string
	decryptOutputFile string
)

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().StringVarP(&decryptInputFile, "input", "i", "", "Входной зашифрованный файл (если не указан, используется --config)")
	decryptCmd.Flags().StringVarP(&decryptOutputFile, "output", "o", "", "Выходной YAML-файл")
	decryptCmd.MarkFlagRequired("output")
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	// Определяем входной файл: либо из -i, либо из глобального --config
	input := decryptInputFile
	if input == "" {
		input = configFile
		if input == "" {
			return fmt.Errorf("необходимо указать входной файл через -i или --config")
		}
	}

	data, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("не удалось прочитать входной файл %s: %w", input, err)
	}

	if !strings.HasPrefix(string(data), "ENCRYPTED:") {
		return fmt.Errorf("файл %s не является зашифрованным (нет префикса ENCRYPTED:)", input)
	}

	key, err := getDecryptionKeyHashed(keyFile)
	if err != nil {
		return fmt.Errorf("не удалось получить ключ шифрования: %w", err)
	}

	decrypted, err := decryptData(data[10:], key)
	if err != nil {
		return fmt.Errorf("не удалось расшифровать данные: %w", err)
	}

	if err := os.WriteFile(decryptOutputFile, decrypted, 0600); err != nil {
		return fmt.Errorf("не удалось записать YAML-файл: %w", err)
	}

	fmt.Printf("Файл успешно расшифрован: %s → %s\n", input, decryptOutputFile)
	return nil
}

// getDecryptionKeyHashed получает ключ шифрования и хеширует его до 32 байт (аналогично encrypt)
func getDecryptionKeyHashed(keyFile string) ([]byte, error) {
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

// decryptData расшифровывает данные с помощью AES-GCM (аналогично internal)
func decryptData(encryptedData []byte, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("не удалось декодировать base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("данные слишком короткие")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось расшифровать: %w", err)
	}

	return plaintext, nil
}
