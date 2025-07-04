#!/bin/bash

# Скрипт для быстрого создания релиза Port Knocker
# Использование: ./docs/scripts/quick-release.sh v1.0.4

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка аргументов
VERSION=$1
if [ -z "$VERSION" ]; then
    log_error "Не указана версия!"
    echo "Использование: $0 <version>"
    echo "Пример: $0 v1.0.4"
    exit 1
fi

# Проверка формата версии
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "Неверный формат версии: $VERSION"
    echo "Используйте формат: vX.Y.Z (например: v1.0.4)"
    exit 1
fi

log_info "Начинаем создание релиза $VERSION..."

# Проверка зависимостей
log_info "Проверяем зависимости..."
command -v go >/dev/null 2>&1 || { log_error "Go не установлен"; exit 1; }
command -v git >/dev/null 2>&1 || { log_error "Git не установлен"; exit 1; }
command -v tar >/dev/null 2>&1 || { log_error "tar не установлен"; exit 1; }
command -v zip >/dev/null 2>&1 || { log_error "zip не установлен"; exit 1; }

# Проверка статуса git
log_info "Проверяем статус Git..."
if [ -n "$(git status --porcelain)" ]; then
    log_warning "Есть незакоммиченные изменения!"
    echo "Текущие изменения:"
    git status --short
    read -p "Продолжить? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Отменено пользователем"
        exit 1
    fi
fi

# Обновление версии
log_info "Обновляем версию в main.go..."
sed -i "s/Version.*=.*\".*\"/Version   = \"$VERSION\"/" main.go

log_info "Обновляем версию в README.md..."
sed -i "s/Версия.*: [0-9.]*/Версия**: ${VERSION#v}/" README.md

# Проверка изменений
log_info "Проверяем изменения..."
if [ -z "$(git diff)" ]; then
    log_warning "Нет изменений для коммита"
else
    echo "Изменения:"
    git diff --stat
fi

# Коммит изменений
log_info "Коммитим изменения..."
git add .
git commit -m "Prepare for release $VERSION"
git push origin main

# Сборка бинарников
log_info "Собираем бинарники для всех платформ..."
export VERSION_NUM="${VERSION#v}"
export BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Функция сборки для платформы
build_for_platform() {
    local os=$1
    local arch=$2
    local suffix=$3
    local binary_name="port-knocker-$suffix"
    
    log_info "Собираем для $os/$arch..."
    GOOS=$os GOARCH=$arch go build \
        -ldflags "-X main.Version=${VERSION_NUM} -X main.BuildTime=${BUILD_TIME} -s -w" \
        -o "$binary_name" .
    
    # Создание архива
    if [[ "$os" == "windows" ]]; then
        zip "${binary_name}.zip" "$binary_name"
    else
        tar -czf "${binary_name}.tar.gz" "$binary_name"
    fi
    
    # Удаление бинарника
    rm "$binary_name"
    
    log_success "Создан архив для $os/$arch"
}

# Сборка для всех платформ
build_for_platform "linux" "amd64" "linux-amd64"
build_for_platform "linux" "arm64" "linux-arm64"
build_for_platform "windows" "amd64" "windows-amd64.exe"
build_for_platform "darwin" "amd64" "darwin-amd64"
build_for_platform "darwin" "arm64" "darwin-arm64"

# Проверка созданных файлов
log_info "Проверяем созданные архивы..."
ls -la port-knocker-*

# Создание Git тега
log_info "Создаем Git тег..."
git tag -a "$VERSION" -m "Release $VERSION

## Изменения
- Обновления и исправления
- Улучшения производительности
- Обновлена документация

## Установка
Скачайте соответствующий архив для вашей платформы:
- Linux AMD64: port-knocker-linux-amd64.tar.gz
- Linux ARM64: port-knocker-linux-arm64.tar.gz
- Windows AMD64: port-knocker-windows-amd64.exe.zip
- macOS AMD64: port-knocker-darwin-amd64.tar.gz
- macOS ARM64: port-knocker-darwin-arm64.tar.gz"

git push origin "$VERSION"

# Проверка GitHub CLI
if command -v gh >/dev/null 2>&1; then
    log_info "GitHub CLI найден. Создаем релиз..."
    
    # Проверка авторизации
    if gh auth status >/dev/null 2>&1; then
        log_info "Создаем релиз на GitHub..."
        gh release create "$VERSION" \
            --title "Port Knocker $VERSION" \
            --notes "## Port Knocker $VERSION

### Изменения
- Обновления и исправления
- Улучшения производительности
- Обновлена документация

### Установка
Скачайте соответствующий архив для вашей платформы:
- **Linux AMD64**: \`port-knocker-linux-amd64.tar.gz\`
- **Linux ARM64**: \`port-knocker-linux-arm64.tar.gz\`
- **Windows AMD64**: \`port-knocker-windows-amd64.exe.zip\`
- **macOS AMD64**: \`port-knocker-darwin-amd64.tar.gz\`
- **macOS ARM64**: \`port-knocker-darwin-arm64.tar.gz\`

### Использование
\`\`\`bash
# Инлайн цели
./port-knocker -t \"tcp:host:port;udp:host:port\" -v

# Конфигурационный файл
./port-knocker -c config.yaml -v

# Пасхалка
./port-knocker -t \"tcp:8.8.8.8:8888\" -v
\`\`\`" \
            --draft=false \
            --prerelease=false
        
        log_info "Загружаем бинарники..."
        gh release upload "$VERSION" port-knocker-*.tar.gz port-knocker-*.zip
        
        log_success "Релиз $VERSION создан и опубликован на GitHub!"
    else
        log_warning "GitHub CLI не авторизован. Создайте релиз вручную."
        log_info "Перейдите на: https://github.com/Direct-Dev-Ru/port-knocker/releases"
        log_info "Загрузите файлы: port-knocker-*.tar.gz port-knocker-*.zip"
    fi
else
    log_warning "GitHub CLI не установлен. Создайте релиз вручную."
    log_info "Перейдите на: https://github.com/Direct-Dev-Ru/port-knocker/releases"
    log_info "Загрузите файлы: port-knocker-*.tar.gz port-knocker-*.zip"
fi

# Очистка
log_info "Очищаем временные файлы..."
rm -f port-knocker-*.tar.gz port-knocker-*.zip

log_success "Релиз $VERSION успешно создан!"
log_info "Тег: $VERSION"
log_info "Релиз: https://github.com/Direct-Dev-Ru/port-knocker/releases/tag/$VERSION"

echo
log_info "Следующие шаги:"
echo "1. Проверьте релиз на GitHub"
echo "2. Протестируйте скачанные бинарники"
echo "3. Обновите документацию если нужно" 