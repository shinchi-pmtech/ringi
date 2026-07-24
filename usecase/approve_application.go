// usecase/approve_application.go
package usecase

import "github.com/shinchi-pmtech/ringi/domain/application"

type ApproveApplication struct {
	repo application.Repository // interface に依存。実装は知らない
}

func NewApproveApplication(repo application.Repository) *ApproveApplication {
	return &ApproveApplication{repo: repo}
}

func (u *ApproveApplication) Execute(id application.ApplicationID, approverID application.ApproverID) error {
	app, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := app.Approve(approverID); err != nil {
		return err
	}
	return u.repo.Save(app)
}
