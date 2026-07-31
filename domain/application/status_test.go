// domain/application/status_test.go
package application

import "testing"

// 遷移表をそのままテーブル駆動テストで検証する
func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{"下書きは提出できる", StatusDraft, StatusSubmitted, true},
		{"下書きをいきなり承認はできない", StatusDraft, StatusApproved, false},
		{"申請中は承認できる", StatusSubmitted, StatusApproved, true},
		{"申請中は差戻しできる", StatusSubmitted, StatusRejected, true},
		{"差戻しは再提出できる", StatusRejected, StatusSubmitted, true},
		{"差戻しをそのまま承認はできない", StatusRejected, StatusApproved, false},
		{"承認済みからは何もできない", StatusApproved, StatusSubmitted, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
				t.Errorf("%s → %s: got %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestNewStatus(t *testing.T) {
	if _, err := NewStatus("submitted"); err != nil {
		t.Errorf("正しい値でエラーになった: %v", err)
	}
	if _, err := NewStatus("banana"); err == nil {
		t.Error("不正な値がエラーにならなかった")
	}
}
