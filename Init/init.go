package Init

import (
	"github.com/hewo233/house-system-backend/db"
	"github.com/hewo233/house-system-backend/utils/OSS"
	"github.com/hewo233/house-system-backend/utils/jwt"
)

func AllInit() {
	db.Init()
	OSS.Init()
	jwt.InitJWTKey()
}
