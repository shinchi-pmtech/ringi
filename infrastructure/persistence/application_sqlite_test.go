// infrastructure/persistence/application_sqlite_test.go
package persistence

import (
	"database/sql"
	"testing"

	"github.com/shinchi-pmtech/ringi/domain/application"
	_ "modernc.org/sqlite"
)

func newTestRepo(t *testing.T) (*ApplicationSQLiteRepository, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	repo, err := NewApplicationSQLiteRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	return repo, db
}

// 保存して読み戻すと、同じ内容の申請が復元される
func TestApplicationSQLiteRepository_SaveAndFindByID(t *testing.T) {
	repo, _ := newTestRepo(t)

	app := application.NewApplication("APP-001", "tanaka", "開発端末の購入")
	if err := app.Submit(); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(app); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByID("APP-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID() != app.ID() || got.ApplicantID() != app.ApplicantID() ||
		got.Title() != app.Title() || got.Status() != app.Status() {
		t.Errorf("復元結果が一致しません: got %+v, want %+v", got, app)
	}
}

// DBに直接仕込まれた不正な状態は、復元時に NewStatus が弾く
func TestApplicationSQLiteRepository_FindByID_InvalidStatus(t *testing.T) {
	repo, db := newTestRepo(t)

	app := application.NewApplication("APP-001", "tanaka", "開発端末の購入")
	if err := repo.Save(app); err != nil {
		t.Fatal(err)
	}

	// リポジトリを迂回してDBを直接書き換える(外の世界からの不正データを模す)
	if _, err := db.Exec(`UPDATE applications SET status = 'banana' WHERE id = 'APP-001'`); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.FindByID("APP-001"); err == nil {
		t.Error("不正な状態が復元できてしまいました")
	}
}

// 存在しないIDはエラーになる
func TestApplicationSQLiteRepository_FindByID_NotFound(t *testing.T) {
	repo, _ := newTestRepo(t)
	if _, err := repo.FindByID("APP-999"); err == nil {
		t.Error("存在しないIDでエラーになりませんでした")
	}
}
