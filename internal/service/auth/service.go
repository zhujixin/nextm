package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/model"
	"github.com/nextm/nextm/internal/pkg/crypto"
)

type Repository interface {
	CreateAccount(ctx context.Context, arg interface{}) (interface{}, error)
	GetAccountByEmail(ctx context.Context, email string) (interface{}, error)
	GetAccountByID(ctx context.Context, id string) (interface{}, error)
	UpdateLastLogin(ctx context.Context, accountID string, lastLoginAt, updatedAt int64) error
	CreateRefreshToken(ctx context.Context, arg interface{}) error
	GetRefreshToken(ctx context.Context, id string, now int64) (interface{}, error)
	RevokeRefreshToken(ctx context.Context, id string) error
	RevokeAllAccountTokens(ctx context.Context, accountID string) error
	CreateSpace(ctx context.Context, arg interface{}) (interface{}, error)
	GetPersonalSpace(ctx context.Context, accountID string) (interface{}, error)
	CreateDefaultTypes(ctx context.Context, arg interface{}) error
}

type Config struct {
	BcryptCost      int
	RateLimit       int
	RateLimitWindow time.Duration
}

type Service struct {
	repo    Repository
	jwt     *crypto.JWTManager
	cfg     Config
}

func NewService(repo Repository, jwt *crypto.JWTManager, cfg Config) *Service {
	return &Service{repo: repo, jwt: jwt, cfg: cfg}
}

func (s *Service) Register(ctx context.Context, req dto.RegisterRequest) (*dto.LoginResponse, error) {
	// 检查邮箱是否已注册
	existing, _ := s.repo.GetAccountByEmail(ctx, req.Email)
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// 哈希密码
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := model.NowMS()
	accountID := uuid.New().String()

	// 创建账号
	_, err = s.repo.CreateAccount(ctx, struct {
		ID           string
		Email        string
		Name         string
		PasswordHash string
		Locale       string
		Timezone     string
		CreatedAt    int64
		UpdatedAt    int64
	}{
		ID: accountID, Email: req.Email, Name: req.Name,
		PasswordHash: passwordHash, Locale: "zh-CN",
		Timezone: "Asia/Shanghai", CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	// 创建个人 Space
	spaceID := uuid.New().String()
	_, err = s.repo.CreateSpace(ctx, struct {
		ID        string
		Name      string
		Type      string
		AccountID string
		CreatedAt int64
		UpdatedAt int64
	}{
		ID: spaceID, Name: req.Name + " 的工作区",
		Type: "personal", AccountID: accountID,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create space: %w", err)
	}

	// 创建默认对象类型
	s.createDefaultTypes(ctx, spaceID, now)

	// 生成 JWT
	accessToken, expiresIn, err := s.jwt.GenerateAccessToken(accountID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// 生成 Refresh Token
	refreshToken, err := s.generateAndStoreRefreshToken(ctx, accountID, now)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Account: dto.AccountDTO{
			ID: accountID, Email: req.Email, Name: req.Name,
		},
		Space: dto.SpaceDTO{
			ID: spaceID, Name: req.Name + " 的工作区",
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// 查找账号
	acct, err := s.repo.GetAccountByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("get account: %w", err)
	}

	account := acct.(*model.Account)

	// 验证密码
	if err := crypto.CheckPassword(req.Password, account.PasswordHash); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 更新最后登录时间
	now := model.NowMS()
	s.repo.UpdateLastLogin(ctx, account.ID, now, now)

	// 获取个人 Space
	sp, _ := s.repo.GetPersonalSpace(ctx, account.ID)

	// 生成 JWT
	accessToken, expiresIn, err := s.jwt.GenerateAccessToken(account.ID, account.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, account.ID, now)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	resp := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Account: dto.AccountDTO{
			ID: account.ID, Email: account.Email,
			Name: account.Name, AvatarURL: account.AvatarURL,
			Locale: account.Locale,
		},
	}

	if sp != nil {
		space := sp.(*model.Space)
		resp.Space = dto.SpaceDTO{
			ID: space.ID, Name: space.Name,
			Type: space.Type, Icon: space.Icon,
			Description: space.Description, ObjectCount: space.ObjectCount,
		}
	}

	return resp, nil
}

func (s *Service) RefreshToken(ctx context.Context, req dto.RefreshRequest) (*dto.RefreshResponse, error) {
	now := model.NowMS()

	// 查找 refresh token
	token, err := s.repo.GetRefreshToken(ctx, req.RefreshToken, now)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	rt := token.(*model.RefreshToken)

	// 废弃旧的 refresh token
	s.repo.RevokeRefreshToken(ctx, rt.ID)

	// 获取账号
	acct, err := s.repo.GetAccountByID(ctx, rt.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	account := acct.(*model.Account)

	// 生成新的 token 对
	accessToken, expiresIn, err := s.jwt.GenerateAccessToken(account.ID, account.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, account.ID, now)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RevokeAllTokens 撤销账号所有 refresh token（登出）
func (s *Service) RevokeAllTokens(ctx context.Context, accountID string) error {
	return s.repo.RevokeAllAccountTokens(ctx, accountID)
}

func (s *Service) generateAndStoreRefreshToken(ctx context.Context, accountID string, now int64) (string, error) {
	tokenStr, err := crypto.RandomToken(32)
	if err != nil {
		return "", err
	}

	tokenHash := crypto.HashToken(tokenStr)
	tokenID := uuid.New().String()
	expiresAt := now + s.jwt.AccessTokenTTL()*6 // refresh token 有效期更长

	err = s.repo.CreateRefreshToken(ctx, struct {
		ID        string
		AccountID string
		TokenHash string
		ExpiresAt int64
		CreatedAt int64
	}{
		ID: tokenID, AccountID: accountID,
		TokenHash: tokenHash, ExpiresAt: expiresAt,
		CreatedAt: now,
	})
	if err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return tokenID, nil
}

func (s *Service) createDefaultTypes(ctx context.Context, spaceID string, now int64) {
	defaults := []struct {
		Name        string
		Icon        string
		Description string
	}{
		{"笔记", "📝", "通用笔记"},
		{"文章", "📄", "网页/文章"},
		{"书籍", "📚", "读书笔记"},
		{"项目", "📋", "项目管理"},
		{"会议", "🎯", "会议记录"},
		{"联系人", "👤", "人脉管理"},
		{"代码", "💻", "代码片段"},
	}

	for _, d := range defaults {
		s.repo.CreateDefaultTypes(ctx, struct {
			ID          string
			SpaceID     string
			Name        string
			Icon        string
			Description string
			IsBuiltin   int
			CreatedAt   int64
			UpdatedAt   int64
		}{
			ID: uuid.New().String(), SpaceID: spaceID,
			Name: d.Name, Icon: d.Icon, Description: d.Description,
			IsBuiltin: 1, CreatedAt: now, UpdatedAt: now,
		})
	}
}
