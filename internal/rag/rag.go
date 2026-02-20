package rag

import (
	"fmt"
	"strings"

	"quantumforge-rag-telegram-bot-go/internal/config"
)

const systemRU = `Ты — корпоративный ассистент QuantumForge Software.
Отвечай ТОЛЬКО на основе предоставленных фрагментов документации (context).
Если в context нет ответа или уверенность низкая — скажи строго: "Я не знаю".
Дальше предложи 1–2 уточняющих вопроса или куда смотреть (названия документов/разделов), НЕ придумывая фактов.
Всегда добавляй раздел "Источники" со списком файлов (source) и короткими цитатами (по 1 строке на источник).
Язык ответа: русский.`

type LLM interface {
	Chat(system string, user string) (string, error)
}

type Retriever interface {
	Retrieve(query string, k int) ([]ContextItem, error)
}

type ContextItem struct {
	Source string
	Text   string
	Score  float64
}

type Service struct {
	Cfg       config.Settings
	LLM       LLM
	Retriever Retriever
}

func (s *Service) Answer(question string) (string, error) {
	items, err := s.Retriever.Retrieve(question, s.Cfg.TopK)
	if err != nil {
		return "", err
	}

	ctx := formatContext(items)
	if len([]rune(ctx)) < s.Cfg.MinContextChars {
		return s.appendSources("Я не знаю Уточните вопрос (сервис/компонент/среда) или попробуйте поиск по README/Confluence.", items), nil
	}

	user := fmt.Sprintf("Вопрос: %s context:%s", question, ctx)
	ans, err := s.LLM.Chat(systemRU, user)
	if err != nil {
		return "", err
	}
	return s.appendSources(ans, items), nil
}

func formatContext(items []ContextItem) string {
	var b strings.Builder
	for i, it := range items {
		text := strings.ReplaceAll(strings.TrimSpace(it.Text), "", " ")
		r := []rune(text)
		if len(r) > 420 {
			text = string(r[:420]) + "…"
		}
		b.WriteString(fmt.Sprintf("[%d] source=%s score=%.4f | %s", i+1, it.Source, it.Score, text))
	}
	return b.String()
}

func (s *Service) appendSources(answer string, items []ContextItem) string {
	if !s.Cfg.AlwaysCiteSources {
		return answer
	}
	seen := map[string]bool{}
	lines := []string{"", "Источники:"}
	if len(items) == 0 {
		lines = append(lines, "- (нет релевантных источников)")
		return answer + "" + strings.Join(lines, "")
	}
	for _, it := range items {
		if seen[it.Source] {
			continue
		}
		seen[it.Source] = true
		sn := strings.Split(strings.TrimSpace(it.Text), "")[0]
		r := []rune(sn)
		if len(r) > 180 {
			sn = string(r[:180])
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", it.Source, sn))
	}
	return answer + "" + strings.Join(lines, "")
}
