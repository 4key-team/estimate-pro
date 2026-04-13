# Настройка Telegram-бота Эсти

Пошаговая инструкция по настройке бота Эсти для EstimatePro.

## Предварительные требования

- Запущенная инфраструктура: PostgreSQL, Redis, MinIO (`just dev-infra`)
- Актуальный бэкенд из ветки `main` (v0.9.0+)
- Аккаунт в Telegram
- API-ключ от LLM провайдера (Claude / OpenAI / Grok) или локальный Ollama

---

## Шаг 1. Создать бота в Telegram

1. Открой Telegram и найди **@BotFather**
2. Отправь команду `/newbot`
3. Введи отображаемое имя бота: `Эсти | EstimatePro`
4. Введи username бота: `esti_pro_bot` (или любой свободный, запомни его)
5. BotFather ответит токеном вида `7123456789:AAF...xyz` — **скопируй его**

> Рекомендуется также:
> - `/setdescription` — "AI-ассистент для управления проектами EstimatePro"
> - `/setabouttext` — "Привет! Я Эсти, помогаю с проектами, оценками и командой"
> - `/setuserpic` — загрузи аватар бота

---

## Шаг 2. Получить LLM API ключ

Бот использует LLM для парсинга сообщений. Выбери провайдера:

| Провайдер | Где получить ключ | Модель по умолчанию |
|-----------|-------------------|---------------------|
| **Claude** (рекомендуется) | https://console.anthropic.com/settings/keys | `claude-sonnet-4-20250514` |
| **OpenAI** | https://platform.openai.com/api-keys | `gpt-4o` |
| **Grok** | https://console.x.ai/ | `grok-3-mini` |
| **Ollama** (бесплатно, локально) | https://ollama.ai — установи и запусти `ollama pull llama3.1` | `llama3.1` |

---

## Шаг 3. Настроить переменные окружения

Добавь в файл `.env` в корне проекта:

```bash
# Telegram Bot
TELEGRAM_BOT_TOKEN=7123456789:AAFxxx...          # токен от BotFather
TELEGRAM_WEBHOOK_SECRET=my-random-secret-string   # любая строка для верификации webhook
TELEGRAM_BOT_USERNAME=esti_pro_bot                # username бота БЕЗ @

# LLM Provider
LLM_PROVIDER=claude                               # claude / openai / grok / ollama
LLM_API_KEY=sk-ant-api03-xxx...                   # API ключ (не нужен для ollama)
LLM_MODEL=claude-sonnet-4-20250514                  # модель (см. таблицу выше)
LLM_BASE_URL=                                     # пусто для cloud-провайдеров, http://localhost:11434 для ollama
```

> `TELEGRAM_WEBHOOK_SECRET` — придумай сам, любая строка. Telegram будет отправлять её в заголовке для верификации.

---

## Шаг 4. Запустить миграции

```bash
just migrate-up
```

Это создаст таблицы:
- `bot_sessions` — состояние диалогов
- `bot_user_links` — связка Telegram ↔ EP аккаунт
- `llm_configs` — настройки LLM провайдера
- `bot_memory` — история разговоров
- `bot_user_prefs` — предпочтения пользователя

---

## Шаг 5. Запустить бэкенд

```bash
just dev-backend
```

Убедись что сервер стартовал:
```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok"}
```

---

## Шаг 6. Настроить публичный URL (webhook)

Telegram отправляет сообщения на webhook — нужен публичный HTTPS URL.

### Вариант А: ngrok (для разработки)

```bash
ngrok http 8080
```

Скопируй HTTPS URL (например `https://abc123.ngrok-free.app`).

### Вариант Б: Production

Используй домен с SSL-сертификатом, например `https://api.estimatepro.com`.

---

## Шаг 7. Зарегистрировать webhook в Telegram

Замени `<TOKEN>` на токен бота, `<URL>` на публичный URL, `<SECRET>` на значение `TELEGRAM_WEBHOOK_SECRET`:

```bash
curl -X POST "https://api.telegram.org/bot<TOKEN>/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "<URL>/api/v1/bot/webhook",
    "secret_token": "<SECRET>"
  }'
```

Ожидаемый ответ:
```json
{"ok": true, "result": true, "description": "Webhook was set"}
```

### Проверить webhook:
```bash
curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo"
```

---

## Шаг 8. Привязать Telegram аккаунт

Чтобы бот узнавал пользователей, каждый должен ввести свой **Telegram Chat ID** в настройках EstimatePro:

1. Открой Telegram, найди бота **@userinfobot**
2. Напиши ему любое сообщение
3. Он ответит твоим числовым ID (например `123456789`)
4. Зайди в EstimatePro → **Настройки** → **Уведомления**
5. Включи тумблер **Telegram**
6. Введи свой Chat ID в поле и нажми **Save**

> Это нужно сделать каждому пользователю, который хочет общаться с Эсти.

---

## Шаг 9. Проверить работу

Открой Telegram и напиши боту:

| Сообщение | Ожидаемый ответ |
|-----------|-----------------|
| `Привет` | Приветствие от Эсти |
| `Покажи мои проекты` | Список проектов пользователя |
| `Помощь` | Справка по командам |
| `Создай проект Тестовый` | Подтверждение с inline-кнопками |
| `Покажи сводку по проекту` | PERT таблица с оценками |

### В групповом чате:

1. Добавь бота в группу
2. Обращайся по имени: `Эсти, покажи проекты` или `@esti_pro_bot список участников`
3. Бот также реагирует на: Эстя, Эстик, Эстюша, Esti

---

## Устранение проблем

### Бот не отвечает

1. Проверь webhook: `curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo`
   - `pending_update_count` > 0 — webhook зарегистрирован, но бэкенд не обрабатывает
   - `last_error_message` — покажет ошибку
2. Проверь логи бэкенда — ищи `BotUsecase.ProcessMessage`
3. Проверь что `TELEGRAM_BOT_TOKEN` и `TELEGRAM_WEBHOOK_SECRET` совпадают

### "Привяжите аккаунт"

Пользователь не ввёл свой Telegram Chat ID в настройках EstimatePro. См. Шаг 8.

### "Ошибка конфигурации LLM"

Проверь `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_MODEL` в `.env`. Для Ollama убедись что он запущен (`ollama serve`).

### Бот не реагирует в группе

- Убедись что бот добавлен в группу как участник
- Обращайся по имени: `Эсти, ...` или `@username ...`
- Или ответь на сообщение бота (reply)

### Webhook на localhost

Telegram не может отправить запрос на `localhost`. Используй ngrok или аналог для проброса.

---

## Переменные окружения (полный список)

| Переменная | Обязательна | Описание | Пример |
|------------|-------------|----------|--------|
| `TELEGRAM_BOT_TOKEN` | Да | Токен бота от BotFather | `7123456789:AAF...` |
| `TELEGRAM_WEBHOOK_SECRET` | Да | Секрет для верификации webhook | `my-secret-123` |
| `TELEGRAM_BOT_USERNAME` | Да | Username бота без @ | `esti_pro_bot` |
| `LLM_PROVIDER` | Нет | Провайдер LLM (по умолчанию `claude`) | `claude` / `openai` / `grok` / `ollama` |
| `LLM_API_KEY` | Да* | API ключ (*не нужен для ollama) | `sk-ant-...` |
| `LLM_MODEL` | Нет | Модель (есть дефолты) | `claude-sonnet-4-20250514` |
| `LLM_BASE_URL` | Нет | Base URL для Ollama | `http://localhost:11434` |
