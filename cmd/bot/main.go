package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"quantumforge-rag-telegram-bot-go/internal/config"
	"quantumforge-rag-telegram-bot-go/internal/embeddings"
	"quantumforge-rag-telegram-bot-go/internal/llm"
	"quantumforge-rag-telegram-bot-go/internal/qdrant"
	"quantumforge-rag-telegram-bot-go/internal/rag"
)

type Retriever struct {
	Cfg config.Settings
	Q   *qdrant.Client
	E   *embeddings.LocalHTTP
}

func (r *Retriever) Retrieve(query string, k int) ([]rag.ContextItem, error) {
	vec, err := r.E.Embed(query)
	if err != nil {
		return nil, err
	}
	payloads, err := r.Q.Search(r.Cfg.QdrantCollection, vec, k)
	if err != nil {
		return nil, err
	}
	items := make([]rag.ContextItem, 0, len(payloads))
	for _, p := range payloads {
		src, _ := p["source"].(string)
		txt, _ := p["text"].(string)
		score, _ := p["_score"].(float64)
		items = append(items, rag.ContextItem{Source: src, Text: txt, Score: score})
	}
	return items, nil
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	emb := embeddings.NewLocalHTTP(cfg.EmbeddingsURL)
	q := qdrant.New(cfg.QdrantURL)

	ragSvc := &rag.Service{
		Cfg:       cfg,
		LLM:       selectLLM(cfg),
		Retriever: &Retriever{Cfg: cfg, Q: q, E: emb},
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized as @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for upd := range updates {
		if upd.Message == nil {
			continue
		}
		if upd.Message.IsCommand() {
			switch upd.Message.Command() {
			case "start":
				bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID,
					"Привет! Я бот базы знаний QuantumForge. Спрашивай — я отвечу по документации и скажу «Я не знаю», если ответа нет."))
			case "help":
				bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID,
					"Команды: /start, /help. Просто напиши вопрос текстом."))
			default:
				bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID, "Неизвестная команда."))
			}
			continue
		}
		ans, err := ragSvc.Answer(upd.Message.Text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID, "Ошибка: "+err.Error()))
			continue
		}
		bot.Send(tgbotapi.NewMessage(upd.Message.Chat.ID, ans))
	}
}

func selectLLM(cfg config.Settings) rag.LLM {
	// Provider selection with safe fallback to Stub.
	// Default provider is yandexgpt (see config), but you can switch via LLM_PROVIDER.
	switch cfg.LLMProvider {
	case "yandexgpt", "yandex", "yandex_gpt":
		if cfg.YandexAPIKey != "" && cfg.YandexFolderID != "" {
			return llm.NewYandexGPT(cfg.YandexAPIKey, cfg.YandexFolderID, cfg.YandexBaseURL, cfg.YandexModel, cfg.YandexAuthScheme)
		}
		log.Printf("LLM_PROVIDER=%s but YANDEX_API_KEY or YANDEX_FOLDER_ID is empty — fallback to stub", cfg.LLMProvider)
	case "openai":
		if cfg.OpenAIAPIKey != "" {
			return llm.NewOpenAI(cfg.OpenAIAPIKey, cfg.OpenAIModel)
		}
		log.Printf("LLM_PROVIDER=openai but OPENAI_API_KEY is empty — fallback to stub")
	default:
		log.Printf("LLM_PROVIDER=%s is not supported — fallback to stub", cfg.LLMProvider)
	}
	return &llm.Stub{}
}
