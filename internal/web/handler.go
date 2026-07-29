package web

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tkmbr/mahjong-score/internal/domain"
	"github.com/tkmbr/mahjong-score/internal/service"
)

type Handler struct {
	repository domain.Repository
}

func NewHandler(repository domain.Repository) http.Handler {
	handler := &Handler{repository: repository}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handler.health)
	mux.HandleFunc("GET /api/sessions", handler.listSessions)
	mux.HandleFunc("POST /api/sessions", handler.createSession)
	mux.HandleFunc("GET /api/sessions/{sessionID}/games", handler.listGames)
	mux.HandleFunc("POST /api/sessions/{sessionID}/games", handler.createGame)

	static, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(static)))
	return recoverMiddleware(logMiddleware(mux))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.repository.ListSessions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "対局日を取得できませんでした")
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		PlayedAt string `json:"playedAt"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "入力形式が正しくありません")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "対局日の名前を入力してください")
		return
	}
	playedAt, err := time.Parse("2006-01-02", input.PlayedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "日付が正しくありません")
		return
	}

	session, err := h.repository.CreateSession(r.Context(), domain.Session{Name: input.Name, PlayedAt: playedAt})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "対局日を保存できませんでした")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (h *Handler) listGames(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := sessionIDFromRequest(w, r)
	if !ok {
		return
	}
	games, err := h.repository.ListGames(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "半荘結果を取得できませんでした")
		return
	}
	writeJSON(w, http.StatusOK, games)
}

func (h *Handler) createGame(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := sessionIDFromRequest(w, r)
	if !ok {
		return
	}
	var input struct {
		Results []domain.Result `json:"results"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "入力形式が正しくありません")
		return
	}
	results, err := service.Calculate(input.Results, domain.StandardRule)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	game, err := h.repository.CreateGame(r.Context(), domain.Game{SessionID: sessionID, Results: results})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "半荘結果を保存できませんでした")
		return
	}
	writeJSON(w, http.StatusCreated, game)
}

func sessionIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("sessionID"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "対局日IDが正しくありません")
		return 0, false
	}
	return id, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
