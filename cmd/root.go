package cmd

import (	
	"fmt"
	"strconv"
	"strings"
	"time"

	"port-knocker/internal"

	"github.com/spf13/cobra"
)

var (
	configFile     string
	keyFile        string
	verbose        bool
	waitConnection bool
	targetsInline  string
	defaultDelay   string
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
- Гибкая настройка ожидания соединения
- Инлайн задание целей без конфигурационного файла`,
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
	rootCmd.PersistentFlags().StringVarP(&targetsInline, "targets", "t", "", "Инлайн цели в формате [proto]:[host]:[port];[proto]:[host]:[port]")
	rootCmd.PersistentFlags().StringVarP(&defaultDelay, "delay", "d", "1s", "Задержка между пакетами (по умолчанию 1s)")

	// НЕ делаем config глобально обязательным - проверяем в runKnock
}

func runKnock(cmd *cobra.Command, args []string) error {
	// Проверяем что указан либо config файл, либо инлайн цели
	if configFile == "" && targetsInline == "" {
		return fmt.Errorf("необходимо указать либо файл конфигурации (-c), либо инлайн цели (-t)")
	}

	if configFile != "" && targetsInline != "" {
		return fmt.Errorf("нельзя одновременно использовать файл конфигурации (-c) и инлайн цели (-t)")
	}

	knocker := internal.NewPortKnocker()

	// Если используем инлайн цели
	if targetsInline != "" {
		config, err := parseInlineTargets(targetsInline, defaultDelay)
		if err != nil {
			return fmt.Errorf("ошибка разбора инлайн целей: %w", err)
		}
		return knocker.ExecuteWithConfig(config, verbose, waitConnection)
	}

	// Иначе используем файл конфигурации
	return knocker.Execute(configFile, keyFile, verbose, waitConnection)
}

// parseInlineTargets разбирает строку инлайн целей в Config
func parseInlineTargets(targetsStr, delayStr string) (*internal.Config, error) {
	// Парсим задержку
	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		return nil, fmt.Errorf("неверная задержка '%s': %w", delayStr, err)
	}

	config := &internal.Config{
		Targets: []internal.Target{},
	}

	// Разбиваем по точкам с запятой
	targetParts := strings.Split(targetsStr, ";")

	for _, targetStr := range targetParts {
		targetStr = strings.TrimSpace(targetStr)
		if targetStr == "" {
			continue
		}

		// Разбираем формат [proto]:[host]:[port]
		parts := strings.Split(targetStr, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("неверный формат цели '%s', ожидается [proto]:[host]:[port]", targetStr)
		}

		protocol := strings.TrimSpace(parts[0])
		host := strings.TrimSpace(parts[1])
		portStr := strings.TrimSpace(parts[2])

		// Проверяем протокол
		if protocol != "tcp" && protocol != "udp" {
			return nil, fmt.Errorf("неподдерживаемый протокол '%s' в цели '%s'", protocol, targetStr)
		}

		// Парсим порт
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("неверный порт '%s' в цели '%s': %w", portStr, targetStr, err)
		}

		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("порт %d вне допустимого диапазона (1-65535) в цели '%s'", port, targetStr)
		}

		// Создаем цель
		target := internal.Target{
			Host:           host,
			Ports:          []int{port},
			Protocol:       protocol,
			Delay:          internal.Duration(delay),
			WaitConnection: false,
			Gateway:        "",
		}

		config.Targets = append(config.Targets, target)
	}

	if len(config.Targets) == 0 {
		return nil, fmt.Errorf("не найдено ни одной валидной цели")
	}

	return config, nil
}
