// domain/application/status.go
package application

import "fmt"

// Status は申請の状態を表す値オブジェクト
type Status string

const (
	StatusDraft     Status = "draft"     // 下書き
	StatusSubmitted Status = "submitted" // 申請中
	StatusApproved  Status = "approved"  // 承認済み
	StatusRejected  Status = "rejected"  // 差戻し
)

// transitions は許可される状態遷移の一覧。
// 「どの状態からどこへ動けるか」というドメインルールはここに集約する
var transitions = map[Status][]Status{
	StatusDraft:     {StatusSubmitted},                // 下書き → 提出
	StatusSubmitted: {StatusApproved, StatusRejected}, // 申請中 → 承認 or 差戻し
	StatusRejected:  {StatusSubmitted},                // 差戻し → 再提出
	StatusApproved:  {},                               // 承認済みは終端状態
}

// NewStatus は外部入力(DBの値、JSONなど)から Status を復元する。
// 定義外の値はエラーにする(値オブジェクトの自己検証)
func NewStatus(value string) (Status, error) {
	s := Status(value)
	if _, ok := transitions[s]; !ok {
		return "", fmt.Errorf("不正な状態です: %q", value)
	}
	return s, nil
}

// CanTransitionTo は next への遷移が許可されているかを返す
func (s Status) CanTransitionTo(next Status) bool {
	for _, allowed := range transitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}
