# Security Guidelines

## PostgreSQL Роли

### Development (текущая настройка)
```sql
Role: app_user
- Права: SELECT, INSERT, UPDATE, DELETE на таблицы
- Без SUPERUSER привилегий
- Используется для приложения
```

### Создание ролей
```bash
# Создать роль app_user (выполнить один раз):
psql -U postgres -d postgres << 'SQL'
CREATE ROLE app_user WITH LOGIN PASSWORD 'app_password';
GRANT CONNECT ON DATABASE todo TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;
SQL
```

## Known Issues & Fixes

### ⚠️ CORS открыт для всех (потенциальная уязвимость)
**Текущая настройка:**
```go
AllowOrigins: []string{"*"}  // Опасно в production!
```

**Рекомендация для production:**
```go
AllowOrigins: []string{"https://yourdomain.com"},  // Только ваш домен
```

### ✅ SQL Injection - защищён
- Используется GORM с параметризованными запросами
- Нет конкатенации строк в SQL

### ✅ Пароли - безопасны
- Хеширование через bcrypt
- Скрыты в JSON ответах (json:"-" tag)
- Не отображаются в логах

### ✅ Авторизация
- X-User-ID header проверяется на все операции с задачами
- Пользователь не может получить доступ к чужим задачам

## Production Checklist

- [ ] Изменить `AllowOrigins` с `{"*"}` на конкретный домен
- [ ] Использовать `app_user` роль вместо `postgres`
- [ ] Установить `GIN_MODE=release` в .env
- [ ] Настроить SSL для подключения к БД (`sslmode=require`)
- [ ] Использовать environment-specific .env файлы
- [ ] Добавить логирование для audit trail
- [ ] Настроить rate limiting на endpoints
- [ ] Двухфакторная аутентификация (если нужна)
