import os
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
import uvicorn

MODEL_NAME = os.getenv("MODEL_NAME", "sentence-transformers/all-MiniLM-L6-v2")
app = FastAPI(title="Local Embeddings Service")

model = SentenceTransformer(MODEL_NAME)
DIM = model.get_sentence_embedding_dimension()

class EmbedRequest(BaseModel):
    texts: list[str]

class EmbedResponse(BaseModel):
    dim: int
    vectors: list[list[float]]

@app.get("/health")
def health():
    return {"status": "ok", "model": MODEL_NAME, "dim": DIM}

@app.post("/embed", response_model=EmbedResponse)
def embed(req: EmbedRequest):
    vecs = model.encode(req.texts, normalize_embeddings=True).tolist()
    return {"dim": DIM, "vectors": vecs}

if __name__ == "__main__":
    port = int(os.getenv("PORT", "8088"))
    uvicorn.run(app, host="0.0.0.0", port=port)
