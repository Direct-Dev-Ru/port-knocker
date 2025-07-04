package cmd

import (
	"fmt"

	"port-knocker/internal"

	"github.com/spf13/cobra"
)

var (
	configFile     string
	keyFile        string
	verbose        bool
	waitConnection bool
)

var rootCmd = &cobra.Command{
	Use:   "port-knocker",
	Short: "Утилита для отправки port knocking пакетов",
	Long: `Port Knocker - утилита для отправки TCP/UDP пакетов на определенные порты
в заданной последовательности для активации портов на удаленных серверах.

Поддерживает:
- TCP и UDP протоколы
- Зашифрованные конфигурационные файлы
- Автоматическое определение зашифрованных файлов
- Ключи шифрования из файла или системной переменной
- Настройка шлюза для отправки пакетов
- Гибкая настройка ожидания соединения`,
	RunE: runKnock,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Путь к файлу конфигурации")
	rootCmd.PersistentFlags().StringVarP(&keyFile, "key", "k", "", "Путь к файлу ключа шифрования")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Подробный вывод")
	rootCmd.PersistentFlags().BoolVarP(&waitConnection, "wait-connection", "w", false, "Ждать установления соединения (по умолчанию не ждать)")

	// НЕ делаем config глобально обязательным - проверяем в runKnock
}

func runKnock(cmd *cobra.Command, args []string) error {
	if configFile == "" {
		return fmt.Errorf("необходимо указать файл конфигурации")
	}

	knocker := internal.NewPortKnocker()
	return knocker.Execute(configFile, keyFile, verbose, waitConnection)
}
