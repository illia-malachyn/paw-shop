# PawShop

Навчальний інтернет-магазин корму для собак на Go. Проєкт створено для демонстрації патернів проєктування (GoF) в контексті реальної предметної області.

## Технічне завдання

**Предметна область:** інтернет-магазин корму для собак з каталогом товарів, кошиком, замовленнями, знижками, пошуком, чатом підтримки та системою оповіщень.

**Мова:** Go 1.23, без зовнішніх залежностей (тільки stdlib).

**Архітектура:** HTTP-сервер на `:8080` зі стандартним `net/http`. Бізнес-логіка розділена на пакети в `internal/`, кожен з яких відповідає за окрему підсистему.

## Структура проєкту

```
paw-shop/
├── cmd/server/main.go           — точка входу, HTTP-сервер, реєстрація роутів
├── static/index.html            — лендінг
└── internal/
    ├── models/                  — доменні моделі (Product, DryFood, WetFood, Treat)
    ├── factory/                 — створення товарів (Factory Method, Abstract Factory)
    ├── bundle/                  — набори корму (Prototype, Builder)
    ├── discount/                — система знижок (Strategy, Command)
    ├── notification/            — підписка на зміну ціни (Observer)
    ├── order/                   — замовлення, стани, звіти (MacroCommand, Template Method, State, Iterator, Chain of Responsibility)
    ├── search/                  — пошукові запити (Interpreter)
    ├── chat/                    — чат підтримки (Mediator)
    ├── cart/                    — кошик з undo (Memento)
    ├── export/                  — експорт замовлення (Visitor)
    ├── notify/                  — оповіщення (Facade)
    ├── logging/                 — логування (Proxy, Bridge)
    └── handler/                 — HTTP-хендлери для всіх ендпоінтів
```

## Лабораторні роботи та патерни

### ЛР 1 — Каталог товарів

**Патерни:** Factory Method, Abstract Factory

**Factory Method** — створення товарів різних типів (сухий корм, вологий корм, ласощі) через фабрики `DryFoodFactory`, `WetFoodFactory`, `TreatFactory`. Кожна фабрика реалізує інтерфейс `ProductFactory` і інкапсулює логіку створення конкретного типу продукту.

**Abstract Factory** — створення сімейств товарів за брендом. `BrandFactory` визначає методи для створення всіх типів корму одного бренду (`RoyalCaninFactory`, `AcanaFactory`). Це гарантує, що характеристики продуктів (якість, цінова категорія) відповідають бренду.

**Пакети:** `internal/models`, `internal/factory`

---

### ЛР 2 — Набори корму

**Патерни:** Prototype, Builder

**Prototype** — клонування шаблонних наборів корму. `Bundle` має метод `Clone()`, а `BundleRegistry` зберігає готові шаблони-прототипи. Користувач обирає шаблон, клонує його та модифікує під себе без впливу на оригінал.

**Builder** — покрокове конструювання кастомного набору через `BundleBuilder` з fluent API. Дозволяє додавати товари, встановлювати назву, застосовувати знижку — і отримати готовий `Bundle` викликом `Build()`.

**Пакет:** `internal/bundle`

---

### ЛР 3 — Знижки та сповіщення

**Патерни:** Strategy, Observer, Command

**Strategy** — різні алгоритми розрахунку знижки: `PercentStrategy` (відсоткова), `FixedStrategy` (фіксована сума), `BuyNGetOneStrategy` (N+1). Алгоритм обирається під час виконання без зміни клієнтського коду.

**Observer** — підписка на зміну ціни товару. `PriceSubject` повідомляє всіх підписників (`PriceObserver`, `LogObserver`, `InMemoryObserver`) при зміні ціни.

**Command** — застосування знижки як команда (`ApplyDiscountCommand`) з можливістю відкату. `CommandHistory` зберігає історію виконаних команд для undo.

**Пакети:** `internal/discount`, `internal/notification`

---

### ЛР 4 — Пакетні дії з замовленнями

**Патерни:** MacroCommand, Template Method

**MacroCommand** — об'єднання кількох команд замовлення в одну. `MacroCommand` приймає список `OrderCommand` (`ConfirmOrderCommand`, `RejectOrderCommand`) і виконує їх послідовно. Дозволяє підтвердити або відхилити кілька замовлень однією дією.

**Template Method** — генерація звітів за шаблоном. Функція `GenerateReport` визначає алгоритм (Header → Body → Footer), а конкретні генератори (`DailyReportGenerator`, `WeeklyReportGenerator`) перевизначають окремі кроки.

**Пакет:** `internal/order`

---

### ЛР 5 — Статуси замовлення та валідація

**Патерни:** Iterator, State, Chain of Responsibility

**Iterator** — обхід колекції замовлень з можливістю фільтрації за статусом. `OrderCollection` створює `OrderIterator` або `FilteredIterator`, що повертає лише замовлення з потрібним статусом.

**State** — життєвий цикл замовлення як набір станів (`NewState` → `ConfirmedState` → `ShippedState` → `DeliveredState`). Кожен стан контролює дозволені переходи та поведінку методів `Next()` і `Cancel()`.

**Chain of Responsibility** — ланцюжок валідаторів при створенні замовлення: `StockValidator` → `AddressValidator` → `PaymentValidator`. Кожен валідатор або пропускає запит далі, або повертає помилку.

**Пакет:** `internal/order`

---

### ЛР 6 — Фільтрація товарів та чат підтримки

**Патерни:** Interpreter, Mediator

