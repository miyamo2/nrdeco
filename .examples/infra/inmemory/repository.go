package inmemory

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/miyamo2/nrdeco/examples/domain/model"
	"github.com/miyamo2/nrdeco/examples/domain/repository"
)

var _ repository.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	m map[string]model.User
}

func (u *UserRepository) GetUserByID(s string) (*model.User, error) {
	return u.GetUserByIDWithContext(context.Background(), s)
}

func (u *UserRepository) GetAllUsers() ([]model.User, error) {
	return u.GetAllUsersWithContext(context.Background())
}

func (u *UserRepository) GetUserByIDWithContext(_ context.Context, s string) (*model.User, error) {
	time.Sleep(900 * time.Millisecond)
	user, ok := u.m[s]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (u *UserRepository) GetAllUsersWithContext(_ context.Context) ([]model.User, error) {
	time.Sleep(900 * time.Millisecond)
	users := make([]model.User, 0, len(u.m))

	keys := slices.Collect(maps.Keys(u.m))
	slices.Sort(keys)

	for _, key := range keys {
		users = append(users, u.m[key])
	}
	return users, nil
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		m: map[string]model.User{
			"1": {
				ID:   "1",
				Name: "Bob",
			},
			"2": {
				ID:   "2",
				Name: "Alice",
			},
		},
	}
}
