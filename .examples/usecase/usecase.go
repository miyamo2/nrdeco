//go:generate go tool nrdeco -s $GOFILE
package usecase

import (
	"context"

	"github.com/miyamo2/nrdeco/examples/domain/repository"
)

type UserDto struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserUseCase interface {
	GetUserByID(string) (*UserDto, error)
	GetAllUsers() ([]UserDto, error)
	GetUserByIDWithContext(context.Context, string) (*UserDto, error)
	GetAllUsersWithContext(context.Context) ([]UserDto, error)
}

type userUseCase struct {
	repo repository.UserRepository
}

func (u *userUseCase) GetAllUsersWithContext(ctx context.Context) ([]UserDto, error) {
	users, err := u.repo.GetAllUsersWithContext(ctx)
	if err != nil {
		return nil, err
	}
	userDTOs := make([]UserDto, 0, len(users))
	for _, user := range users {
		userDTOs = append(userDTOs, UserDto{
			ID:   user.ID,
			Name: user.Name,
		})
	}
	return userDTOs, nil
}

func (u *userUseCase) GetUserByIDWithContext(ctx context.Context, id string) (*UserDto, error) {
	user, err := u.repo.GetUserByIDWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserDto{
		ID:   user.ID,
		Name: user.Name,
	}, nil
}

func (u *userUseCase) GetUserByID(s string) (*UserDto, error) {
	return u.GetUserByIDWithContext(context.Background(), s)
}

func (u *userUseCase) GetAllUsers() ([]UserDto, error) {
	return u.GetAllUsersWithContext(context.Background())
}

func NewUserUseCase(repo repository.UserRepository) UserUseCase {
	return &userUseCase{
		repo: repo,
	}
}
