# Ручное создание релизов

Эта инструкция описывает процесс ручного создания релизов для Port Knocker без использования GitHub Actions.

## Предварительные требования

1. **Go 1.21+** установлен и настроен
2. **Git** настроен с доступом к репозиторию
3. **make** утилита (опционально, но рекомендуется)
4. **zip** и **tar** для создания архивов

## Шаг 1: Подготовка к релизу

### 1.1 Обновление версии

Перед созданием релиза убедитесь, что версия обновлена в коде:

```bash
# Проверить текущую версию в main.go
grep "Version.*=" main.go

# Если нужно обновить версию, отредактируйте main.go
# Например, для версии 1.0.2:
# var (
#     Version   = "1.0.2"
#     BuildTime = "unknown"
# )
```

### 1.2 Обновление документации

Обновите версию в README.md:

```bash
# Найти и заменить версию в README.md
sed -i 's/Версия.*: [0-9.]*/Версия**: 1.0.2/' README.md
```

### 1.3 Проверка изменений

```bash
# Проверить статус репозитория
git status

# Посмотреть последние коммиты
git log --oneline -10

# Убедиться, что все изменения закоммичены
git add .
git commit -m "Prepare for release v1.0.2"
git push origin main
```

## Шаг 2: Сборка бинарников

### 2.1 Сборка для всех платформ

```bash
# Установить переменные окружения
export VERSION="1.0.2"
export BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Сборка для Linux AMD64
GOOS=linux GOARCH=amd64 go build \
  -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
  -o port-knocker-linux-amd64 .

# Сборка для Linux ARM64
GOOS=linux GOARCH=arm64 go build \
  -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
  -o port-knocker-linux-arm64 .

# Сборка для Windows AMD64
GOOS=windows GOARCH=amd64 go build \
  -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
  -o port-knocker-windows-amd64.exe .

# Сборка для macOS AMD64
GOOS=darwin GOARCH=amd64 go build \
  -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
  -o port-knocker-darwin-amd64 .

# Сборка для macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build \
  -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w" \
  -o port-knocker-darwin-arm64 .
```

### 2.2 Создание архивов

```bash
# Создать архивы для Linux
tar -czf port-knocker-linux-amd64.tar.gz port-knocker-linux-amd64
tar -czf port-knocker-linux-arm64.tar.gz port-knocker-linux-arm64

# Создать архивы для Windows
zip port-knocker-windows-amd64.exe.zip port-knocker-windows-amd64.exe

# Создать архивы для macOS
tar -czf port-knocker-darwin-amd64.tar.gz port-knocker-darwin-amd64
tar -czf port-knocker-darwin-arm64.tar.gz port-knocker-darwin-arm64

# Проверить созданные файлы
ls -la port-knocker-*
```

### 2.3 Проверка бинарников

```bash
# Проверить версию Linux бинарника
./port-knocker-linux-amd64 version 2>/dev/null || echo "Version check not available"

# Проверить версию Windows бинарника
file port-knocker-windows-amd64.exe

# Проверить размеры файлов
du -h port-knocker-*
```

## Шаг 3: Создание Git тега

### 3.1 Создание аннотированного тега

```bash
# Создать тег с сообщением
git tag -a v1.0.2 -m "Release v1.0.2

## Изменения
- Добавлена новая функциональность
- Исправлены баги
- Обновлена документация

## Установка
Скачайте соответствующий архив для вашей платформы:
- Linux AMD64: port-knocker-linux-amd64.tar.gz
- Linux ARM64: port-knocker-linux-arm64.tar.gz
- Windows AMD64: port-knocker-windows-amd64.exe.zip
- macOS AMD64: port-knocker-darwin-amd64.tar.gz
- macOS ARM64: port-knocker-darwin-arm64.tar.gz"

# Отправить тег в репозиторий
git push origin v1.0.2
```

### 3.2 Альтернативный способ с make

Если у вас настроен Makefile:

```bash
# Создать тег через make
make release-tag VERSION=v1.0.2
```

## Шаг 4: Создание релиза на GitHub

### 4.1 Через веб-интерфейс GitHub

1. Перейдите на страницу релизов: https://github.com/Direct-Dev-Ru/port-knocker/releases
2. Нажмите **"Create a new release"**
3. Выберите тег **v1.0.2**
4. Заполните информацию:
   - **Release title**: `Port Knocker v1.0.2`
   - **Description**: Скопируйте описание из тега или создайте новое
