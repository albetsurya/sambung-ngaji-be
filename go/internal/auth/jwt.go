package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"pengajian-backend/internal/model"
)

var ErrInvalidToken = errors.New("token tidak valid")

type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTManager(secret string, ttlHours int) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    time.Duration(ttlHours) * time.Hour,
	}
}

func (m *JWTManager) TTL() time.Duration { return m.ttl }

func (m *JWTManager) Generate(sessionID, userID, role string, groupID ...string) (string, error) {
	claims := jwt.MapClaims{
		"sid":  sessionID,
		"uid":  userID,
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(m.ttl).Unix(),
	}
	if len(groupID) > 0 && groupID[0] != "" {
		claims["gid"] = groupID[0]
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenStr string) (*model.SessionClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	sid, _ := claims["sid"].(string)
	uid, _ := claims["uid"].(string)
	role, _ := claims["role"].(string)
	var gid *string
	if g, ok := claims["gid"].(string); ok && g != "" {
		gid = &g
	}
	if sid == "" || uid == "" {
		return nil, ErrInvalidToken
	}
	return &model.SessionClaims{SessionID: sid, UserID: uid, Role: role, GroupID: gid}, nil
}
