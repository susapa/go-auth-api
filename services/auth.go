package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/user/go-auth-api/config"
	"github.com/user/go-auth-api/models"
	"github.com/user/go-auth-api/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

func Register(db *sql.DB, req models.RegisterRequest) (*models.AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	user, err := repository.CreateUser(db, req.Name, req.Email, string(hashed))
	if err != nil {
		return nil, err
	}

	return issueTokenPair(db, user)
}

func Login(db *sql.DB, req models.LoginRequest) (*models.AuthResponse, error) {
	user, err := repository.FindByEmail(db, req.Email)
	if err == repository.ErrNotFound {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return issueTokenPair(db, user)
}

func Me(db *sql.DB, userID int64) (*models.User, error) {
	return repository.FindByID(db, userID)
}

func Refresh(db *sql.DB, refreshToken string) (*models.AuthResponse, error) {
	userID, err := repository.FindRefreshToken(db, refreshToken)
	if err != nil {
		return nil, err
	}

	if err := repository.DeleteRefreshToken(db, refreshToken); err != nil {
		return nil, err
	}

	user, err := repository.FindByID(db, userID)
	if err != nil {
		return nil, err
	}

	return issueTokenPair(db, user)
}

func issueTokenPair(db *sql.DB, user *models.User) (*models.AuthResponse, error) {
	accessToken, err := generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(config.C.RefreshTokenExpiryDays) * 24 * time.Hour)
	if err := repository.SaveRefreshToken(db, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func generateAccessToken(user *models.User) (string, error) {
	claims := models.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.C.JWTExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.JWTSecret))
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
