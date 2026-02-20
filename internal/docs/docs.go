package docs

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type DocChunk struct {
	Source  string
	ChunkID int
	Text    string
}

func LoadAndChunk(root string, chunkSize int, overlap int) ([]DocChunk, error) {
	var chunks []DocChunk
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(b)
		rel, _ := filepath.Rel(root, path)
		cid := 0
		for _, c := range splitIntoChunks(text, chunkSize, overlap) {
			cc := strings.TrimSpace(c)
			if cc == "" {
				continue
			}
			chunks = append(chunks, DocChunk{Source: rel, ChunkID: cid, Text: cc})
			cid++
		}
		return nil
	})
	return chunks, err
}

func splitIntoChunks(s string, size int, overlap int) []string {
	if size <= 0 {
		return []string{s}
	}
	r := []rune(s)
	step := size - overlap
	if step <= 0 {
		step = size
	}
	var out []string
	for start := 0; start < len(r); start += step {
		end := start + size
		if end > len(r) {
			end = len(r)
		}
		out = append(out, string(r[start:end]))
		if end == len(r) {
			break
		}
	}
	return out
}
