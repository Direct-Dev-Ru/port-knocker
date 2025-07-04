package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"crypto/aes"
	"crypto/cipher"

	"github.com/spf13/cobra"
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Расшифровать зашифрованный конфиг в открытый YAML",
	Long:  `Расшифровывает зашифрованный конфигурационный файл (ENCRYPTED:...) в обычный YAML-файл`,
	RunE:  runDecrypt,
}

var (
	decryptInputFile  string
	decryptOutputFile string
)

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().StringVarP(&decryptInputFile, "input", "i", "", "Входной зашифрованный файл")
	decryptCmd.Flags().StringVarP(&decryptOutputFile, "output", "o", "", "Выходной YAML-файл")
	decryptCmd.MarkFlagRequired("input")
	decryptCmd.MarkFlagRequired("output")
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(decryptInputFile)
	if err != nil {
		return fmt.Errorf("не удалось прочитать входной файл: %w", err)
	}

	if !strings.HasPrefix(string(data), "ENCRYPTED:") {
		return fmt.Errorf("файл не является зашифрованным (нет префикса ENCRYPTED:)")
	}

	key, err := getEncryptionKey(keyFile)
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

	fmt.Printf("Файл успешно расшифрован: %s\n", decryptOutputFile)
	return nil
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
