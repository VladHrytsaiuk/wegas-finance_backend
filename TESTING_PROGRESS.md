# Статус тестового покриття WeGaS Finance

## Поточні показники (Станом на 12 червня 2026)

| Пакет | Покриття | Статус |
| :--- | :--- | :--- |
| **pkg/config** | 100.0% | ✅ Повністю протестовано |
| **routes** | 100.0% | ✅ Повністю протестовано (Route Integrity) |
| **utils** | 87.7% | ✅ Високе (додано Crypto та Parsers) |
| **middlewares** | 83.8% | ✅ Високе покриття |
| **repositories** | 63.8% | 🟡 Добре (додано Resilience) |
| **pkg/telegram** | 46.7% | 🟡 Базове тестування клієнта |
| **services** | 43.7% | 🟡 Середнє (додано Security та Export) |
| **controllers** | 42.9% | 🟡 Середнє (додано WSController) |
| **services/parsers** | 32.8% | 🔴 Потребує більше прикладів виписок |

---

## План дій (Roadmap до 100% покриття)

### Етап 8: Статистика та Аналітика (Stats & Dashboard)
- [x] **StatsRepository:** Покрити тестами `GetTrend`, `GetBalances`, `GetTopFlow`.
- [x] **StatsService:** Додати тести для складних розрахунків трендів та конвертації валют.
- [x] **DashboardController:** Протестувати агреговані відповіді для головного екрану.

### Етап 9: Інтеграція з Monobank
- [x] **MonobankService:** Покрити методи `RefreshClientInfo`, `SaveSettings`, `GetUserData`.
- [x] **Webhook Handling:** Розширити тести для обробки вхідних транзакцій через вебхуки.
- [x] **Resync Logic:** Протестувати глобальну синхронізацію контрагентів.

### Етап 10: Спеціалізовані сервіси та Репозиторії
- [x] **MedicalService/Repo:** Додати покриття для медичного модуля.
- [x] **StorageService:** Тестування логіки завантаження та видалення файлів (mock local storage).
- [x] **TransactionRepository:** Додати складні тести для `BatchCreate` з великою кількістю зв'язків.

### Етап 11: Real-time та Інфраструктура
- [x] **WSController:** Тестування WebSocket з'єднань через mock-клієнтів.
- [x] **WSHub:** Тестування розсилки (broadcast) повідомлень по родинах.
- [x] **Routes:** Тестування цілісності таблиці маршрутизації (перевірка наявності всіх зареєстрованих шляхів).

### Етап 12: Стрес-тести та Edge Cases
- [x] **Resilience:** Тестування поведінки при розриві з'єднання з БД.
- [x] **Security:** Перевірка безпеки (IDOR, Role-based Access) на рівні сервісів (Account, Category, Export).
- [ ] **Limits:** Перевірка лімітів завантаження великих PDF/CSV файлів.

---

## Як запускати тести
```bash
# Запуск всіх тестів з покриттям
go test ./... -cover

# Генерація детального звіту
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```
