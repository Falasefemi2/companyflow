package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

type AuthHandler struct {
	authService services.IAuthService
}

func NewAuthHandler(authService services.IAuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/auth/login", h.Login).Methods(http.MethodPost)
}

// Login godoc
// @Summary Login
// @Description Authenticate user and return JWT token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} utils.APIResponse{data=dto.LoginResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	loginResponse, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    loginResponse,
	})
}
