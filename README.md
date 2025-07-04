# Port Knocker

Утилита для отправки port knocking пакетов на удаленные серверы с поддержкой шифрования конфигурации.

**Версия**: 1.0.2

## Возможности

- ✅ Отправка TCP и UDP пакетов
- ✅ Настраиваемые последовательности портов
- ✅ Зашифрованные конфигурационные файлы
- ✅ Автоматическое определение зашифрованных файлов
- ✅ Ключи шифрования из файла или системной переменной
- ✅ Кроссплатформенная сборка (Linux, Windows, macOS)
- ✅ Совместимость со старыми версиями ОС (Ubuntu 18.04+)
- ✅ Подробный вывод для отладки
- ✅ **Расшифровка зашифрованных конфигов в открытый YAML (команда decrypt)**
- ✅ **Пасхалка для любознательных пользователей** 🎯

## Установка

### Сборка из исходников

```bash
# Клонировать репозиторий
git clone <repository-url>
cd port-knocker

# Установить зависимости
make deps

# Собрать для текущей платформы
make build

# Или собрать для всех платформ
make build-all
```

### Установка в систему

```bash
make install
```

## Использование

### Основная команда

```bash
# С файлом конфигурации
port-knocker -c config.yaml [-k key.txt] [-v]

# С инлайн целями
port-knocker -t "tcp:host:port;udp:host:port" [-d delay] [-v]
```

### Параметры

- `-c, --config` - Путь к файлу конфигурации
- `-t, --targets` - Инлайн цели в формате `[proto]:[host]:[port];[proto]:[host]:[port]`
- `-d, --delay` - Задержка между пакетами (по умолчанию 1s)
- `-k, --key` - Путь к файлу ключа шифрования
- `-v, --verbose` - Подробный вывод
- `-w, --wait-connection` - Ждать установления соединения

**Примечание**: Нужно указать либо `-c` (файл), либо `-t` (инлайн цели), но не оба одновременно.

### Шифрование конфигурации

```bash
# Используя глобальную опцию --config
port-knocker encrypt -c config.yaml -o config.encrypted -k key.txt

# Или используя опцию -i
port-knocker encrypt -i config.yaml -o config.encrypted -k key.txt
```

### **Расшифровка зашифрованного конфига**

```bash
# Используя глобальную опцию --config
port-knocker decrypt -c config.encrypted -o config.decrypted.yaml -k key.txt

# Или используя опцию -i
port-knocker decrypt -i config.encrypted -o config.decrypted.yaml -k key.txt
```

- `-c/--config` или `-i/--input` — путь к файлу (если не указан -i, используется --config)
- `-o/--output` — путь к выходному файлу
- `-k/--key` — путь к ключу (или используйте переменную окружения PORT_KNOCKER_KEY)

**Важно**: Ключ любой длины автоматически хешируется SHA256 до 32 байт для AES-256.

## Конфигурация

Конфигурационный файл должен быть в формате YAML:

```yaml
targets:
  - host: "192.168.1.100"
    ports: [1000, 2000, 3000]
    protocol: "tcp"
    delay: "1s"
  
  - host: "10.0.0.50"
    ports: [5000, 6000, 7000, 8000]
    protocol: "udp"
    delay: "500ms"
```

### Параметры цели

- `host` - IP-адрес или доменное имя цели
- `ports` - Массив портов для knocking
- `protocol` - Протокол: `tcp` или `udp`
- `delay` - Задержка между пакетами (например: `1s`, `500ms`, `2m`)

## Шифрование

### Создание ключа

Ключ может быть любой длины (автоматически хешируется до 32 байт):

```bash
# Создать ключ в файле (любая длина)
echo "my-secret-password" > key.txt

# Или установить системную переменную
export PORT_KNOCKER_KEY="my-secret-password"

# Можно использовать длинные пароли
echo "this-is-a-very-long-password-that-will-be-hashed-to-32-bytes" > key.txt
```

### Шифрование конфигурации

```bash
# Шифрование с ключом из файла
port-knocker encrypt -c config.yaml -o config.encrypted -k key.txt

# Шифрование с ключом из системной переменной
export PORT_KNOCKER_KEY="my-secret-password"
port-knocker encrypt -c config.yaml -o config.encrypted
```

### **Расшифровка зашифрованной конфигурации**

