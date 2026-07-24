// domain/application/status.go
package application

type Status string

const (
	StatusDraft     Status = "draft"     // 下書き
	StatusSubmitted Status = "submitted" // 申請中
	StatusApproved  Status = "approved"  // 承認済み
	StatusRejected  Status = "rejected"  // 差戻し
)
