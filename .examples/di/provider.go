package di

import (
	"net/http"
	"os"

	"github.com/google/wire"
	"github.com/miyamo2/nrdeco/examples/domain/repository"
	"github.com/miyamo2/nrdeco/examples/infra/inmemory"
	"github.com/miyamo2/nrdeco/examples/interfaces"
	"github.com/miyamo2/nrdeco/examples/usecase"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func NewRelic() *newrelic.Application {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("NEW_RELIC_CONFIG_APP_NAME")),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_CONFIG_LICENSE")),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		panic(err)
	}
	return app
}

func NewServeMux(app *newrelic.Application, handler *interfaces.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(newrelic.WrapHandleFunc(app, "/users", handler.GetAllUsers))
	mux.HandleFunc(newrelic.WrapHandleFunc(app, "/users/{id}", handler.GetUserByID))
	return mux
}

func NewUserRepository() *repository.NRUserRepository {
	return &repository.NRUserRepository{
		UserRepository: inmemory.NewUserRepository(),
	}
}

func NewUserUseCase(userRepository repository.UserRepository) *usecase.NRUserUseCase {
	return &usecase.NRUserUseCase{
		UserUseCase: usecase.NewUserUseCase(userRepository),
	}
}

var (
	UserRepositorySet = wire.NewSet(
		NewUserRepository,
		wire.Bind(new(repository.UserRepository), new(*repository.NRUserRepository)))
	UserUseCaseSet = wire.NewSet(
		NewUserUseCase,
		wire.Bind(new(usecase.UserUseCase), new(*usecase.NRUserUseCase)))
	NewRelicSet = wire.NewSet(NewRelic)
	HandlerSet  = wire.NewSet(interfaces.NewHandler)
	ServeMuxSet = wire.NewSet(NewServeMux)
)
