# Task 3 — Создание векторного индекса базы знаний

## Модель эмбеддингов
- Название: `sentence-transformers/all-MiniLM-L6-v2`
- Тип: локальная модель Sentence-Transformers
- Размерность эмбеддингов: **384**

## База знаний
- Источник: Task 2 (переименованная вселенная Asterion Saga)
- Формат документов: `.md`
- Количество документов: 38

## Чанкинг
- Параметры: `CHUNK_SIZE=1200`, `CHUNK_OVERLAP=50`
- Количество чанков в индексе: **37**

## Векторная БД
- Выбрана: **Qdrant**
- Причины выбора: прод-готовый сервис, HTTP/gRPC API, удобная интеграция с Go, поддержка метаданных и фильтрации

## Время индексации
- Время построения индекса: **~0.76 сек**

## Пример запроса к индексу

Команда:
```bash
docker compose run --rm query "Что такое Synth Flux?"
```
1) score=0.5939 source=knowledge_base/synth_flux.md chunk_id=0
   Synth Flux — феномен во вселенной Asterion Saga...

Вывод

Векторный индекс успешно создан, поиск по базе знаний работает корректно: по запросу о сущности Synth Flux первым возвращается соответствующий документ с наибольшей релевантностью.


## повтор:
```bash
docker compose up -d qdrant embeddings
curl -s http://localhost:8088/health
docker compose run --rm ingest
docker compose run --rm query "Что такое Synth Flux?"
```

