package consts

import "time"

const (
	Issuer = "house-system-backend"
	User   = "user"
	Admin  = "admin"

	OneDay    = 24 * time.Hour
	ThreeDays = 3 * OneDay

	MB     = 1024 * 1024
	TreeMB = 3 * MB
)
