package rag

import (
	"fmt"
	"os"
	"strings"

	"quantumforge-rag-telegram-bot-go/internal/config"
)

// We follow the assignment requirements:
// - RAG: answer only from retrieved context
// - Few-shot: include 1-2 domain examples
// - CoT: provide a short, explicit explanation (2-4 bullets) based on context
//   (we do NOT expose private chain-of-thought; we produce a concise justification from sources).

const systemRU = `Ты — помощник по базе знаний.

Правила:
1) Отвечай ТОЛЬКО на основе context ниже (это фрагменты документов).
2) Если в context нет ответа или уверенность низкая — скажи строго: "Я не знаю".
3) Не выдумывай фактов. Если данных нет — так и скажи.
4) Структура ответа:
   - Ответ (1–3 предложения)
   - Короткое объяснение (2–4 пункта) — только факты из context
   - Источники (список source)
Язык: русский.`

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

	topScore := 0.0
	if len(items) > 0 {
		topScore = items[0].Score
	}

	// Guardrail "I don't know" when retrieval is weak.
	if len(items) == 0 || topScore < s.Cfg.MinScore {
		return s.appendSources("Я не знаю.\n\nУточните вопрос или добавьте информацию в базу знаний.", items), nil
	}

	ctx := formatContext(items)
	if len([]rune(ctx)) < s.Cfg.MinContextChars {
		return s.appendSources("Я не знаю.\n\nУточните вопрос (сущность/технология/событие) или попробуйте переформулировать.", items), nil
	}

	fewshot := loadFewShot(s.Cfg.FewShotPath)

	user := fmt.Sprintf("%s\n\nQ: %s\nA:", fewshot, strings.TrimSpace(question))
	system := systemRU + "\n\ncontext:\n" + ctx

	ans, err := s.LLM.Chat(system, user)
	if err != nil {
		return "", err
	}
	return s.appendSources(strings.TrimSpace(ans), items), nil
}

func loadFewShot(path string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		// Soft-fail: bot still works without few-shot examples.
		return ""
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return ""
	}
	return s
}

func formatContext(items []ContextItem) string {
	var b strings.Builder
	for i, it := range items {
		text := strings.ReplaceAll(strings.TrimSpace(it.Text), "\n", " ")
		r := []rune(text)
		if len(r) > 900 {
			text = string(r[:900]) + "…"
		}
		b.WriteString(fmt.Sprintf("[%d] source=%s score=%.4f | %s\n", i+1, it.Source, it.Score, text))
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
		return answer + "\n" + strings.Join(lines, "\n")
	}
	for _, it := range items {
		if seen[it.Source] {
			continue
		}
		seen[it.Source] = true
		sn := strings.Split(strings.TrimSpace(it.Text), "\n")[0]
		r := []rune(sn)
		if len(r) > 180 {
			sn = string(r[:180]) + "…"
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", it.Source, sn))
	}
	return answer + "\n" + strings.Join(lines, "\n")
}
