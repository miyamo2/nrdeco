//go:build wireinject

package di

import (
	"net/http"

	"github.com/google/wire"
)

func ServeMux() *http.ServeMux {
	wire.Build(
		UserRepositorySet,
		UserUseCaseSet,
		HandlerSet,
		NewRelicSet,
		ServeMuxSet)
	return nil
}
