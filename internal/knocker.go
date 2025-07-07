package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed jokes.md
var jokesFile string

func GetRandomJoke() string {
	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

	jokes := strings.Split(jokesFile, "**********")

	var cleanJokes []string
	for _, joke := range jokes {
		if trimmed := strings.TrimSpace(joke); trimmed != "" {
			cleanJokes = append(cleanJokes, trimmed)
		}
	}

	if len(cleanJokes) == 0 {
		return "Шутки не найдены"
	}

	return cleanJokes[rand.Intn(len(cleanJokes))]
}

const (
	// Системная переменная для ключа шифрования
	EncryptionKeyEnvVar = "PORT_KNOCKER_KEY"
)

// Config представляет конфигурацию port knocking
type Config struct {
	Targets []Target `yaml:"targets"`
}

// Target представляет цель для port knocking
type Target struct {
	Host           string   `yaml:"host"`
	Ports          []int    `yaml:"ports"`
	Protocol       string   `yaml:"protocol"`        // "tcp" или "udp"
	Delay          Duration `yaml:"delay"`           // задержка между пакетами
	WaitConnection bool     `yaml:"wait_connection"` // ждать ли установления соединения
	Gateway        string   `yaml:"gateway"`         // шлюз для отправки (опционально)
}

// Duration для поддержки YAML десериализации времени
type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err != nil {
		return err
	}

	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	*d = Duration(duration)
	return nil
}

// PortKnocker основная структура для выполнения port knocking
type PortKnocker struct{}

// NewPortKnocker создает новый экземпляр PortKnocker
func NewPortKnocker() *PortKnocker {
	return &PortKnocker{}
}

// Execute выполняет port knocking на основе конфигурации
func (pk *PortKnocker) Execute(configFile, keyFile string, verbose bool, globalWaitConnection bool) error {
	// Читаем конфигурацию
	config, err := pk.loadConfig(configFile, keyFile)
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	return pk.ExecuteWithConfig(config, verbose, globalWaitConnection)
}

// ExecuteWithConfig выполняет port knocking с готовой конфигурацией
func (pk *PortKnocker) ExecuteWithConfig(config *Config, verbose bool, globalWaitConnection bool) error {
	if verbose {
		fmt.Printf("Загружена конфигурация с %d целей\n", len(config.Targets))
	}

	// Выполняем port knocking для каждой цели
	for i, target := range config.Targets {
		if verbose {
			fmt.Printf("Цель %d/%d: %s:%v (%s)\n", i+1, len(config.Targets), target.Host, target.Ports, target.Protocol)
		}

		// Применяем глобальный флаг если не задан локально
		if globalWaitConnection && !target.WaitConnection {
			target.WaitConnection = true
		}

		if err := pk.knockTarget(target, verbose); err != nil {
			return fmt.Errorf("ошибка при knocking цели %s: %w", target.Host, err)
		}
	}

	if verbose {
		fmt.Println("Port knocking завершен успешно")
	}
	return nil
}

// loadConfig загружает конфигурацию из файла с поддержкой шифрования
func (pk *PortKnocker) loadConfig(configFile, keyFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации: %w", err)
	}

	// Проверяем, зашифрован ли файл (начинается с "ENCRYPTED:")
	if strings.HasPrefix(string(data), "ENCRYPTED:") {
		fmt.Println("Обнаружен зашифрованный файл конфигурации")

		// Получаем ключ шифрования
		key, err := pk.getEncryptionKey(keyFile)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить ключ шифрования: %w", err)
		}

		// Расшифровываем данные
		decryptedData, err := pk.decrypt(data[10:], key) // пропускаем "ENCRYPTED:"
		if err != nil {
			return nil, fmt.Errorf("не удалось расшифровать конфигурацию: %w", err)
		}
		data = decryptedData
	}

	// Парсим YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("не удалось разобрать YAML: %w", err)
	}

	return &config, nil
}

// getEncryptionKey получает ключ шифрования из файла или системной переменной и хеширует его
func (pk *PortKnocker) getEncryptionKey(keyFile string) ([]byte, error) {
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
		key := os.Getenv(EncryptionKeyEnvVar)
		if key == "" {
			return nil, fmt.Errorf("ключ шифрования не найден ни в файле, ни в переменной %s", EncryptionKeyEnvVar)
		}
		rawKey = []byte(key)
	}

	// Хешируем ключ SHA256 чтобы получить всегда 32 байта для AES-256
	hash := sha256.Sum256(rawKey)
	return hash[:], nil
}

