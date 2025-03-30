package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/hewo233/house-system-backend/shared/consts"
	"time"
)

//TODO: Delete this and use config.yaml

var JWTKey = []byte("iwoqdoiajsd")

type Claims struct {
	jwt.StandardClaims
}

func GenerateJWT(phone string, audience string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(consts.ThreeDays)

	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Audience:  audience,
			IssuedAt:  nowTime.Unix(),
			Issuer:    consts.Issuer,
			Id:        phone,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(JWTKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}
