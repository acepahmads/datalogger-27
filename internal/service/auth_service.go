package service

import (
	"errors"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/model"
	"datalogger/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   *repository.SystemRepository
	config *config.Config
}

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(repo *repository.SystemRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, config: cfg}
}

func (s *AuthService) Login(username, password, ipAddress, userAgent string) (string, *model.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return "", nil, errors.New("invalid username or password")
	}

	if !user.IsActive {
		return "", nil, errors.New("user account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid username or password")
	}

	// Update last login
	_ = s.repo.UpdateUserLastLogin(user.ID)

	// Log audit trail
	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &user.ID,
		Username:  user.Username,
		Action:    "LOGIN",
		Resource:  "AUTH",
		Details:   "User successfully logged in",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	})

	roleName := "Viewer"
	if user.Role != nil {
		roleName = user.Role.Name
	}

	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.JWTExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "datalogger-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}
