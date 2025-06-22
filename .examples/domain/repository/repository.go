//go:generate go tool nrdeco -s $GOFILE
package repository

import (
	"context"

	"github.com/miyamo2/nrdeco/examples/domain/model"
)

type UserRepository interface {
	GetUserByID(string) (*model.User, error)
	GetAllUsers() ([]model.User, error)
	GetUserByIDWithContext(context.Context, string) (*model.User, error)
	GetAllUsersWithContext(context.Context) ([]model.User, error)
}
