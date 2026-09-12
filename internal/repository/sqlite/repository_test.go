package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

func TestListGamesWithMixedTimezones(t *testing.T) {
	r, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	_, err = r.db.Exec(`
INSERT INTO sessions(id, name, played_at) VALUES(1, 'test', '2026-01-01T00:00:00Z');
INSERT INTO games(id, session_id, created_at) VALUES
 (1, 1, '2026-01-01T00:30:00+09:00'),
 (2, 1, '2025-12-31T16:00:00Z');
INSERT INTO game_results(game_id, position, player_name, score) VALUES
 (1, 0, 'A', 0), (2, 0, 'A', 0);`)
	if err != nil {
		t.Fatal(err)
	}
	games, err := r.ListGames(context.Background(), 1)
	if err != nil || len(games) != 2 || games[0].ID != 1 || games[1].ID != 2 {
		t.Fatalf("games=%v err=%v", games, err)
	}
}

func TestUpdateAndDeleteRecords(t *testing.T) {
	repository, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	ctx := context.Background()
	session, err := repository.CreateSession(ctx, domain.Session{
		Name:     "変更前",
		PlayedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	session.Name = "変更後"
	session.PlayedAt = time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	if found, err := repository.UpdateSession(ctx, session); err != nil || !found {
		t.Fatalf("UpdateSession() = (%v, %v), want (true, nil)", found, err)
	}

	game, err := repository.CreateGame(ctx, domain.Game{
		SessionID: session.ID,
		Results: []domain.Result{
			{PlayerName: "A", Score: 30},
			{PlayerName: "B", Score: 10},
			{PlayerName: "C", Score: -10},
			{PlayerName: "D", Score: -30},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	game.Results[0].Score = 40
	game.Results[1].Score = 0
	game.RuleCitation = &domain.RuleCitation{
		Revision: "8baf4cb3b3861d89badfd53424a6d5cce73e904e",
		Path:     "5等サンマ/rule.pdf",
		Title:    "5等サンマ",
	}
	if found, err := repository.UpdateGame(ctx, game); err != nil || !found {
		t.Fatalf("UpdateGame() = (%v, %v), want (true, nil)", found, err)
	}

	games, err := repository.ListGames(ctx, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 || games[0].Results[0].Score != 40 || games[0].RuleCitation == nil || games[0].RuleCitation.Title != "5等サンマ" {
		t.Fatalf("ListGames() after update = %#v", games)
	}
	if found, err := repository.DeleteGame(ctx, session.ID+1, game.ID); err != nil || found {
		t.Fatalf("DeleteGame() for another session = (%v, %v), want (false, nil)", found, err)
	}
	if found, err := repository.DeleteGame(ctx, session.ID, game.ID); err != nil || !found {
		t.Fatalf("DeleteGame() = (%v, %v), want (true, nil)", found, err)
	}

	game, err = repository.CreateGame(ctx, domain.Game{SessionID: session.ID, Results: game.Results})
	if err != nil {
		t.Fatal(err)
	}
	if found, err := repository.DeleteSession(ctx, session.ID); err != nil || !found {
		t.Fatalf("DeleteSession() = (%v, %v), want (true, nil)", found, err)
	}
	var resultCount int
	if err := repository.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_results WHERE game_id = ?`, game.ID,
	).Scan(&resultCount); err != nil {
		t.Fatal(err)
	}
	if resultCount != 0 {
		t.Fatalf("game result count after deleting session = %d, want 0", resultCount)
	}
}

func TestUpdatesReturnNotFound(t *testing.T) {
	repository, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	ctx := context.Background()
	if found, err := repository.UpdateSession(ctx, domain.Session{ID: 999, Name: "なし"}); err != nil || found {
		t.Fatalf("UpdateSession() = (%v, %v), want (false, nil)", found, err)
	}
	if found, err := repository.UpdateGame(ctx, domain.Game{ID: 999, SessionID: 999}); err != nil || found {
		t.Fatalf("UpdateGame() = (%v, %v), want (false, nil)", found, err)
	}
}

func TestImportSessions(t *testing.T) {
	repository, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	createdAt := time.Date(2026, 7, 30, 14, 30, 0, 0, time.FixedZone("JST", 9*60*60))
	laterCreatedAt := createdAt.Add(2 * time.Hour)
	results := []domain.Result{
		{PlayerName: "A", Score: 30},
		{PlayerName: "B", Score: 10},
		{PlayerName: "C", Score: -10},
		{PlayerName: "D", Score: -30},
	}
	citation := &domain.RuleCitation{
		Revision: "8baf4cb3b3861d89badfd53424a6d5cce73e904e",
		Path:     "5等サンマ/rule.pdf",
		Title:    "5等サンマ",
	}
	err = repository.ImportSessions(context.Background(), []domain.SessionWithGames{{
		Session: domain.Session{
			Name:     "インポート",
			PlayedAt: time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC),
		},
		Games: []domain.Game{
			{RuleCitation: citation, CreatedAt: laterCreatedAt, Results: results},
			{CreatedAt: createdAt, Results: results},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	sessions, err := repository.ListSessions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	games, err := repository.ListGames(context.Background(), sessions[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || len(games) != 2 || !games[0].CreatedAt.Equal(createdAt) || !games[1].CreatedAt.Equal(laterCreatedAt) {
		t.Fatalf("imported sessions = %#v, games = %#v", sessions, games)
	}
	if games[0].ID >= games[1].ID {
		t.Fatalf("imported game IDs = (%d, %d), want chronological IDs", games[0].ID, games[1].ID)
	}
	if games[0].RuleCitation != nil || games[1].RuleCitation == nil || games[1].RuleCitation.Title != "5等サンマ" {
		t.Fatalf("imported rule citations = (%#v, %#v)", games[0].RuleCitation, games[1].RuleCitation)
	}
}
