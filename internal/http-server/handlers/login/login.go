package login

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	storage "github.com/maestro-milagro/User_Service_PB/internal"
	"github.com/maestro-milagro/User_Service_PB/internal/lib/jwt"
	"github.com/maestro-milagro/User_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/User_Service_PB/internal/models"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"time"
)

type Request struct {
	email    string
	passHash []byte
}

type Response struct {
	Token string `json:"token"`
	models.Response
}

type UserGetter interface {
	User(ctx context.Context, email string) (models.User, error)
}

func New(log *slog.Logger, userGetter UserGetter, secret string, tokenTTL time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.login.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		//		var req Request

		email := chi.URLParam(r, "email")
		if email == "" {
			log.Info("email is empty")

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("invalid request"))

			return
		}

		passHash := chi.URLParam(r, "pass_hash")
		if passHash == "" {
			log.Info("password is empty")

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("invalid request"))

			return
		}

		user, err := userGetter.User(r.Context(), email)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Warn("user not found", sl.Err(err))

				render.Status(r, http.StatusNotFound)

				render.JSON(w, r, models.Error("invalid credentials"))

				return
			}
			log.Error("failed to get user", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to get user"))

			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(passHash)); err != nil {
			log.Info("invalid credentials", sl.Err(err))

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("invalid credentials"))

			return
		}

		token, err := jwt.NewToken(user, secret, tokenTTL)
		if err != nil {
			log.Error("failed to get user", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to get user"))

			return
		}

		render.JSON(w, r, Response{
			Token:    token,
			Response: models.OK(),
		})
	}
}
