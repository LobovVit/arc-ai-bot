package config

import (
	"os"
	"strconv"
)

type Settings struct {
	TelegramBotToken string

	LLMProvider  string
	OpenAIAPIKey string
	OpenAIModel  string

	YandexAPIKey     string
	YandexFolderID   string
	YandexBaseURL    string
	YandexModel      string
	YandexAuthScheme string

	EmbeddingsProvider string
	EmbeddingsURL      string

	QdrantURL        string
	QdrantCollection string

	DocsPath     string
	ChunkSize    int
	ChunkOverlap int
	TopK         int

	AlwaysCiteSources bool
	MinContextChars   int
	MinScore          float64
	FewShotPath       string

	RAGGuardMode string
}

func Load() Settings {
	return Settings{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),

		LLMProvider:  getenvDefault("LLM_PROVIDER", "yandexgpt"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:  getenvDefault("OPENAI_MODEL", "gpt-4o-mini"),

		YandexAPIKey:     os.Getenv("YANDEX_API_KEY"),
		YandexFolderID:   os.Getenv("YANDEX_FOLDER_ID"),
		YandexBaseURL:    getenvDefault("YANDEX_BASE_URL", "https://llm.api.cloud.yandex.net/v1"),
		YandexModel:      getenvDefault("YANDEX_MODEL", ""),
		YandexAuthScheme: getenvDefault("YANDEX_AUTH_SCHEME", "Bearer"),

		EmbeddingsProvider: getenvDefault("EMBEDDINGS_PROVIDER", "local_http"),
		EmbeddingsURL:      getenvDefault("EMBEDDINGS_URL", "http://localhost:8088"),

		QdrantURL:        getenvDefault("QDRANT_URL", "http://localhost:6333"),
		QdrantCollection: getenvDefault("QDRANT_COLLECTION", "knowledge_base"),

		DocsPath:     getenvDefault("DOCS_PATH", "./data/docs"),
		ChunkSize:    atoiDefault("CHUNK_SIZE", 1200),
		ChunkOverlap: atoiDefault("CHUNK_OVERLAP", 50),
		TopK:         atoiDefault("TOP_K", 5),

		AlwaysCiteSources: getenvBoolDefault("ALWAYS_CITE_SOURCES", true),
		MinContextChars:   atoiDefault("MIN_CONTEXT_CHARS", 200),
		MinScore:          atofDefault("MIN_SCORE", 0.45),
		FewShotPath:       getenvDefault("FEWSHOT_PATH", "./prompts/fewshot_ru.md"),

		RAGGuardMode: getenvDefault("RAG_GUARD_MODE", "on"),
	}
}

func getenvDefault(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func atoiDefault(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getenvBoolDefault(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func atofDefault(k string, def float64) float64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}
