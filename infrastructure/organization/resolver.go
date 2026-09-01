// infrastructure/organization/resolver.go
package organization

import (
	"fmt"

	"github.com/shinchi-pmtech/ringi/domain/application"
)

// MemoryApproverResolver は組織図をメモリ上に持つ簡易な実装。
// 実際には人事システムや組織マスタを参照することになる
type MemoryApproverResolver struct {
	// 申請者ID → 役職 → 承認者ID
	chart map[application.ApplicantID]map[application.Position]application.ApproverID
}

func NewMemoryApproverResolver() *MemoryApproverResolver {
	return &MemoryApproverResolver{
		chart: map[application.ApplicantID]map[application.Position]application.ApproverID{
			"tanaka": {
				application.PositionKacho: "kacho",
				application.PositionBucho: "bucho",
			},
		},
	}
}

func (r *MemoryApproverResolver) Resolve(
	applicantID application.ApplicantID,
	position application.Position,
) (application.ApproverID, error) {
	positions, ok := r.chart[applicantID]
	if !ok {
		return "", fmt.Errorf("%w: 申請者 %s の組織情報がありません",
			application.ErrApproverNotFound, applicantID)
	}
	approverID, ok := positions[position]
	if !ok {
		return "", fmt.Errorf("%w: 申請者 %s の%sが登録されていません",
			application.ErrApproverNotFound, applicantID, position)
	}
	return approverID, nil
}
