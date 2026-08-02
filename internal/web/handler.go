package web

import (
	"encoding/json"
	"io"
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
	mux.HandleFunc("GET /api/export", handler.exportData)
	mux.HandleFunc("POST /api/import", handler.importData)
	mux.HandleFunc("GET /api/sessions", handler.listSessions)
	mux.HandleFunc("POST /api/sessions", handler.createSession)
	mux.HandleFunc("PUT /api/sessions/{sessionID}", handler.updateSession)
	mux.HandleFunc("DELETE /api/sessions/{sessionID}", handler.deleteSession)
	mux.HandleFunc("GET /api/sessions/{sessionID}/games", handler.listGames)
	mux.HandleFunc("POST /api/sessions/{sessionID}/games", handler.createGame)
	mux.HandleFunc("PUT /api/sessions/{sessionID}/games/{gameID}", handler.updateGame)
	mux.HandleFunc("DELETE /api/sessions/{sessionID}/games/{gameID}", handler.deleteGame)

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

func (h *Handler) exportData(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.repository.ListSessions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "エクスポートデータを取得できませんでした")
		return
	}
	backup := service.Backup{
		Format:     service.BackupFormat,
		Version:    service.BackupVersion,
		ExportedAt: time.Now(),
		Sessions:   make([]service.BackupSession, 0, len(sessions)),
	}
	for _, session := range sessions {
		games, err := h.repository.ListGames(r.Context(), session.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "エクスポートデータを取得できませんでした")
			return
		}
		exportSession := service.BackupSession{
			Name:     session.Name,
			PlayedAt: session.PlayedAt.Format("2006-01-02"),
			Games:    make([]service.BackupGame, 0, len(games)),
		}
		for _, game := range games {
			exportSession.Games = append(exportSession.Games, service.BackupGame{
				RuleCitation: game.RuleCitation,
				CreatedAt:    game.CreatedAt,
				Results:      game.Results,
			})
		}
		backup.Sessions = append(backup.Sessions, exportSession)
	}
	w.Header().Set("Content-Disposition", `attachment; filename="mahjong-score-backup.json"`)
	writeJSON(w, http.StatusOK, backup)
}

func (h *Handler) importData(w http.ResponseWriter, r *http.Request) {
	const maxBackupSize = 10 << 20
	var backup service.Backup
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBackupSize))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&backup); err != nil {
		writeError(w, http.StatusBadRequest, "バックアップファイルの形式が正しくありません")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		writeError(w, http.StatusBadRequest, "バックアップファイルの形式が正しくありません")
		return
	}
	sessions, err := service.ValidateBackup(backup)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.repository.ImportSessions(r.Context(), sessions); err != nil {
		writeError(w, http.StatusInternalServerError, "バックアップをインポートできませんでした")
		return
	}
	gameCount := 0
	for _, session := range sessions {
		gameCount += len(session.Games)
	}
	writeJSON(w, http.StatusCreated, map[string]int{
		"sessions": len(sessions),
		"games":    gameCount,
	})
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
	h.saveSession(w, r, 0)
}

func (h *Handler) updateSession(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := positiveIDFromRequest(w, r, "sessionID", "対局日")
	if !ok {
		return
	}
	h.saveSession(w, r, sessionID)
}

func (h *Handler) saveSession(w http.ResponseWriter, r *http.Request, sessionID int64) {
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

	session := domain.Session{ID: sessionID, Name: input.Name, PlayedAt: playedAt}
	if sessionID > 0 {
		found, err := h.repository.UpdateSession(r.Context(), session)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "対局日を更新できませんでした")
			return
		}
		if !found {
			writeError(w, http.StatusNotFound, "対局日が見つかりません")
			return
		}
		writeJSON(w, http.StatusOK, session)
		return
	}

	session, err = h.repository.CreateSession(r.Context(), session)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "対局日を保存できませんでした")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (h *Handler) deleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := sessionIDFromRequest(w, r)
	if !ok {
		return
	}
	found, err := h.repository.DeleteSession(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "対局日を削除できませんでした")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "対局日が見つかりません")
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
	h.saveGame(w, r, 0)
}

func (h *Handler) updateGame(w http.ResponseWriter, r *http.Request) {
	gameID, ok := positiveIDFromRequest(w, r, "gameID", "半荘")
	if !ok {
		return
	}
	h.saveGame(w, r, gameID)
}

func (h *Handler) saveGame(w http.ResponseWriter, r *http.Request, gameID int64) {
	sessionID, ok := sessionIDFromRequest(w, r)
	if !ok {
		return
	}
	var input struct {
		RuleCitation *domain.RuleCitation `json:"ruleCitation"`
		Results      []domain.Result      `json:"results"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "入力形式が正しくありません")
		return
	}
	results, err := service.ValidateResults(input.Results)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	citation, err := service.ValidateRuleCitation(input.RuleCitation)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	game := domain.Game{ID: gameID, SessionID: sessionID, RuleCitation: citation, Results: results}
	if gameID > 0 {
		found, err := h.repository.UpdateGame(r.Context(), game)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "半荘結果を更新できませんでした")
			return
		}
		if !found {
			writeError(w, http.StatusNotFound, "半荘結果が見つかりません")
			return
		}
		writeJSON(w, http.StatusOK, game)
		return
	}

	game, err = h.repository.CreateGame(r.Context(), game)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "半荘結果を保存できませんでした")
		return
	}
	writeJSON(w, http.StatusCreated, game)
}

func (h *Handler) deleteGame(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := sessionIDFromRequest(w, r)
	if !ok {
		return
	}
	gameID, ok := positiveIDFromRequest(w, r, "gameID", "半荘")
	if !ok {
		return
	}
	found, err := h.repository.DeleteGame(r.Context(), sessionID, gameID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "半荘結果を削除できませんでした")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "半荘結果が見つかりません")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func sessionIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return positiveIDFromRequest(w, r, "sessionID", "対局日")
}

func positiveIDFromRequest(w http.ResponseWriter, r *http.Request, key, label string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, label+"IDが正しくありません")
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
