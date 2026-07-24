// domain/application/application.go
package application

import "errors"

// ApplicationID は申請の識別子。
// ただの string だが、型を分けることで他のIDとの取り違えをコンパイルエラーにできる
type ApplicationID string

type ApplicantID string
type ApproverID string

var (
	ErrNotSubmitted = errors.New("申請中の申請のみ承認できます")
	ErrSelfApproval = errors.New("自分の申請を自分で承認することはできません")
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
	if a.status != StatusDraft {
		return errors.New("下書きの申請のみ提出できます")
	}
	a.status = StatusSubmitted
	return nil
}

// Approve は申請を承認する。
// 「承認者は申請者と同一人物であってはならない」というビジネスルールはここに書く
func (a *Application) Approve(approverID ApproverID) error {
	if a.status != StatusSubmitted {
		return ErrNotSubmitted
	}
	if string(approverID) == string(a.applicantID) {
		return ErrSelfApproval
	}
	a.status = StatusApproved
	return nil
}

func (a *Application) ID() ApplicationID { return a.id }
func (a *Application) Status() Status    { return a.status }
