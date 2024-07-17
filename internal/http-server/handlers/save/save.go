package save

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	storage "github.com/maestro-milagro/User_Service_PB/internal"
	"github.com/maestro-milagro/User_Service_PB/internal/lib/sl"
	"github.com/maestro-milagro/User_Service_PB/internal/models"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	models.User
}

type Response struct {
	models.Response
}

type UserSaver interface {
	SaveUser(ctx context.Context, user models.User) (int64, error)
}

func New(log *slog.Logger, userSaver UserSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		//var req Request
		var req models.User

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("empty request"))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		log.Info("registering user")

		passHash, err := bcrypt.GenerateFromPassword([]byte(req.PassHash), bcrypt.DefaultCost)

		req.PassHash = string(passHash)

		if err != nil {
			log.Error("failed to generate password hash", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to generate password hash"))

			return
		}

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.ValidationError(validateErr))

			return
		}
		bruh := string(req.PassHash)
		fmt.Println(bruh)
		id, err := userSaver.SaveUser(context.Background(), req)
		if errors.Is(err, storage.ErrUserExist) {
			log.Info("user already exists", slog.String("user", req.Email))

			render.Status(r, http.StatusBadRequest)

			render.JSON(w, r, models.Error("user already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add user", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)

			render.JSON(w, r, models.Error("failed to add user"))

			return
		}

		log.Info("user added", slog.Int64("id", id))

		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: models.OK(),
	})
}
