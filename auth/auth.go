package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/edgetx/cloudbuild/config"
	"github.com/edgetx/cloudbuild/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	accessKeyLen         = 16
	secretKeyLen         = 32
	secretHashCost       = 10
	alphaNumericTable    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphaNumericTableLen = byte(len(alphaNumericTable))
)

var (
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrTokenExpired         = errors.New("token expired")
	ErrBadRandomData        = errors.New("not enough random data")
)

type AuthToken struct {
	AccessKey  string `gorm:"primary_key"`
	SecretKey  string
	User       string `gorm:"index:user_idx"`
	ValidUntil *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (AuthToken) TableName() string {
	return "auth_tokens"
}

func init() {
	database.RegisterModels(&AuthToken{})
}

type AuthTokenDB struct {
	db *gorm.DB
}

func NewAuthTokenDB(db *gorm.DB) *AuthTokenDB {
	return &AuthTokenDB{
		db: db,
	}
}

func NewAuthTokenDBFromConfig(c *config.CloudbuildOpts) (*AuthTokenDB, error) {
	db, err := database.New(c.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	return NewAuthTokenDB(db), nil
}

func (at *AuthTokenDB) CreateToken(user string, validity *time.Duration) (*AuthToken, error) {
	token, err := GenerateAuthToken(user)
	if err != nil {
		return nil, err
	}

	if validity != nil {
		expires := time.Now().Add(*validity)
		token.ValidUntil = &expires
	}

	secretKey := token.SecretKey
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(secretKey), secretHashCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash secret key: %w", err)
	}
	token.SecretKey = string(hashedKey)

	err = at.db.Create(token).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication token: %w", err)
	}

	// write SecretKey back so the user can it read it once.
	token.SecretKey = secretKey

	return token, nil
}

func (at *AuthTokenDB) Authenticate(accessKey, secretKey string) error {
	var token AuthToken
	err := at.db.Take(&token, "access_key = ?", accessKey).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// let's use some fake hashed key ("00000000")
		token = AuthToken{
			AccessKey: accessKey,
			SecretKey: "$2a$10$9z5e/ds1HHk1CZLRb59ok.pihaR47T/IK4gYb/2q08X20mpKaR1Oe",
		}
	} else if err != nil {
		return err
	}

	secretKeyByte := []byte(secretKey)
	hashedKeyBytes := []byte(token.SecretKey)

	err = bcrypt.CompareHashAndPassword(hashedKeyBytes, secretKeyByte)

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrAuthenticationFailed
	}

	if err != nil {
		return err
	}

	if token.ValidUntil != nil && !token.ValidUntil.After(time.Now()) {
		// token expired
		return ErrTokenExpired
	}
	return nil
}

func (at *AuthTokenDB) ListTokens() (*[]AuthToken, error) {
	var tokens []AuthToken
	err := at.db.Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

func (at *AuthTokenDB) RemoveToken(accessKey string) error {
	return at.db.Delete(&AuthToken{
		AccessKey: accessKey,
	}).Error
}

func generateCredentials() (accessKey, secretKey string, err error) {
	readBytes := func(size int) (data []byte, err error) {
		data = make([]byte, size)
		var n int
		if n, err = rand.Read(data); err != nil {
			return nil, err
		} else if n != size {
			return nil, ErrBadRandomData
		}
		return data, nil
	}

	// Generate access key.
	keyBytes, err := readBytes(accessKeyLen)
	if err != nil {
		return "", "", err
	}
	for i := 0; i < accessKeyLen; i++ {
		keyBytes[i] = alphaNumericTable[keyBytes[i]%alphaNumericTableLen]
	}
	accessKey = string(keyBytes)

	// Generate secret key.
	keyBytes, err = readBytes(secretKeyLen)
	if err != nil {
		return "", "", err
	}

	secretKey = string([]byte(base64.StdEncoding.EncodeToString(keyBytes))[:secretKeyLen])
	return accessKey, secretKey, nil
}

func GenerateAuthToken(user string) (*AuthToken, error) {
	accessKey, secretKey, err := generateCredentials()
	if err != nil {
		return nil, err
	}
	return &AuthToken{
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		User:       user,
		ValidUntil: nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}
