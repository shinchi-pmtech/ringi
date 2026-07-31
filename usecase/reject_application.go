// usecase/reject_application.go
package usecase

import "github.com/shinchi-pmtech/ringi/domain/application"

type RejectApplication struct {
	repo application.Repository
}

func NewRejectApplication(repo application.Repository) *RejectApplication {
	return &RejectApplication{repo: repo}
}

func (u *RejectApplication) Execute(id application.ApplicationID, approverID application.ApproverID) error {
	app, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := app.Reject(approverID); err != nil {
		return err
	}
	return u.repo.Save(app)
}
