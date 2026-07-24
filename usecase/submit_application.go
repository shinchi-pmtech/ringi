// usecase/submit_application.go
package usecase

import "github.com/shinchi-pmtech/ringi/domain/application"

type SubmitApplication struct {
	repo application.Repository
}

func NewSubmitApplication(repo application.Repository) *SubmitApplication {
	return &SubmitApplication{repo: repo}
}

func (u *SubmitApplication) Execute(id application.ApplicationID) error {
	app, err := u.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := app.Submit(); err != nil {
		return err
	}
	return u.repo.Save(app)
}