// decrypt расшифровывает данные с помощью AES-GCM
func (pk *PortKnocker) decrypt(encryptedData []byte, key []byte) ([]byte, error) {
	// Декодируем base64
	data, err := base64.StdEncoding.DecodeString(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("не удалось декодировать base64: %w", err)
	}

	// Создаем AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать AES cipher: %w", err)
	}

	// Создаем GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать GCM: %w", err)
	}

	// Извлекаем nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("данные слишком короткие")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Расшифровываем
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось расшифровать: %w", err)
	}

	return plaintext, nil
}

// knockTarget выполняет port knocking для одной цели
func (pk *PortKnocker) knockTarget(target Target, verbose bool) error {
	// Проверяем на "шутливую" цель 1
	if target.Host == "8.8.8.8" && len(target.Ports) == 1 && target.Ports[0] == 8888 {
		pk.showEasterEgg()
		return nil
	}

	// Проверяем на "шутливую" цель 2
	if target.Host == "1.1.1.1" && len(target.Ports) == 1 && target.Ports[0] == 1111 {
		pk.showRandomJoke()
		return nil
	}

	protocol := strings.ToLower(target.Protocol)
	if protocol != "tcp" && protocol != "udp" {
		return fmt.Errorf("неподдерживаемый протокол: %s", target.Protocol)
	}

	// Вычисляем таймаут как половину интервала между пакетами
	timeout := time.Duration(target.Delay) / 2
	if timeout < 100*time.Millisecond {
		timeout = 100 * time.Millisecond // минимальный таймаут
	}

	for i, port := range target.Ports {
		if verbose {
			fmt.Printf("  Отправка пакета на %s:%d (%s)\n", target.Host, port, protocol)
		}

		if err := pk.sendPacket(target.Host, port, protocol, target.WaitConnection, timeout, target.Gateway); err != nil {
			if target.WaitConnection {
				return fmt.Errorf("ошибка отправки пакета на порт %d: %w", port, err)
			} else {
				if verbose {
					fmt.Printf("  Предупреждение: не удалось отправить пакет на порт %d: %v\n", port, err)
				}
			}
		}

		// Задержка между пакетами (кроме последнего)
		if i < len(target.Ports)-1 {
			delay := time.Duration(target.Delay)
			if delay > 0 {
				if verbose {
					fmt.Printf("  Ожидание %v...\n", delay)
				}
				time.Sleep(delay)
			}
		}
	}

	return nil
}

// sendPacket отправляет один пакет на указанный хост и порт
func (pk *PortKnocker) sendPacket(host string, port int, protocol string, waitConnection bool, timeout time.Duration, gateway string) error {
	address := fmt.Sprintf("%s:%d", host, port)

	var conn net.Conn
	var err error

	// Настройка локального адреса если указан шлюз
	var localAddr net.Addr
	if gateway != "" {
		if strings.Contains(gateway, ":") {
			localAddr, err = net.ResolveTCPAddr("tcp", gateway)
			if err != nil {
				return fmt.Errorf("не удалось разрешить адрес шлюза %s: %w", gateway, err)
			}
		} else {
			// Если указан только IP, добавляем порт 0
			localAddr, err = net.ResolveTCPAddr("tcp", gateway+":0")
			if err != nil {
				return fmt.Errorf("не удалось разрешить адрес шлюза %s: %w", gateway, err)
			}
		}
	}

	switch protocol {
	case "tcp":
		if localAddr != nil {
			dialer := &net.Dialer{
				LocalAddr: localAddr,
				Timeout:   timeout,
			}
			conn, err = dialer.Dial("tcp", address)
		} else {
			conn, err = net.DialTimeout("tcp", address, timeout)
		}
	case "udp":
		if localAddr != nil {
			dialer := &net.Dialer{
				LocalAddr: localAddr,
				Timeout:   timeout,
			}
			conn, err = dialer.Dial("udp", address)
		} else {
			conn, err = net.DialTimeout("udp", address, timeout)
		}
	default:
		return fmt.Errorf("неподдерживаемый протокол: %s", protocol)
	}

	if err != nil {
		if waitConnection {
			return fmt.Errorf("не удалось подключиться к %s: %w", address, err)
		} else {
			// Для UDP и TCP без ожидания соединения просто отправляем пакет
			return pk.sendPacketWithoutConnection(host, port, protocol, localAddr)
		}
	}
	defer conn.Close()

	// Отправляем пустой пакет
	_, err = conn.Write([]byte{})
	if err != nil {
		return fmt.Errorf("не удалось отправить пакет: %w", err)
	}

	return nil
}

