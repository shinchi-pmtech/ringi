// domain/application/route_policy_test.go
package application

import (
	"errors"
	"testing"
)

// テスト用の組織図。ドメインサービスのテストにDBは要らない
type fakeResolver struct {
	positions map[Position]ApproverID
}

func (f fakeResolver) Resolve(_ ApplicantID, position Position) (ApproverID, error) {
	approverID, ok := f.positions[position]
	if !ok {
		return "", ErrApproverNotFound
	}
	return approverID, nil
}

func newResolver() fakeResolver {
	return fakeResolver{positions: map[Position]ApproverID{
		PositionKacho: "kacho",
		PositionBucho: "bucho",
	}}
}

func TestDecideApprovalRoute_金額で承認ルートが変わる(t *testing.T) {
	tests := []struct {
		name string
		yen  int
		want ApprovalRoute
	}{
		{"10万円未満は課長のみ", 99999, ApprovalRoute{"kacho"}},
		{"ちょうど10万円は部長まで", 100000, ApprovalRoute{"kacho", "bucho"}},
		{"10万円超は部長まで", 100001, ApprovalRoute{"kacho", "bucho"}},
		{"0円でも課長の承認は要る", 0, ApprovalRoute{"kacho"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := NewMoney(tt.yen)
			if err != nil {
				t.Fatal(err)
			}

			got, err := DecideApprovalRoute("tanaka", amount, newResolver())
			if err != nil {
				t.Fatal(err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("ルートの段数が %d です(期待 %d)", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("%d段目が %s です(期待 %s)", i+1, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDecideApprovalRoute_承認者が見つからない(t *testing.T) {
	amount, err := NewMoney(200000)
	if err != nil {
		t.Fatal(err)
	}

	// 部長が登録されていない組織図
	resolver := fakeResolver{positions: map[Position]ApproverID{
		PositionKacho: "kacho",
	}}

	if _, err := DecideApprovalRoute("tanaka", amount, resolver); !errors.Is(err, ErrApproverNotFound) {
		t.Errorf("承認者不在がエラーになりませんでした: %v", err)
	}
}

func TestNewMoney_負の金額は作れない(t *testing.T) {
	if _, err := NewMoney(-1); !errors.Is(err, ErrNegativeAmount) {
		t.Errorf("負の金額が作れてしまいました: %v", err)
	}
}
