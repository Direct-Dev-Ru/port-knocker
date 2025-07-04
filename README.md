# Port Knocker

Утилита для отправки port knocking пакетов на удаленные серверы с поддержкой шифрования конфигурации.

## Возможности

- ✅ Отправка TCP и UDP пакетов
- ✅ Настраиваемые последовательности портов
- ✅ Зашифрованные конфигурационные файлы
- ✅ Автоматическое определение зашифрованных файлов
- ✅ Ключи шифрования из файла или системной переменной
- ✅ Кроссплатформенная сборка (Linux, Windows)
- ✅ Подробный вывод для отладки
- ✅ **Расшифровка зашифрованных конфигов в открытый YAML (команда decrypt)**

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
port-knocker -c config.yaml [-k key.txt] [-v]
```

### Параметры

- `-c, --config` - Путь к файлу конфигурации (обязательно)
- `-k, --key` - Путь к файлу ключа шифрования
- `-v, --verbose` - Подробный вывод

### Шифрование конфигурации

```bash
port-knocker encrypt -i config.yaml -o config.encrypted -k key.txt
```

### **Расшифровка зашифрованного конфига**

```bash
port-knocker decrypt -i config.encrypted -o config.decrypted.yaml -k key.txt
```

- `-i` — путь к зашифрованному файлу (ENCRYPTED:...)
- `-o` — путь к выходному YAML-файлу
- `-k` — путь к ключу (или используйте переменную окружения PORT_KNOCKER_KEY)

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

Ключ должен быть ровно 32 байта для AES-256:

```bash
# Создать ключ в файле
echo "my-secret-key-32-bytes-long!!" > key.txt

# Или установить системную переменную
export PORT_KNOCKER_KEY="my-secret-key-32-bytes-long!!"
```

### Шифрование конфигурации

```bash
# Шифрование с ключом из файла
port-knocker encrypt -i config.yaml -o config.encrypted -k key.txt

# Шифрование с ключом из системной переменной
export PORT_KNOCKER_KEY="my-secret-key-32-bytes-long!!"
port-knocker encrypt -i config.yaml -o config.encrypted
```

### **Расшифровка зашифрованной конфигурации**

```bash
# Расшифровка с ключом из файла
port-knocker decrypt -i config.encrypted -o config.decrypted.yaml -k key.txt

# Расшифровка с ключом из системной переменной
export PORT_KNOCKER_KEY="my-secret-key-32-bytes-long!!"
port-knocker decrypt -i config.encrypted -o config.decrypted.yaml
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

### Пример 1: Простой knocking

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

### Пример 2: Зашифрованная конфигурация

```bash
# Создать ключ
echo "my-secret-key-32-bytes-long!!" > key.txt

# Зашифровать конфигурацию
port-knocker encrypt -i config.yaml -o config.encrypted -k key.txt

# Использовать зашифрованную конфигурацию
port-knocker -c config.encrypted -k key.txt -v

# Расшифровать обратно для редактирования
port-knocker decrypt -i config.encrypted -o config.decrypted.yaml -k key.txt
```

### Пример 3: Множественные цели

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

## Лицензия

MIT License 