// sendPacketWithoutConnection отправляет пакет без установления соединения
func (pk *PortKnocker) sendPacketWithoutConnection(host string, port int, protocol string, localAddr net.Addr) error {
	address := fmt.Sprintf("%s:%d", host, port)

	switch protocol {
	case "udp":
		// Для UDP просто отправляем пакет
		var conn net.Conn
		var err error

		if localAddr != nil {
			dialer := &net.Dialer{
				LocalAddr: localAddr,
			}
			conn, err = dialer.Dial("udp", address)
		} else {
			conn, err = net.Dial("udp", address)
		}

		if err != nil {
			return fmt.Errorf("не удалось создать UDP соединение к %s: %w", address, err)
		}
		defer conn.Close()

		_, err = conn.Write([]byte{})
		if err != nil {
			return fmt.Errorf("не удалось отправить UDP пакет: %w", err)
		}

	case "tcp":
		// Для TCP без ожидания соединения используем короткий таймаут
		var conn net.Conn
		var err error

		if localAddr != nil {
			dialer := &net.Dialer{
				LocalAddr: localAddr,
				Timeout:   100 * time.Millisecond,
			}
			conn, err = dialer.Dial("tcp", address)
		} else {
			conn, err = net.DialTimeout("tcp", address, 100*time.Millisecond)
		}

		if err != nil {
			// Для TCP без ожидания соединения игнорируем ошибки подключения
			return nil
		}
		defer conn.Close()

		_, err = conn.Write([]byte{})
		if err != nil {
			return fmt.Errorf("не удалось отправить TCP пакет: %w", err)
		}
	}

	return nil
}

// showEasterEgg показывает забавный ASCII-арт
func (pk *PortKnocker) showEasterEgg() {
	fmt.Println("\n🎯 🎯 🎯  EASTER EGG ACTIVATED! 🎯 🎯 🎯")
	fmt.Println()

	// Анимированный ASCII-арт
	frames := []string{
		`
    ╭─────────────────╮
    │   🚀 PORT       │
    │   KNOCKER       │
    │   🎯 1.0.1      │
    │                 │
    │   🎮 GAME ON!   │
    ╰─────────────────╯
`,
		`
    ╭─────────────────╮
    │   🚀 PORT       │
    │   KNOCKER       │
    │   🎯 1.0.1      │
    │                 │
    │   🎯 BULLSEYE!  │
    ╰─────────────────╯
`,
		`
    ╭─────────────────╮
    │   🚀 PORT       │
    │   KNOCKER       │
    │   🎯 1.0.1      │
    │                 │
    │   🎪 MAGIC!     │
    ╰─────────────────╯
`,
	}

	for i := 0; i < 3; i++ {
		fmt.Print("\033[2J\033[H") // Очистка экрана
		fmt.Println(frames[i%len(frames)])
		time.Sleep(1500 * time.Millisecond)
	}

	fmt.Println("\n🎉 Поздравляем! Вы нашли пасхалку!")
	fmt.Println("🎯 Попробуйте: ./port-knocker -t \"tcp:8.8.8.8:8888\"")
	fmt.Println("🚀 Port Knocker - теперь с пасхалками!")
	fmt.Println()
}

