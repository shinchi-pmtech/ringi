// domain/application/route_policy.go
package application

import "errors"

var ErrApproverNotFound = errors.New("承認者が見つかりません")

// Position は組織上の役職
type Position string

const (
	PositionKacho Position = "課長"
	PositionBucho Position = "部長"
)

// ApproverResolver は「その申請者にとっての課長は誰か」を答える。
// 組織図の知識は申請ドメインの外にあるので、interface として要求だけを書く
type ApproverResolver interface {
	Resolve(applicantID ApplicantID, position Position) (ApproverID, error)
}

// buchoThreshold はこの金額以上の申請に部長承認が必要になる境界
var buchoThreshold = Money{yen: 100000}

// DecideApprovalRoute は申請金額から承認ルートを決める。
//
// この判定は申請にも承認ルートにも属さない。金額(申請の情報)と
// 承認の閾値(組織のルール)と組織図(誰が課長か)という、
// 3つの知識をまたぐためである。どのモデルにも属さないが
// ドメインの知識であるものは、こうして独立した関数として置く。
func DecideApprovalRoute(applicantID ApplicantID, amount Money, resolver ApproverResolver) (ApprovalRoute, error) {
	kacho, err := resolver.Resolve(applicantID, PositionKacho)
	if err != nil {
		return nil, err
	}

	route := ApprovalRoute{kacho}

	// 10万円以上は部長の承認も必要
	if amount.GreaterThanOrEqual(buchoThreshold) {
		bucho, err := resolver.Resolve(applicantID, PositionBucho)
		if err != nil {
			return nil, err
		}
		route = append(route, bucho)
	}

	return route, nil
}