**Interpreter** — парсинг пошукових запитів у дерево виразів. Підтримує синтаксис `brand:Royal`, `price:<500`, `category:dry` та комбінацію через `AND`. Кожен вираз (`BrandExpression`, `PriceLessThanExpression`, `CategoryExpression`, `AndExpression`) реалізує інтерфейс `Expression` з методом `Interpret()`.

**Mediator** — координація між учасниками чату підтримки. `SupportChatMediator` маршрутизує повідомлення між `Customer` та `Manager`. Учасники не знають один про одного — спілкуються тільки через медіатор.

**Пакети:** `internal/search`, `internal/chat`

---

### ЛР 7 — Кошик з undo та експорт

**Патерни:** Memento, Visitor

**Memento** — збереження і відновлення стану кошика. Перед кожною дією (додавання, видалення товару) `Cart` зберігає snapshot у `CartMemento`. `CartHistory` тримає стек мементо для undo.

**Visitor** — обхід елементів кошика для генерації різних форматів. `JSONExportVisitor` генерує JSON-представлення, `TextReceiptVisitor` — текстовий чек. Обидва реалізують інтерфейс `OrderVisitor` з методами `VisitItem()` та `VisitCart()`.

**Пакети:** `internal/cart`, `internal/export`

---

### ЛР 8 — Оповіщення та логування

**Патерни:** Facade, Proxy, Bridge

**Facade** — спрощений інтерфейс для відправки оповіщень. `NotificationFacade` приховує роботу з кількома каналами (`ConsoleNotifier`, `FileNotifier`) за методами `NotifyUser()` та `NotifyOrderStatusChanged()`.

**Proxy** — проксі для логера з додатковою поведінкою. `LoggerProxy` виконує lazy-ініціалізацію `FileLogger` при першому виклику, рахує кількість записів та фільтрує за рівнем (info, warn, error).

**Bridge** — розділення форматування від виводу. Абстракція (`TextFormatter`, `JSONFormatter`) визначає як форматувати дані, а реалізація (`ConsoleWriter`, `FileWriter`) — куди виводити. Будь-який форматер комбінується з будь-яким writer.

**Пакети:** `internal/notify`, `internal/logging`

## API

| Метод | Ендпоінт | Опис |
|-------|----------|------|
| GET | `/api/products` | Каталог товарів |
| GET | `/api/products/search?q=...` | Пошук з фільтрацією |
| GET | `/api/bundles/templates` | Шаблони наборів |
| POST | `/api/bundles` | Створити набір (Builder) |
| POST | `/api/bundles/clone` | Клонувати шаблон (Prototype) |
| POST | `/api/discounts/apply` | Застосувати знижку |
| POST | `/api/discounts/undo` | Відмінити знижку |
| POST | `/api/products/{id}/subscribe` | Підписка на зміну ціни |
| POST | `/api/orders` | Створити замовлення (валідація) |
| GET | `/api/orders?status=...` | Список замовлень (Iterator) |
| POST | `/api/orders/batch` | Пакетна дія (MacroCommand) |
| PATCH | `/api/orders/{id}/status` | Зміна статусу (State) |
| GET | `/api/reports/{type}` | Звіт (daily/weekly) |
| POST | `/api/chat/send` | Надіслати повідомлення |
| GET | `/api/chat/history?participant=...` | Історія чату |
| POST | `/api/cart/add` | Додати в кошик |
| POST | `/api/cart/remove` | Видалити з кошика |
| POST | `/api/cart/undo` | Undo останньої дії |
| GET | `/api/cart` | Стан кошика |
| GET | `/api/cart/export?format=...` | Експорт (json/text) |
| POST | `/api/notifications/send` | Надіслати оповіщення |
| GET | `/api/logs?level=...` | Записи логу |
| GET | `/api/logs/stats` | Статистика логів |

## Зведена таблиця патернів

| # | Патерн | Тип | Пакет | ЛР |
|---|--------|-----|-------|----|
| 1 | Factory Method | Породжуючий | `internal/factory` | 1 |
| 2 | Abstract Factory | Породжуючий | `internal/factory` | 1 |
| 3 | Prototype | Породжуючий | `internal/bundle` | 2 |
| 4 | Builder | Породжуючий | `internal/bundle` | 2 |
| 5 | Strategy | Поведінковий | `internal/discount` | 3 |
| 6 | Observer | Поведінковий | `internal/notification` | 3 |
| 7 | Command | Поведінковий | `internal/discount` | 3 |
| 8 | MacroCommand | Поведінковий | `internal/order` | 4 |
| 9 | Template Method | Поведінковий | `internal/order` | 4 |
| 10 | Iterator | Поведінковий | `internal/order` | 5 |
| 11 | State | Поведінковий | `internal/order` | 5 |
| 12 | Chain of Responsibility | Поведінковий | `internal/order` | 5 |
| 13 | Interpreter | Поведінковий | `internal/search` | 6 |
| 14 | Mediator | Поведінковий | `internal/chat` | 6 |
| 15 | Memento | Поведінковий | `internal/cart` | 7 |
| 16 | Visitor | Поведінковий | `internal/export` | 7 |
| 17 | Facade | Структурний | `internal/notify` | 8 |
| 18 | Proxy | Структурний | `internal/logging` | 8 |
| 19 | Bridge | Структурний | `internal/logging` | 8 |

## Запуск

```bash
go run cmd/server/main.go
```

Сервер запуститься на `http://localhost:8080`.

## Тести

```bash
go test ./...
```
