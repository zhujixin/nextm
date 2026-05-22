package model

import "time"

type Account struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	AvatarURL     string `json:"avatarUrl"`
	AuthProvider  string `json:"authProvider"`
	PasswordHash  string `json:"-"`
	MFaEnabled    bool   `json:"mfaEnabled"`
	Locale        string `json:"locale"`
	Timezone      string `json:"timezone"`
	IsActive      bool   `json:"isActive"`
	LastLoginAt   *int64 `json:"lastLoginAt"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

func (a *Account) IsZero() bool {
	return a.ID == ""
}

type Space struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	AccountID    string `json:"accountId"`
	Icon         string `json:"icon"`
	Description  string `json:"description"`
	Encrypted    bool   `json:"encrypted"`
	ObjectCount  int    `json:"objectCount"`
	SyncStatus   string `json:"syncStatus"`
	IsDeleted    bool   `json:"isDeleted"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

func NowMS() int64 {
	return time.Now().UnixMilli()
}

type RefreshToken struct {
	ID         string `json:"id"`
	AccountID  string `json:"accountId"`
	TokenHash  string `json:"-"`
	DeviceID   string `json:"deviceId"`
	ExpiresAt  int64  `json:"expiresAt"`
	Revoked    bool   `json:"revoked"`
	CreatedAt  int64  `json:"createdAt"`
}
