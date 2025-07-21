package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

var logger *zap.SugaredLogger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	logger = log.Sugar()
}

type Manager struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func New(secret string, access, refresh time.Duration) *Manager {
	return &Manager{
		Secret:     secret,
		AccessTTL:  access,
		RefreshTTL: refresh,
	}
}

func (m *Manager) GenerateAccess(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(m.AccessTTL).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(m.Secret))
	if err != nil {
		logger.Errorw("failed to sign access token", "userID", userID, "error", err)
	}
	return signed, err
}

func (m *Manager) GenerateRefresh() (tokenStr string, id uuid.UUID, err error) {
	id = uuid.New()
	claims := jwt.MapClaims{
		"jti": id.String(),
		"exp": time.Now().Add(m.RefreshTTL).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err = token.SignedString([]byte(m.Secret))
	if err != nil {
		logger.Errorw("failed to sign refresh token", "jti", id.String(), "error", err)
	}
	return
}

func (m *Manager) ParseRefresh(refreshToken string) (string, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.Secret), nil
	})
	if err != nil || !token.Valid {
		logger.Errorw("invalid refresh token", "token", refreshToken, "error", err)
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logger.Errorw("invalid claims type in refresh token", "token", refreshToken)
		return "", errors.New("invalid claims type")
	}
	jti, ok := claims["jti"].(string)
	if !ok {
		logger.Errorw("jti not found in refresh token claims", "claims", claims)
		return "", errors.New("jti not found")
	}
	return jti, nil
}

func (m *Manager) ParseAccess(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.Secret), nil
	})
	if err != nil || !token.Valid {
		logger.Errorw("invalid access token", "token", tokenStr, "error", err)
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logger.Errorw("invalid claims type in access token", "token", tokenStr)
		return 0, errors.New("invalid claims")
	}

	uidFloat, ok := claims["sub"].(float64)
	if !ok {
		logger.Errorw("user_id not found in access token claims", "claims", claims)
		return 0, errors.New("user_id not found in token")
	}
	return int64(uidFloat), nil
}
