// domain/application/approval_test.go
package application

import (
	"errors"
	"testing"
)

func newSubmitted(t *testing.T) *Application {
	t.Helper()
	amount, err := NewMoney(150000)
	if err != nil {
		t.Fatal(err)
	}
	app, err := NewApplication("APP-001", "tanaka", "開発端末の購入", amount,
		ApprovalRoute{"kacho", "bucho"})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Submit(); err != nil {
		t.Fatal(err)
	}
	return app
}

// 1段目を承認しても、まだ「申請中」のまま
func TestApprove_中間の段では承認済みにならない(t *testing.T) {
	app := newSubmitted(t)

	if err := app.Approve("kacho"); err != nil {
		t.Fatal(err)
	}
	if app.Status() != StatusSubmitted {
		t.Errorf("1段目の承認後の状態が %s になっています", app.Status())
	}
}

// 最終段まで承認されると「承認済み」になる
func TestApprove_最終段で承認済みになる(t *testing.T) {
	app := newSubmitted(t)

	if err := app.Approve("kacho"); err != nil {
		t.Fatal(err)
	}
	if err := app.Approve("bucho"); err != nil {
		t.Fatal(err)
	}
	if app.Status() != StatusApproved {
		t.Errorf("最終段の承認後の状態が %s になっています", app.Status())
	}
}

// 順番を飛ばした承認はできない
func TestApprove_順番を飛ばせない(t *testing.T) {
	app := newSubmitted(t)

	err := app.Approve("bucho") // 1段目は kacho
	if !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("順番を飛ばした承認が通ってしまいました: %v", err)
	}
}

// 承認ルートに含まれない人は承認できない
func TestApprove_ルート外の人は承認できない(t *testing.T) {
	app := newSubmitted(t)

	err := app.Approve("suzuki")
	if !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("ルート外の承認が通ってしまいました: %v", err)
	}
}

// 差戻した時点で、それまでの承認は無効になる
func TestReject_承認の進捗がリセットされる(t *testing.T) {
	app := newSubmitted(t)

	if err := app.Approve("kacho"); err != nil {
		t.Fatal(err)
	}
	if err := app.Reject("bucho"); err != nil {
		t.Fatal(err)
	}

	for _, s := range app.Steps() {
		if s.Approved() {
			t.Errorf("差戻し後に %d段目の承認が残っています", s.Order())
		}
	}
}

// 再提出すると、1段目の承認者から承認をやり直す
func TestResubmit_1段目からやり直しになる(t *testing.T) {
	app := newSubmitted(t)

	if err := app.Approve("kacho"); err != nil {
		t.Fatal(err)
	}
	if err := app.Reject("bucho"); err != nil {
		t.Fatal(err)
	}
	if err := app.Resubmit(); err != nil {
		t.Fatal(err)
	}

	// 2段目の承認者はまだ承認できない
	if err := app.Approve("bucho"); !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("再提出後にいきなり2段目の承認が通りました: %v", err)
	}
	// 1段目からやり直せる
	if err := app.Approve("kacho"); err != nil {
		t.Errorf("再提出後に1段目の承認ができません: %v", err)
	}
}

// 申請者自身を承認ルートに含めることはできない
func TestNewApplication_申請者は承認ルートに入れない(t *testing.T) {
	amount, _ := NewMoney(150000)
	_, err := NewApplication("APP-001", "tanaka", "開発端末の購入", amount,
		ApprovalRoute{"kacho", "tanaka"})
	if !errors.Is(err, ErrSelfApproval) {
		t.Errorf("申請者を含むルートが通ってしまいました: %v", err)
	}
}

// 同じ承認者が複数回登場するルートは作れない
func TestNewApplication_重複した承認者は入れない(t *testing.T) {
	amount, _ := NewMoney(150000)
	_, err := NewApplication("APP-001", "tanaka", "開発端末の購入", amount,
		ApprovalRoute{"kacho", "kacho"})
	if !errors.Is(err, ErrDuplicateRoute) {
		t.Errorf("重複したルートが通ってしまいました: %v", err)
	}
}

// Steps() の戻り値を書き換えても、集約の内部には影響しない
func TestSteps_戻り値を書き換えても内部は変わらない(t *testing.T) {
	app := newSubmitted(t)

	steps := app.Steps()
	steps[0].approved = true

	if app.Steps()[0].Approved() {
		t.Error("外から集約の内部を書き換えられてしまいました")
	}
}
