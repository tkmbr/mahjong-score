package service

import (
	"testing"
	"time"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

func TestValidateBackup(t *testing.T) {
	input := Backup{
		Format:  BackupFormat,
		Version: BackupVersion,
		Sessions: []BackupSession{{
			Name:     " 水曜セット ",
			PlayedAt: "2026-07-30",
			Games: []BackupGame{{
				CreatedAt: time.Date(2026, 7, 30, 14, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
				Results: []domain.Result{
					{PlayerName: " A ", Score: 30},
					{PlayerName: "B", Score: 10},
					{PlayerName: "C", Score: -10},
					{PlayerName: "D", Score: -30},
				},
			}},
		}},
	}

	got, err := ValidateBackup(input)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Session.Name != "水曜セット" || got[0].Games[0].Results[0].PlayerName != "A" {
		t.Fatalf("ValidateBackup() = %#v", got)
	}
}

func TestValidateBackupRejectsFormatAndInvalidGame(t *testing.T) {
	if _, err := ValidateBackup(Backup{Format: "other", Version: BackupVersion}); err == nil {
		t.Fatal("ValidateBackup() accepted another format")
	}
	input := Backup{
		Format:  BackupFormat,
		Version: BackupVersion,
		Sessions: []BackupSession{{
			Name:     "test",
			PlayedAt: "2026-07-30",
			Games: []BackupGame{{
				CreatedAt: time.Now(),
				Results:   []domain.Result{{PlayerName: "A", Score: 1}},
			}},
		}},
	}
	if _, err := ValidateBackup(input); err == nil {
		t.Fatal("ValidateBackup() accepted an invalid game")
	}
}
