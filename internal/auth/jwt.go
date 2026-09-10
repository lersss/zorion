// internal/auth/jwt.go
package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ==================== СЕКРЕТ ====================
//
// Секрет задаётся один раз при старте сервера через InitJWTSecret
// (значение читается из переменной окружения JWT_SECRET).
// До инициализации GenerateToken и VerifyToken возвращают ошибку —
// это защита от случайного запуска без секрета.

// MinJWTSecretLength — минимальная длина секрета в байтах.
// 32 байта (256 бит) — рекомендованная длина ключа для HMAC-SHA256.
const MinJWTSecretLength = 32

var (
	jwtSecretMu    sync.RWMutex
	jwtSecret      []byte
	jwtSecretReady bool
)

// InitJWTSecret устанавливает секрет для подписи JWT.
// Вызывается из main.go при старте, до запуска HTTP-сервера.
//
// Возвращает ошибку, если:
//   - секрет короче MinJWTSecretLength;
//   - секрет уже был установлен ранее.
func InitJWTSecret(secret string) error {
	if len(secret) < MinJWTSecretLength {
		return fmt.Errorf(
			"JWT-секрет слишком короткий: %d байт, нужно минимум %d",
			len(secret), MinJWTSecretLength,
		)
	}

	jwtSecretMu.Lock()
	defer jwtSecretMu.Unlock()

	if jwtSecretReady {
		return errors.New("JWT-секрет уже инициализирован")
	}

	jwtSecret = []byte(secret)
	jwtSecretReady = true
	return nil
}

// getJWTSecret — безопасное чтение секрета из горутин.
// Возвращает ошибку, если InitJWTSecret не вызывался.
func getJWTSecret() ([]byte, error) {
	jwtSecretMu.RLock()
	defer jwtSecretMu.RUnlock()

	if !jwtSecretReady {
		return nil, errors.New("JWT-секрет не инициализирован")
	}
	return jwtSecret, nil
}

// ==================== ТОКЕНЫ ====================

// Claims — полезная нагрузка JWT.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken создаёт JWT для пользователя.
// TTL — 24 часа.
func GenerateToken(userID string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// VerifyToken проверяет JWT и возвращает userID.
func VerifyToken(tokenString string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Явная проверка алгоритма — защита от подмены на "none" или RS256.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}
	return claims.UserID, nil
}