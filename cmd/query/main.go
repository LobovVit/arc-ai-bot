package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"quantumforge-rag-telegram-bot-go/internal/config"
	"quantumforge-rag-telegram-bot-go/internal/embeddings"
	"quantumforge-rag-telegram-bot-go/internal/qdrant"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if len(os.Args) < 2 {
		log.Fatal("Usage: /app/bin/query ваш запрос")
	}
	query := strings.Join(os.Args[1:], " ")

	emb := embeddings.NewLocalHTTP(cfg.EmbeddingsURL)
	q := qdrant.New(cfg.QdrantURL)

	vec, err := emb.Embed(query)
	if err != nil {
		log.Fatal(err)
	}

	payloads, err := q.Search(cfg.QdrantCollection, vec, cfg.TopK)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Query: %s\n\nTop-%d results:\n\n", query, cfg.TopK)
	for i, p := range payloads {
		src, _ := p["source"].(string)
		chunkID := p["chunk_id"]
		text, _ := p["text"].(string)
		text = strings.ReplaceAll(strings.TrimSpace(text), "\n", " ")
		r := []rune(text)
		if len(r) > 220 {
			text = string(r[:220]) + "…"
		}
		score, _ := p["_score"].(float64)
		fmt.Printf("%d) score=%.4f source=%s chunk_id=%v\n   %s\n\n", i+1, score, src, chunkID, text)
	}
}