```bash
# Расшифровка с ключом из файла
port-knocker decrypt -c config.encrypted -o config.decrypted.yaml -k key.txt

# Расшифровка с ключом из системной переменной
export PORT_KNOCKER_KEY="my-secret-password"
port-knocker decrypt -c config.encrypted -o config.decrypted.yaml
```

### Использование зашифрованной конфигурации

```bash
# С ключом из файла
port-knocker -c config.encrypted -k key.txt -v

# С ключом из системной переменной
export PORT_KNOCKER_KEY="my-secret-key-32-bytes-long!!"
port-knocker -c config.encrypted -v
```

## Примеры

### Пример 1: Быстрое использование с инлайн целями

```bash
# Простая последовательность TCP портов
port-knocker -t "tcp:192.168.1.100:1000;tcp:192.168.1.100:2000;tcp:192.168.1.100:3000" -v

# Смешанные протоколы с настройкой задержки
port-knocker -t "tcp:server.com:22;udp:server.com:53;tcp:server.com:80" -d 500ms -v

# Одиночный порт
port-knocker -t "tcp:192.168.1.1:22" -v

# С ожиданием соединения
port-knocker -t "tcp:192.168.1.100:443" -w -v
```

### Пример 2: Конфигурационный файл

```bash
# Создать конфигурацию
cat > config.yaml << EOF
targets:
  - host: "192.168.1.100"
    ports: [1000, 2000, 3000]
    protocol: "tcp"
    delay: "1s"
EOF

# Запустить
port-knocker -c config.yaml -v
```

### Пример 3: Зашифрованная конфигурация

```bash
# Создать ключ
echo "my-secret-password" > key.txt

# Зашифровать конфигурацию
port-knocker encrypt -c config.yaml -o config.encrypted -k key.txt

# Использовать зашифрованную конфигурацию
port-knocker -c config.encrypted -k key.txt -v

# Расшифровать обратно для редактирования
port-knocker decrypt -c config.encrypted -o config.decrypted.yaml -k key.txt
```

### Пример 4: Множественные цели

```bash
cat > config.yaml << EOF
targets:
  - host: "server1.example.com"
    ports: [22, 80, 443]
    protocol: "tcp"
    delay: "2s"
  
  - host: "server2.example.com"
    ports: [5000, 6000, 7000, 8000]
    protocol: "udp"
    delay: "500ms"
EOF

port-knocker -c config.yaml -v
```

## 🎯 Пасхалка

Попробуйте найти скрытую функцию! Запустите:

```bash
port-knocker -t "tcp:8.8.8.8:8888" -v
```

И посмотрите, что произойдет! 🚀

## Совместимость

### Поддерживаемые системы

**Linux:**
- Ubuntu 18.04+ (GLIBC 2.27+)
- CentOS 7+ (GLIBC 2.17+)
- Debian 9+ (GLIBC 2.24+)
- RHEL 7+ (GLIBC 2.17+)

**Windows:**
- Windows 7+
- Windows Server 2012+

**macOS:**
- macOS 10.14+ (Mojave)
- macOS 11+ (Big Sur)
- macOS 12+ (Monterey)

### Сборка для старых систем

Для максимальной совместимости бинарники собираются на Ubuntu 18.04 с Go 1.20.

## Безопасность

- Ключи шифрования должны быть достаточно длинными и случайными
- Зашифрованные файлы имеют права доступа 600
- Системная переменная `PORT_KNOCKER_KEY` должна быть защищена
- Рекомендуется использовать файлы ключей вместо системных переменных в продакшене

## Сборка

### Для Linux

```bash
make build-linux
```

### Для Windows

```bash
make build-windows
```

### Для всех платформ

```bash
make build-all
```

## Разработка

### Запуск тестов

```bash
make test
```

### Очистка

```bash
make clean
```

### Справка

```bash
make help
```

## 📚 Дополнительная документация

Для более подробной информации о создании релизов и других аспектах проекта:

- **[Документация](docs/)** - Подробные инструкции и руководства
- **[Ручное создание релизов](docs/manual-release.md)** - Пошаговая инструкция
- **[Чек-лист релизов](docs/release-checklist.md)** - Быстрый чек-лист
- **[Скрипт быстрого релиза](docs/scripts/quick-release.sh)** - Автоматизация

## Лицензия

MIT License 