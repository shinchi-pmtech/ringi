// usecase/resubmit_application.go
package usecase

import "github.com/shinchi-pmtech/ringi/domain/application"

type ResubmitApplication struct {
	repo application.Repository
}

func NewResubmitApplication(repo application.Repository) *ResubmitApplication {
	return &ResubmitApplication{repo: repo}
}

func (u *ResubmitApplication) Execute(id application.ApplicationID) error {
	app, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := app.Resubmit(); err != nil {
		return err
	}
	return u.repo.Save(app)
}
