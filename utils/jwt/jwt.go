package jwt

import (
	"bufio"
	"github.com/dgrijalva/jwt-go"
	"github.com/hewo233/house-system-backend/shared/consts"
	"log"
	"os"
	"strings"
	"time"
)

//TODO: Delete this and use config.yaml

var JWTKey []byte

func InitJWTKey() {
	file, err := os.Open(consts.JWTKeyFile)
	if err != nil {
		log.Fatal("failed to open JWT key file: " + err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var jwtKeyString string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "JWTKEY=") {
			jwtKeyString = strings.TrimPrefix(line, "JWTKey=")
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("failed to read JWT key file: " + err.Error())
	}

	if jwtKeyString == "" {
		log.Fatal("JWT key in system is empty")
	}

	JWTKey = []byte(jwtKeyString)
}

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
