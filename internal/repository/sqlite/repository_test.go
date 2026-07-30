package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

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
	if found, err := repository.UpdateGame(ctx, game); err != nil || !found {
		t.Fatalf("UpdateGame() = (%v, %v), want (true, nil)", found, err)
	}

	games, err := repository.ListGames(ctx, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 || games[0].Results[0].Score != 40 {
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
