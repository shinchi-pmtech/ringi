// domain/application/approval.go
package application

import "errors"

var (
	ErrEmptyRoute     = errors.New("承認ルートが空です")
	ErrDuplicateRoute = errors.New("同じ承認者が承認ルートに複数含まれています")
	ErrNotYourTurn    = errors.New("現在の承認者ではありません")
)

// ApprovalRoute は承認者の並び。「課長 → 部長」のような多段承認を表す。
// 本来は組織のマスタから引いてくる想定だが、本連載では外から渡す形にしている
type ApprovalRoute []ApproverID

// validate は承認ルートとして成立しているかを検証する
func (r ApprovalRoute) validate(applicantID ApplicantID) error {
	if len(r) == 0 {
		return ErrEmptyRoute
	}
	seen := map[ApproverID]bool{}
	for _, approverID := range r {
		if string(approverID) == string(applicantID) {
			return ErrSelfApproval
		}
		if seen[approverID] {
			return ErrDuplicateRoute
		}
		seen[approverID] = true
	}
	return nil
}

// ApprovalStep は承認ルートの1段分。何段目を誰が承認するか、承認済みかを持つ。
// Application 集約の内側でのみ生成・変更される
type ApprovalStep struct {
	order      int
	approverID ApproverID
	approved   bool
}

func (s ApprovalStep) Order() int             { return s.order }
func (s ApprovalStep) ApproverID() ApproverID { return s.approverID }
func (s ApprovalStep) Approved() bool         { return s.approved }

// ReconstructStep は永続化された値から ApprovalStep を復元する。
// リポジトリ実装(infrastructure層)からの利用を想定している
func ReconstructStep(order int, approverID ApproverID, approved bool) ApprovalStep {
	return ApprovalStep{
		order:      order,
		approverID: approverID,
		approved:   approved,
	}
}
