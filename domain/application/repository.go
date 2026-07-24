// domain/application/repository.go
package application

type Repository interface {
	Save(app *Application) error
	FindByID(id ApplicationID) (*Application, error)
}
