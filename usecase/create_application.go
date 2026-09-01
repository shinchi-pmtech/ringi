// usecase/create_application.go
package usecase

import "github.com/shinchi-pmtech/ringi/domain/application"

type CreateApplication struct {
	repo     application.Repository
	resolver application.ApproverResolver
}

func NewCreateApplication(
	repo application.Repository,
	resolver application.ApproverResolver,
) *CreateApplication {
	return &CreateApplication{repo: repo, resolver: resolver}
}

// Execute は申請を作成して保存する。
// 承認ルートの決定はドメインサービスに任せ、ここでは組み立てるだけ
func (u *CreateApplication) Execute(
	id application.ApplicationID,
	applicantID application.ApplicantID,
	title string,
	yen int,
) error {
	amount, err := application.NewMoney(yen)
	if err != nil {
		return err
	}

	route, err := application.DecideApprovalRoute(applicantID, amount, u.resolver)
	if err != nil {
		return err
	}

	app, err := application.NewApplication(id, applicantID, title, amount, route)
	if err != nil {
		return err
	}

	return u.repo.Save(app)
}