5. Загрузите все созданные архивы в раздел **"Attach binaries"**
6. Нажмите **"Publish release"**

### 4.2 Через GitHub CLI

```bash
# Установить GitHub CLI (если не установлен)
# https://cli.github.com/

# Авторизоваться
gh auth login

# Создать релиз
gh release create v1.0.2 \
  --title "Port Knocker v1.0.2" \
  --notes "## Изменения
- Добавлена новая функциональность
- Исправлены баги
- Обновлена документация

## Установка
Скачайте соответствующий архив для вашей платформы:
- Linux AMD64: port-knocker-linux-amd64.tar.gz
- Linux ARM64: port-knocker-linux-arm64.tar.gz
- Windows AMD64: port-knocker-windows-amd64.exe.zip
- macOS AMD64: port-knocker-darwin-amd64.tar.gz
- macOS ARM64: port-knocker-darwin-arm64.tar.gz" \
  --draft=false \
  --prerelease=false

# Загрузить бинарники
gh release upload v1.0.2 port-knocker-linux-amd64.tar.gz
gh release upload v1.0.2 port-knocker-linux-arm64.tar.gz
gh release upload v1.0.2 port-knocker-windows-amd64.exe.zip
gh release upload v1.0.2 port-knocker-darwin-amd64.tar.gz
gh release upload v1.0.2 port-knocker-darwin-arm64.tar.gz
```

## Шаг 5: Очистка

### 5.1 Удаление временных файлов

```bash
# Удалить бинарники
rm port-knocker-linux-amd64
rm port-knocker-linux-arm64
rm port-knocker-windows-amd64.exe
rm port-knocker-darwin-amd64
rm port-knocker-darwin-arm64

# Удалить архивы (если не нужны локально)
rm port-knocker-*.tar.gz
rm port-knocker-*.zip
```

### 5.2 Проверка релиза

```bash
# Проверить, что релиз создан
gh release view v1.0.2

# Проверить загруженные файлы
gh release view v1.0.2 --json assets
```

## Автоматизация с помощью скрипта

Создайте файл `scripts/release.sh`:

```bash
#!/bin/bash

# Скрипт для автоматического создания релиза
# Использование: ./scripts/release.sh v1.0.2

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Использование: $0 <version>"
    echo "Пример: $0 v1.0.2"
    exit 1
fi

echo "Создание релиза $VERSION..."

# Обновить версию в main.go
sed -i "s/Version.*=.*\".*\"/Version   = \"$VERSION\"/" main.go

# Обновить версию в README.md
sed -i "s/Версия.*: [0-9.]*/Версия**: $VERSION/" README.md

# Закоммитить изменения
git add .
git commit -m "Prepare for release $VERSION"
git push origin main

# Создать тег
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION

echo "Релиз $VERSION подготовлен!"
echo "Теперь создайте релиз на GitHub и загрузите бинарники."
```

Сделайте скрипт исполняемым:

```bash
chmod +x scripts/release.sh
```

## Проверка качества релиза

### Тестирование бинарников

```bash
# Протестировать Linux бинарник
./port-knocker-linux-amd64 -t "tcp:8.8.8.8:8888" -v

# Протестировать Windows бинарник (в WSL или Wine)
wine port-knocker-windows-amd64.exe -t "tcp:8.8.8.8:8888" -v
```

### Проверка совместимости

```bash
# Проверить зависимости
ldd port-knocker-linux-amd64

# Проверить архитектуру
file port-knocker-*
```

## Устранение проблем

### Проблема: Ошибка при сборке

```bash
# Очистить кэш Go
go clean -cache

# Проверить версию Go
go version

# Обновить зависимости
go mod tidy
```

### Проблема: Ошибка при загрузке на GitHub

```bash
# Проверить права доступа
gh auth status

# Проверить размер файлов (GitHub ограничивает 100MB)
ls -lh port-knocker-*
```

### Проблема: Неправильная версия в бинарнике

```bash
# Проверить переменные окружения
echo "VERSION: $VERSION"
echo "BUILD_TIME: $BUILD_TIME"

# Пересобрать с явными флагами
go build -ldflags "-X main.Version=v1.0.2 -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') -s -w" -o port-knocker .
```

## Заключение

Этот процесс позволяет создавать релизы вручную, что полезно когда:
- GitHub Actions недоступны
- Нужен полный контроль над процессом
- Требуется кастомизация сборки
- Нужно быстро исправить релиз

Для регулярных релизов рекомендуется использовать автоматизированный процесс через GitHub Actions. 