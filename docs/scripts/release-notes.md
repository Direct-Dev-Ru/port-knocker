# Port Knocker $VERSION

## Изменения

- Обновления и исправления
- Улучшения производительности
- Обновлена документация

## Установка

Скачайте соответствующий архив для вашей платформы:

- **Linux AMD64**: \`port-knocker-linux-amd64.tar.gz\`
- **Linux ARM64**: \`port-knocker-linux-arm64.tar.gz\`
- **Windows AMD64**: \`port-knocker-windows-amd64.exe.zip\`
- **macOS AMD64**: \`port-knocker-darwin-amd64.tar.gz\`
- **macOS ARM64**: \`port-knocker-darwin-arm64.tar.gz\`

### Использование

```bash

# Инлайн цели

./port-knocker -t \"tcp:host:port;udp:host:port\" -v

# Конфигурационный файл

./port-knocker -c config.yaml -v

# Пасхалка

./port-knocker -t \"tcp:8.8.8.8:8888\"

# Шутки

./port-knocker -t \"tcp:1.1.1.1:1111\"

```
