package interfaces

import (
	"encoding/json"
	"net/http"

	"github.com/miyamo2/nrdeco/examples/usecase"
)

type Handler struct {
	userUseCase usecase.UserUseCase
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.userUseCase.GetUserByIDWithContext(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}
	v, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(v)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	user, err := h.userUseCase.GetAllUsersWithContext(r.Context())
	if err != nil {
		http.Error(w, "failed to get users", http.StatusInternalServerError)
		return
	}
	v, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "failed to get users", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(v)
}

func NewHandler(userUseCase usecase.UserUseCase) *Handler {
	return &Handler{
		userUseCase: userUseCase,
	}
}
