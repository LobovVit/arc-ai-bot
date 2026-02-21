package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"

	"quantumforge-rag-telegram-bot-go/internal/config"
	"quantumforge-rag-telegram-bot-go/internal/docs"
	"quantumforge-rag-telegram-bot-go/internal/embeddings"
	"quantumforge-rag-telegram-bot-go/internal/qdrant"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	emb := embeddings.NewLocalHTTP(cfg.EmbeddingsURL)
	q := qdrant.New(cfg.QdrantURL)

	dim := emb.Dim()
	log.Printf("Embeddings dim=%d", dim)

	if err := q.DeleteCollection(cfg.QdrantCollection); err != nil {
		log.Fatal(err)
	}
	if err := q.CreateCollection(cfg.QdrantCollection, dim); err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	chunks, err := docs.LoadAndChunk(cfg.DocsPath, cfg.ChunkSize, cfg.ChunkOverlap)
	if err != nil {
		log.Fatal(err)
	}
	if len(chunks) == 0 {
		log.Fatalf("No .md/.txt docs found in %s", cfg.DocsPath)
	}

	const batchSize = 64
	points := make([]qdrant.Point, 0, batchSize)
	texts := make([]string, 0, batchSize)
	id := 1
	total := 0

	flush := func() {
		if len(texts) == 0 {
			return
		}
		vecs, err := emb.EmbedBatch(texts)
		if err != nil {
			log.Fatal(err)
		}
		for i := range points {
			points[i].Vector = vecs[i]
		}
		if err := q.Upsert(cfg.QdrantCollection, points); err != nil {
			log.Fatal(err)
		}
		texts = texts[:0]
		points = points[:0]
	}

	for _, ch := range chunks {
		texts = append(texts, ch.Text)
		points = append(points, qdrant.Point{
			ID: id,
			Payload: map[string]interface{}{
				"source":   ch.Source,
				"chunk_id": ch.ChunkID,
				"text":     ch.Text,
			},
		})
		id++
		total++
		if len(texts) >= batchSize {
			flush()
		}
	}
	flush()

	log.Printf("Ingest completed at %s. chunks=%d time=%s", time.Now().Format(time.RFC3339), total, time.Since(start))
}
