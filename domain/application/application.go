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
	amount      Money
	status      Status
	steps       []ApprovalStep // 承認の進捗。集約の内側でのみ変更される
}

// NewApplication は「下書き」状態の申請を生成する。
// 承認ルートは申請時点の内容を写し取る(以後、組織のマスタが変わっても影響を受けない)
func NewApplication(id ApplicationID, applicantID ApplicantID, title string, amount Money, route ApprovalRoute) (*Application, error) {
	if err := route.validate(applicantID); err != nil {
		return nil, err
	}

	steps := make([]ApprovalStep, 0, len(route))
	for i, approverID := range route {
		steps = append(steps, ApprovalStep{
			order:      i + 1,
			approverID: approverID,
			approved:   false,
		})
	}

	return &Application{
		id:          id,
		applicantID: applicantID,
		title:       title,
		amount:      amount,
		status:      StatusDraft,
		steps:       steps,
	}, nil
}

// Submit は申請を提出する(下書き → 申請中)
func (a *Application) Submit() error {
	return a.transitionTo(StatusSubmitted)
}

// Approve は現在の段の承認者による承認を記録する。
// 最終段まで終わったときだけ、申請そのものが「承認済み」になる
func (a *Application) Approve(approverID ApproverID) error {
	step, err := a.currentStep(approverID)
	if err != nil {
		return err
	}
	if a.status != StatusSubmitted {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, a.status, StatusApproved)
	}

	step.approved = true

	if a.allApproved() {
		return a.transitionTo(StatusApproved)
	}
	return nil // まだ後続の段が残っているので、状態は「申請中」のまま
}

// Reject は現在の段の承認者による差戻し(申請中 → 差戻し)。
// 差戻しは承認プロセスの中断なので、それまでの承認はこの時点で無効になる
func (a *Application) Reject(approverID ApproverID) error {
	if _, err := a.currentStep(approverID); err != nil {
		return err
	}
	if err := a.transitionTo(StatusRejected); err != nil {
		return err
	}
	a.resetSteps()
	return nil
}

// Resubmit は差戻された申請を再提出する(差戻し → 申請中)。
// 承認の進捗は差戻しの時点でリセット済みなので、ここでは状態を戻すだけでよい
func (a *Application) Resubmit() error {
	return a.transitionTo(StatusSubmitted)
}

// resetSteps はすべての段を未承認に戻す
func (a *Application) resetSteps() {
	for i := range a.steps {
		a.steps[i].approved = false
	}
}

// currentStep は次に承認されるべき段を返す。
// approverID がその段の承認者でなければエラーにする
func (a *Application) currentStep(approverID ApproverID) (*ApprovalStep, error) {
	for i := range a.steps {
		if a.steps[i].approved {
			continue
		}
		if a.steps[i].approverID != approverID {
			return nil, fmt.Errorf("%w: %d段目の承認者は %s です",
				ErrNotYourTurn, a.steps[i].order, a.steps[i].approverID)
		}
		return &a.steps[i], nil
	}
	return nil, fmt.Errorf("%w: すべての段が承認済みです", ErrNotYourTurn)
}

func (a *Application) allApproved() bool {
	for _, s := range a.steps {
		if !s.approved {
			return false
		}
	}
	return true
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
// 新規作成(NewApplication)と違い、状態や進捗をそのまま組み立てる。
// リポジトリ実装(infrastructure層)からの利用を想定している
func Reconstruct(id ApplicationID, applicantID ApplicantID, title string, amount Money, status Status, steps []ApprovalStep) *Application {
	return &Application{
		id:          id,
		applicantID: applicantID,
		title:       title,
		amount:      amount,
		status:      status,
		steps:       steps,
	}
}

func (a *Application) ID() ApplicationID        { return a.id }
func (a *Application) ApplicantID() ApplicantID { return a.applicantID }
func (a *Application) Title() string            { return a.title }
func (a *Application) Amount() Money            { return a.amount }
func (a *Application) Status() Status           { return a.status }

// Steps は承認の進捗を読み取り専用で返す。
// スライスの複製を返すことで、外から要素を書き換えられないようにする
func (a *Application) Steps() []ApprovalStep {
	copied := make([]ApprovalStep, len(a.steps))
	copy(copied, a.steps)
	return copied
}
