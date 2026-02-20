# QuantumForge RAG Telegram Bot — Go + Local Embeddings + Qdrant

## Что поднимается
- Qdrant (vector DB)
- embeddings_service (FastAPI + Sentence-Transformers)
- Go: ingest (индексация), query (проверка), bot (Telegram)

## Быстрый старт
```bash
cp .env .env
mkdir -p data/docs
# положите сюда документы из Task2 (knowledge_base/*.md)

docker compose up -d qdrant embeddings
docker compose run --rm ingest
docker compose run --rm query "Что такое Synth Flux?"
# затем:
# docker compose up -d bot
```

## Эмбеддинги
По умолчанию:
- sentence-transformers/all-MiniLM-L6-v2 (dim=384)
