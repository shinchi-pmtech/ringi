// usecase/approve_application_test.go
package usecase_test

import (
	"errors"
	"testing"

	"github.com/shinchi-pmtech/ringi/domain/application"
	"github.com/shinchi-pmtech/ringi/infrastructure/persistence"
	"github.com/shinchi-pmtech/ringi/usecase"
)

// インメモリ実装をテストダブルとして使う。DBなしでユースケースを検証できる
func TestApproveApplication_自己承認は拒否される(t *testing.T) {
	repo := persistence.NewApplicationMemoryRepository()
	app := application.NewApplication("APP-001", "tanaka", "開発端末の購入")
	if err := app.Submit(); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(app); err != nil {
		t.Fatal(err)
	}

	u := usecase.NewApproveApplication(repo)
	err := u.Execute("APP-001", "tanaka")

	if !errors.Is(err, application.ErrSelfApproval) {
		t.Errorf("自己承認が拒否されませんでした: %v", err)
	}
}

func TestApproveApplication_別人なら承認できる(t *testing.T) {
	repo := persistence.NewApplicationMemoryRepository()
	app := application.NewApplication("APP-001", "tanaka", "開発端末の購入")
	if err := app.Submit(); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(app); err != nil {
		t.Fatal(err)
	}

	u := usecase.NewApproveApplication(repo)
	if err := u.Execute("APP-001", "suzuki"); err != nil {
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
