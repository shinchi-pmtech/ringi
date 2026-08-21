// usecase/approve_application_test.go
package usecase_test

import (
	"errors"
	"testing"

	"github.com/shinchi-pmtech/ringi/domain/application"
	"github.com/shinchi-pmtech/ringi/infrastructure/persistence"
	"github.com/shinchi-pmtech/ringi/usecase"
)

func newSubmittedApp(t *testing.T, repo application.Repository) {
	t.Helper()
	app, err := application.NewApplication("APP-001", "tanaka", "開発端末の購入",
		application.ApprovalRoute{"kacho", "bucho"})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Submit(); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(app); err != nil {
		t.Fatal(err)
	}
}

// インメモリ実装をテストダブルとして使う。DBなしでユースケースを検証できる
func TestApproveApplication_順番どおりでない承認は拒否される(t *testing.T) {
	repo := persistence.NewApplicationMemoryRepository()
	newSubmittedApp(t, repo)

	u := usecase.NewApproveApplication(repo)
	err := u.Execute("APP-001", "bucho") // 1段目は kacho

	if !errors.Is(err, application.ErrNotYourTurn) {
		t.Errorf("順番を飛ばした承認が拒否されませんでした: %v", err)
	}
}

func TestApproveApplication_全段の承認で承認済みになる(t *testing.T) {
	repo := persistence.NewApplicationMemoryRepository()
	newSubmittedApp(t, repo)

	u := usecase.NewApproveApplication(repo)
	if err := u.Execute("APP-001", "kacho"); err != nil {
		t.Fatal(err)
	}
	if err := u.Execute("APP-001", "bucho"); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByID("APP-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status() != application.StatusApproved {
		t.Errorf("承認後の状態が %s になっています", got.Status())
	}
}
