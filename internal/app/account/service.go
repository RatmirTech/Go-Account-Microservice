package account

import (
	"context"
	"errors"

	"AccountService/internal/db"
	"AccountService/internal/token"
	"go.uber.org/zap"
)

type Service struct {
	DB     *db.Postgres
	Tokens *token.Manager
}

func NewService(db *db.Postgres, tokens *token.Manager) *Service {
	return &Service{DB: db, Tokens: tokens}
}

func init() {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	logger = log.Sugar()
}

func (s *Service) Register(ctx context.Context, email, password string) error {
	hashed, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.DB.CreateUser(ctx, email, hashed)
}

func (s *Service) Login(ctx context.Context, email, password, deviceID, ip, ua string) (*TokenResponse, error) {
	user, err := s.DB.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Errorw("user not found", "email", email, "error", err)
		return nil, errors.New("user not found")
	}
	if !CheckPasswordHash(password, user.PasswordHash) {
		logger.Warnw("invalid password", "email", email)
		return nil, errors.New("invalid password")
	}

	accessToken, err := s.Tokens.GenerateAccess(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshID, err := s.Tokens.GenerateRefresh()
	if err != nil {
		return nil, err
	}

	err = s.DB.SaveRefreshToken(
		ctx,
		refreshID, user.ID,
		refreshToken,
		deviceID, ip, ua,
		s.Tokens.RefreshTTL,
	)
	if err != nil {
		return nil, err
	}

	logger.Infow("user logged in", "userID", user.ID, "email", email, "ip", ip, "ua", ua)
	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken, deviceID string) (*TokenResponse, error) {
	logger.Infow("refresh token attempt", "deviceID", deviceID)

	jti, err := s.Tokens.ParseRefresh(refreshToken)
	if err != nil {
		logger.Errorw("failed to parse refresh token", "error", err)
		return nil, errors.New("invalid refresh token")
	}

	logger.Infow("parsed refresh token", "jti", jti)

	tokenRecord, err := s.DB.GetRefreshTokenByID(ctx, jti)
	if err != nil {
		logger.Errorw("failed to get refresh token from DB", "jti", jti, "error", err)
		return nil, errors.New("refresh token not found or mismatched")
	}

	logger.Infow("found token record", "tokenRecord.Token", tokenRecord.Token[:20]+"...", "tokenRecord.DeviceID", tokenRecord.DeviceID)

	if tokenRecord.Token != refreshToken {
		logger.Errorw("token mismatch", "expected", tokenRecord.Token[:20]+"...", "received", refreshToken[:20]+"...")
		return nil, errors.New("refresh token not found or mismatched")
	}

	if tokenRecord.DeviceID != deviceID {
		logger.Errorw("device ID mismatch", "expected", tokenRecord.DeviceID, "received", deviceID)
		return nil, errors.New("refresh token not found or mismatched")
	}

	accessToken, err := s.Tokens.GenerateAccess(tokenRecord.UserID)
	if err != nil {
		logger.Errorw("failed to generate access token", "error", err)
		return nil, err
	}

	newRefreshToken, newJti, err := s.Tokens.GenerateRefresh()
	if err != nil {
		logger.Errorw("failed to generate new refresh token", "error", err)
		return nil, err
	}

	err = s.DB.UpdateRefreshToken(
		ctx,
		jti,
		newJti.String(),
		newRefreshToken,
		s.Tokens.RefreshTTL,
	)
	if err != nil {
		logger.Errorw("failed to update refresh token in DB", "error", err)
		return nil, err
	}

	logger.Infow("refresh successful", "userID", tokenRecord.UserID)
	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, userID int64, deviceID string) error {
	return s.DB.DeleteRefreshToken(ctx, userID, deviceID)
}

func (s *Service) GetUserByID(ctx context.Context, userID int64) (*db.User, error) {
	user, err := s.DB.GetUserByID(ctx, userID)
	if err != nil {
		logger.Errorw("failed to get user by ID", "userID", userID, "error", err)
		return nil, err
	}
	return user, nil
}
