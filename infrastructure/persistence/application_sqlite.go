// infrastructure/persistence/application_sqlite.go
package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/shinchi-pmtech/ringi/domain/application"
)

type ApplicationSQLiteRepository struct {
	db *sql.DB
}

// NewApplicationSQLiteRepository はリポジトリを生成し、テーブルがなければ作る。
// (本格的なマイグレーション管理は本連載では扱わない)
func NewApplicationSQLiteRepository(db *sql.DB) (*ApplicationSQLiteRepository, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS applications (
			id           TEXT PRIMARY KEY,
			applicant_id TEXT NOT NULL,
			title        TEXT NOT NULL,
			status       TEXT NOT NULL
		)`)
	if err != nil {
		return nil, fmt.Errorf("テーブル作成に失敗しました: %w", err)
	}
	return &ApplicationSQLiteRepository{db: db}, nil
}

func (r *ApplicationSQLiteRepository) Save(app *application.Application) error {
	_, err := r.db.Exec(`
		INSERT INTO applications (id, applicant_id, title, status)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			applicant_id = excluded.applicant_id,
			title        = excluded.title,
			status       = excluded.status`,
		string(app.ID()), string(app.ApplicantID()), app.Title(), string(app.Status()),
	)
	return err
}

func (r *ApplicationSQLiteRepository) FindByID(id application.ApplicationID) (*application.Application, error) {
	row := r.db.QueryRow(
		`SELECT applicant_id, title, status FROM applications WHERE id = ?`,
		string(id),
	)

	var applicantID, title, statusRaw string
	if err := row.Scan(&applicantID, &title, &statusRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("application not found")
		}
		return nil, err
	}

	// DBから来た文字列は「外の世界」の値。NewStatus の自己検証を必ず通す
	status, err := application.NewStatus(statusRaw)
	if err != nil {
		return nil, fmt.Errorf("復元に失敗しました: %w", err)
	}

	return application.Reconstruct(
		id,
		application.ApplicantID(applicantID),
		title,
		status,
	), nil
}
