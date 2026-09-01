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

// NewApplicationSQLiteRepository はリポジトリを生成し、テーブルを作り直す。
//
// 既存のテーブルを削除してから作るのは、連載でスキーマが変わっても
// 古い ringi.db が残っている環境でそのまま動かせるようにするため。
// 本来スキーマの変更はマイグレーションの領分だが、本連載では扱わない。
func NewApplicationSQLiteRepository(db *sql.DB) (*ApplicationSQLiteRepository, error) {
	_, err := db.Exec(`
		DROP TABLE IF EXISTS applications;
		DROP TABLE IF EXISTS approval_steps;
		CREATE TABLE applications (
			id           TEXT PRIMARY KEY,
			applicant_id TEXT NOT NULL,
			title        TEXT NOT NULL,
			amount       INTEGER NOT NULL,
			status       TEXT NOT NULL
		);
		CREATE TABLE approval_steps (
			application_id TEXT    NOT NULL,
			step_order     INTEGER NOT NULL,
			approver_id    TEXT    NOT NULL,
			approved       INTEGER NOT NULL,
			PRIMARY KEY (application_id, step_order)
		)`)
	if err != nil {
		return nil, fmt.Errorf("テーブル作成に失敗しました: %w", err)
	}
	return &ApplicationSQLiteRepository{db: db}, nil
}

// Save は申請と承認ステップをまとめて保存する。
// 集約は常に丸ごと保存されるので、2つのテーブルへの書き込みは1つのトランザクションにまとめる
func (r *ApplicationSQLiteRepository) Save(app *application.Application) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Commit 済みなら何も起きない

	if _, err := tx.Exec(`
		INSERT INTO applications (id, applicant_id, title, amount, status)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			applicant_id = excluded.applicant_id,
			title        = excluded.title,
			amount       = excluded.amount,
			status       = excluded.status`,
		string(app.ID()), string(app.ApplicantID()), app.Title(), app.Amount().Yen(), string(app.Status()),
	); err != nil {
		return err
	}

	// ステップは差分更新せず、いったん消してから入れ直す。
	// 集約の内側の構造をリポジトリが知りすぎないための割り切り
	if _, err := tx.Exec(
		`DELETE FROM approval_steps WHERE application_id = ?`, string(app.ID()),
	); err != nil {
		return err
	}
	for _, step := range app.Steps() {
		if _, err := tx.Exec(`
			INSERT INTO approval_steps (application_id, step_order, approver_id, approved)
			VALUES (?, ?, ?, ?)`,
			string(app.ID()), step.Order(), string(step.ApproverID()), step.Approved(),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// FindByID は申請と承認ステップをまとめて読み出し、集約として組み立て直す
func (r *ApplicationSQLiteRepository) FindByID(id application.ApplicationID) (*application.Application, error) {
	row := r.db.QueryRow(
		`SELECT applicant_id, title, amount, status FROM applications WHERE id = ?`,
		string(id),
	)

	var applicantID, title, statusRaw string
	var yen int
	if err := row.Scan(&applicantID, &title, &yen, &statusRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("application not found")
		}
		return nil, err
	}

	// DBから来た値は「外の世界」の値。値オブジェクトの自己検証を必ず通す
	status, err := application.NewStatus(statusRaw)
	if err != nil {
		return nil, fmt.Errorf("復元に失敗しました: %w", err)
	}
	amount, err := application.NewMoney(yen)
	if err != nil {
		return nil, fmt.Errorf("復元に失敗しました: %w", err)
	}

	steps, err := r.findSteps(id)
	if err != nil {
		return nil, err
	}

	return application.Reconstruct(
		id,
		application.ApplicantID(applicantID),
		title,
		amount,
		status,
		steps,
	), nil
}

func (r *ApplicationSQLiteRepository) findSteps(id application.ApplicationID) ([]application.ApprovalStep, error) {
	rows, err := r.db.Query(`
		SELECT step_order, approver_id, approved
		FROM approval_steps
		WHERE application_id = ?
		ORDER BY step_order`,
		string(id),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []application.ApprovalStep
	for rows.Next() {
		var order int
		var approverID string
		var approved bool
		if err := rows.Scan(&order, &approverID, &approved); err != nil {
			return nil, err
		}
		steps = append(steps, application.ReconstructStep(
			order, application.ApproverID(approverID), approved,
		))
	}
	return steps, rows.Err()
}
