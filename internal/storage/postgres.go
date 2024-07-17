package postgres

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"github.com/maestro-milagro/User_Service_PB/internal/models"
)

type Storage struct {
	db *sqlx.DB
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func New(cfg Config) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, user models.User) (int64, error) {
	const op = "Storage/postgres/SaveUser"
	//stmt, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES(?, ?)")
	//if err != nil {
	//	return 0, fmt.Errorf("%s: %w", op, err)
	//}
	//
	//res, err := stmt.ExecContext(ctx, user.Email, user.PassHash)
	//if err != nil {
	//	// var sqliteErr sqlite3.Error
	//	// if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
	//	// 	return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExist)
	//	// }
	//	return 0, fmt.Errorf("%s: %w", op, err)
	//}
	//
	//id, err := res.LastInsertId()
	//if err != nil {
	//	return 0, fmt.Errorf("%s: %w", op, err)
	//}
	//
	//return id, nil

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int
	createListQuery := fmt.Sprintf("INSERT INTO users (email, pass_hash) VALUES ($1, $2) RETURNING id")
	row := tx.QueryRow(createListQuery, user.Email, string(user.PassHash))
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return int64(id), tx.Commit()
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "Storage/postgres/User"
	//
	//stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = ? ")
	//
	//if err != nil {
	//	return models.User{}, fmt.Errorf("%s: %w", op, err)
	//}
	//
	//row := stmt.QueryRowContext(ctx, email)
	//
	//var user models.User
	//
	//err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	//if err != nil {
	//	if errors.Is(err, sql.ErrNoRows) {
	//		return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	//	}
	//
	//	return models.User{}, fmt.Errorf("%s: %w", op, err)
	//}
	//return user, nil

	var user models.User

	createListQuery := fmt.Sprintf("SELECT * FROM users WHERE email = $1")

	row := s.db.QueryRow(createListQuery, email)

	err := row.Scan(&user.ID, &user.Email, &user.PassHash)

	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

// func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
// 	const op = "storage.sqlite.IsAdmin"

// 	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
// 	if err != nil {
// 		return false, fmt.Errorf("%s: %w", op, err)
// 	}

// 	row := stmt.QueryRowContext(ctx, userID)

// 	var isAdmin bool

// 	err = row.Scan(&isAdmin)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
// 		}

// 		return false, fmt.Errorf("%s: %w", op, err)
// 	}

// 	return isAdmin, nil
// }