func (pk *PortKnocker) showRandomJoke() {
	joke := GetRandomJoke()

	// ANSI цветовые коды
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
		colorWhite  = "\033[37m"
		colorBold   = "\033[1m"
	)

	// Функция для подсчета видимой длины строки (без ANSI кодов) в рунах
	visibleLength := func(s string) int {
		// Удаляем ANSI escape последовательности
		clean := s
		for strings.Contains(clean, "\033[") {
			start := strings.Index(clean, "\033[")
			end := strings.Index(clean[start:], "m")
			if end == -1 {
				break
			}
			clean = clean[:start] + clean[start+end+1:]
		}
		// Возвращаем количество рун, а не байт
		return len([]rune(clean))
	}

	// Функция для умного разбиения строки
	splitLine := func(line string, maxWidth int) []string {
		runes := []rune(line)
		if len(runes) <= maxWidth {
			return []string{line}
		}

		var result []string
		remaining := line

		for len([]rune(remaining)) > maxWidth {
			// Ищем позицию для разрыва в пределах maxWidth
			breakPos := maxWidth
			remainingRunes := []rune(remaining)

			for i := maxWidth; i >= 0; i-- {
				if i < len(remainingRunes) {
					char := remainingRunes[i]
					// Разрываем на пробеле, знаке пунктуации или в конце строки
					if char == ' ' || char == ',' || char == '.' || char == '!' ||
						char == '?' || char == ':' || char == ';' || char == '-' {
						breakPos = i + 1
						break
					}
				}
			}

			// Если не нашли подходящего места, разрываем по maxWidth
			if breakPos == maxWidth {
				breakPos = maxWidth
			}

			// Создаем строку из рун
			breakString := string(remainingRunes[:breakPos])
			result = append(result, strings.TrimSpace(breakString))
			remaining = strings.TrimSpace(string(remainingRunes[breakPos:]))
		}

		if len([]rune(remaining)) > 0 {
			result = append(result, remaining)
		}

		return result
	}

	// Разбиваем исходную шутку на строки
	originalLines := strings.Split(joke, "\n")

	// Обрабатываем каждую строку и разбиваем длинные
	var processedLines []string
	for _, line := range originalLines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		splitLines := splitLine(line, 80)
		processedLines = append(processedLines, splitLines...)
	}

	// Находим максимальную длину строки для рамки (в рунах)
	maxLength := 0
	for _, line := range processedLines {
		lineLength := len([]rune(line))
		if lineLength > maxLength {
			maxLength = lineLength
		}
	}

	// Убеждаемся, что maxLength не меньше минимальной ширины для заголовков
	minWidth := 60 // Минимальная ширина для заголовков
	if maxLength < minWidth {
		maxLength = minWidth
	}

	fmt.Println()
	fmt.Printf("%s%s╭%s", colorPurple, colorBold, colorReset)
	fmt.Printf("%s%s", colorYellow, strings.Repeat("─", maxLength+2))
	fmt.Printf("%s%s╮%s\n", colorPurple, colorBold, colorReset)

	headerText := " Зацени Анектотец! 🤣 "
	fmt.Printf("%s%s│%s", colorPurple, colorBold, colorReset)
	fmt.Printf("%s%s%s%s", colorCyan, colorBold, headerText, colorReset)
	fmt.Printf("%s%s", colorYellow, strings.Repeat(" ", 1+maxLength-visibleLength(headerText)))
	fmt.Printf("%s%s│%s\n", colorPurple, colorBold, colorReset)

	fmt.Printf("%s%s├%s", colorPurple, colorBold, colorReset)
	fmt.Printf("%s%s", colorYellow, strings.Repeat("─", maxLength+2))
	fmt.Printf("%s%s┤%s\n", colorPurple, colorBold, colorReset)

	// Выводим обработанные строки шутки
	for _, line := range processedLines {
		fmt.Printf("%s%s│%s", colorPurple, colorBold, colorReset)
		fmt.Printf("%s%s%s", colorWhite, line, colorReset)
		fmt.Printf("%s%s", colorYellow, strings.Repeat(" ", 2+maxLength-len([]rune(line))))
		fmt.Printf("%s%s│%s\n", colorPurple, colorBold, colorReset)
	}

	fmt.Printf("%s%s├%s", colorPurple, colorBold, colorReset)
	fmt.Printf("%s%s", colorYellow, strings.Repeat("─", maxLength+2))
	fmt.Printf("%s%s┤%s\n", colorPurple, colorBold, colorReset)

	// Вычисляем правильную ширину для нижних строк
	cmdText := "Попробуйте: ./port-knocker -t \"tcp:1.1.1.1:1111\""
	titleText := "🚀 Port Knocker - теперь с шутками! 🤣"

	fmt.Printf("%s%s│%s", colorPurple, colorBold, colorReset)
	fmt.Printf("%s%s%s%s", colorGreen, colorBold, cmdText, colorReset)
	fmt.Printf("%s%s", colorYellow, strings.Repeat(" ", 2+maxLength-visibleLength(cmdText)))
	fmt.Printf("%s%s│%s\n", colorPurple, colorBold, colorReset)

	fmt.Printf("%s%s│%s", colorPurple, colorBold, colorReset)
	fmt.Printf("%s%s%s%s", colorBlue, colorBold, titleText, colorReset)
	fmt.Printf("%s%s", colorYellow, strings.Repeat(" ", maxLength-visibleLength(titleText)))
	fmt.Printf("%s%s│%s\n", colorPurple, colorBold, colorReset)

	fmt.Printf("%s%s╰%s", colorPurple, colorBold, colorReset)
	fmt.Printf("%s%s", colorYellow, strings.Repeat("─", maxLength+2))
	fmt.Printf("%s%s╯%s\n", colorPurple, colorBold, colorReset)
	fmt.Println()
}
