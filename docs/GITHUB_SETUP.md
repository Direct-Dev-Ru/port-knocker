# Настройка GitHub репозитория для Port Knocker

## 1. Создание репозитория на GitHub

1. Перейдите на <https://github.com>
2. Нажмите кнопку "New repository"
3. Введите название: `port-knocker`
4. Добавьте описание: "Утилита для port knocking с поддержкой шифрования"
5. Выберите "Public" или "Private" по желанию
6. НЕ инициализируйте с README, .gitignore или лицензией (у нас уже есть)
7. Нажмите "Create repository"

## 2. Подключение локального репозитория

Выполните команды в терминале:

```bash
# Добавить remote origin
git remote add origin https://github.com/YOUR_USERNAME/port-knocker.git

# Или если используете SSH
git remote add origin git@github.com:YOUR_USERNAME/port-knocker.git

# Отправить код в репозиторий
git branch -M main
git push -u origin main
```

## 3. Настройка автоматических релизов

После первого push GitHub Actions автоматически будет:

### При каждом push

- Запускать тесты (`.github/workflows/test.yml`)
- Проверять сборку для разных платформ
- Выполнять линтинг кода

### При создании тега

- Собирать бинарные файлы для всех платформ
- Создавать release с архивами
- Публиковать artifacts

## 4. Создание первого релиза

```bash
# Создать и отправить тег
make release-tag VERSION=v1.0.0

# Или вручную
git tag -a v1.0.0 -m "Initial release v1.0.0"
git push origin v1.0.0
```

После этого GitHub Actions автоматически:

1. Соберет бинарники для всех платформ
2. Создаст release на GitHub
3. Прикрепит архивы с бинарниками

## 5. Что исключено из репозитория

Благодаря `.gitignore` НЕ попадают в репозиторий:

### Персональные конфигурации

- `*knock*.yaml` - все файлы с "knock" в имени
- `*knock*.yml`
- `*.encrypted` - зашифрованные конфиги
- `key.txt`, `*.key` - файлы ключей
- `*secret*` - любые файлы с "secret"

### Сборочные артефакты

- `port-knocker` - основной бинарник
- `port-knocker-*` - платформо-специфичные бинарники
- `*.exe` - Windows исполняемые файлы

### Служебные файлы

- IDE настройки (`.vscode/`, `.idea/`)
- Системные файлы (`.DS_Store`, `Thumbs.db`)
- Логи и временные файлы

## 6. Рабочий процесс разработки

### Обычная разработка

```bash
# Проверить статус и что исключено
make check-git

# Добавить изменения
git add .
git commit -m "Описание изменений"
git push
```

### Создание релиза

```bash
# Обновить версию в коде
# Создать тег и release
make release-tag VERSION=v1.0.1
```

### Тестирование локально

```bash
# Сборка и тесты
make build
make test

# Тестирование с примером конфигурации
make example-run
```

## 7. Структура релизов

Каждый релиз будет содержать:

- `port-knocker-linux-amd64.tar.gz` - Linux x64
- `port-knocker-linux-arm64.tar.gz` - Linux ARM64  
- `port-knocker-windows-amd64.exe.zip` - Windows x64
- `port-knocker-darwin-amd64.tar.gz` - macOS Intel
- `port-knocker-darwin-arm64.tar.gz` - macOS Apple Silicon

Каждый архив включает:

- Бинарный файл
- README.md
- Пример конфигурации

## 8. Безопасность

✅ **Защищено от случайной публикации:**

- Персональные конфигурации с реальными хостами
- Ключи шифрования
- Зашифрованные файлы
- Рабочие конфигурации

✅ **Публикуется безопасно:**

- Исходный код
- Примеры конфигураций
- Документация
- Бинарные файлы через GitHub Releases
