// domain/application/application.go
package application

import (
	"errors"
	"fmt"
)

// ApplicationID は申請の識別子。
// ただの string だが、型を分けることで他のIDとの取り違えをコンパイルエラーにできる
type ApplicationID string

type ApplicantID string
type ApproverID string

var (
	ErrSelfApproval      = errors.New("自分の申請を自分で承認・差戻しすることはできません")
	ErrInvalidTransition = errors.New("この状態からその操作はできません")
)

type Application struct {
	id          ApplicationID
	applicantID ApplicantID
	title       string
	status      Status
}

// NewApplication は「下書き」状態の申請を生成する
func NewApplication(id ApplicationID, applicantID ApplicantID, title string) *Application {
	return &Application{
		id:          id,
		applicantID: applicantID,
		title:       title,
		status:      StatusDraft,
	}
}

// Submit は申請を提出する(下書き → 申請中)
func (a *Application) Submit() error {
	return a.transitionTo(StatusSubmitted)
}

// Approve は申請を承認する(申請中 → 承認済み)。
// 「承認者は申請者と同一人物であってはならない」というビジネスルールはここに書く
func (a *Application) Approve(approverID ApproverID) error {
	if string(approverID) == string(a.applicantID) {
		return ErrSelfApproval
	}
	return a.transitionTo(StatusApproved)
}

// Reject は申請を差戻す(申請中 → 差戻し)
func (a *Application) Reject(approverID ApproverID) error {
	if string(approverID) == string(a.applicantID) {
		return ErrSelfApproval
	}
	return a.transitionTo(StatusRejected)
}

// Resubmit は差戻された申請を再提出する(差戻し → 申請中)。
// 実装は Submit と同じ遷移先だが、業務上は別のふるまいなので別メソッドにする
func (a *Application) Resubmit() error {
	return a.transitionTo(StatusSubmitted)
}

// transitionTo は遷移ルールを検証してから状態を変更する。
// 状態変更の入口をここ1箇所に絞る
func (a *Application) transitionTo(next Status) error {
	if !a.status.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, a.status, next)
	}
	a.status = next
	return nil
}

// Reconstruct は永続化された値から Application を復元する。
// 新規作成(NewApplication)と違い、状態を「下書き」に固定しない。
// リポジトリ実装(infrastructure層)からの利用を想定している
func Reconstruct(id ApplicationID, applicantID ApplicantID, title string, status Status) *Application {
	return &Application{
		id:          id,
		applicantID: applicantID,
		title:       title,
		status:      status,
	}
}

func (a *Application) ID() ApplicationID        { return a.id }
func (a *Application) ApplicantID() ApplicantID { return a.applicantID }
func (a *Application) Title() string            { return a.title }
func (a *Application) Status() Status           { return a.status }
