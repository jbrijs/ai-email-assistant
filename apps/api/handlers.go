package main

import (
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func ThreadsListHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: query Postgres
	JSON(w, http.StatusOK, map[string]any{
		"threads": []any{},
	})
}

func ThreadDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Path params with stdlib (Go 1.22+):
	id := r.PathValue("id")
	// TODO: fetch thread + summary by id
	JSON(w, http.StatusOK, map[string]any{
		"id":      id,
		"subject": "stub",
	})
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: read body {query:string}, do embedding + pgvector KNN
	JSON(w, http.StatusOK, map[string]any{
		"results": []any{},
	})
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: read {question}, retrieve K docs, call LLM, return answer + citations
	JSON(w, http.StatusOK, map[string]any{
		"answer":    "stub",
		"citations": []any{},
	})
}
