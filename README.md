# SwitchCraft 🚀

**Zero-Downtime Blue-Green Deployment Manager**

SwitchCraft - это простой и эффективный инструмент для управления blue-green деплоем с веб-интерфейсом. Легко интегрируется в любую инфраструктуру без зависимостей.

![Blue-Green Deployment](https://img.shields.io/badge/Deployment-Blue__Green-green)
![Go](https://img.shields.io/badge/Go-1.19+-blue)
![Docker](https://img.shields.io/badge/Docker-Ready-blue)
![Zero-Downtime](https://img.shields.io/badge/Zero--Downtime-✓-success)

## ✨ Особенности

- 🎯 **Zero-Downtime Deployments** - переключение без прерывания сервиса
- 🖥️ **Веб-интерфейс** - интуитивное управление через браузер
- 🔄 **Автоматические health-checks** - мониторинг состояния сред
- 🐳 **Docker-совместимость** - готов к работе с контейнерами
- ⚡ **Простота** - один бинарный файл, без зависимостей
- 🔒 **Безопасность** - проверка здоровья перед переключением
- 📊 **Мониторинг** - отслеживание версий и статусов

## 🏗️ Архитектура

```
Пользователи → SwitchCraft (:8080) → Активная среда (Blue/Green)
                      │
                      ├── Веб-интерфейс управления
                      ├── API для автоматизации
                      └── Health-check мониторинг
```

## 🚀 Быстрый старт

### 1. Клонирование и сборка

```bash
git clone https://gitlab.com/your-project/switchcraft.git
cd switchcraft
docker-compose up -d
```

### 2. Открытие веб-интерфейса

Перейдите по адресу: `http://localhost:8080`

### 3. Настройка сред

В веб-интерфейсе укажите порты ваших сред:
- **Blue Environment**: 5176
- **Green Environment**: 5177

## 📋 Конфигурация

### Переменные окружения

```bash
BLUE_PORT=5176          # Порт blue среды
GREEN_PORT=5177         # Порт green среды  
PROXY_PORT=8080         # Порт SwitchCraft
SERVICE_NAME="My App"   # Название сервиса
```

### Docker Compose

```yaml
services:
  switchcraft:
    image: your-registry/switchcraft:latest
    ports:
      - "8080:8080"
    environment:
      - BLUE_PORT=5176
      - GREEN_PORT=5177
      - SERVICE_NAME="Production Frontend"
```

## 🎯 Использование

### Веб-интерфейс

1. **Откройте панель управления** - `http://localhost:8080`
2. **Проверьте статус** - убедитесь, что обе среды здоровы
3. **Выполните деплой** в неактивную среду
4. **Переключите трафик** одной кнопкой

### API Endpoints

```bash
# Получить статус
GET /api/status

# Переключить среду
POST /api/switch

# Обновить конфигурацию  
POST /api/config

# Запустить деплой
POST /api/deploy
```

### Пример API вызова

```bash
# Переключение на green среду
curl -X POST http://localhost:8080/api/switch

# Проверка статуса
curl http://localhost:8080/api/status
```

## 🔧 Интеграция с CI/CD

### GitLab CI Example

```yaml
deploy:blue:
  script:
    - docker-compose up -d app-blue
    - sleep 10
    - curl -f http://app-blue:5176/health || exit 1

switch-to-blue:
  script:
    - curl -X POST http://switchcraft:8080/api/switch
```

## 🏷️ Health Checks

Ваше приложение должно предоставлять эндпоинты:

- `GET /health` - возвращает `200 OK` если сервис здоров
- `GET /version` - возвращает версию приложения (опционально)

## 📦 Развертывание

### 1. Локальная разработка

```bash
go run main.go
```

### 2. Production с Docker

```bash
docker build -t switchcraft .
docker run -p 8080:8080 -e BLUE_PORT=5176 switchcraft
```

### 3. Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: switchcraft
spec:
  template:
    spec:
      containers:
      - name: switchcraft
        image: switchcraft:latest
        ports:
        - containerPort: 8080
        env:
        - name: BLUE_PORT
          value: "5176"
```

## 🛠️ Разработка

### Требования

- Go 1.19+
- Docker (опционально)

### Локальная разработка

```bash
# Клонирование
git clone https://gitlab.com/your-project/switchcraft.git

# Запуск
go mod tidy
go run main.go
```

### Сборка

```bash
go build -o switchcraft main.go
```

## 🤝 Вклад в проект

Мы приветствуем contributions! 

1. Форкните репозиторий
2. Создайте feature branch
3. Commit ваши изменения
4. Push в branch
5. Создайте Merge Request

## 📄 Лицензия

MIT License - смотрите файл [LICENSE](LICENSE) для деталей.



**SwitchCraft** - сделайте ваши деплои безболезненными! ✨
