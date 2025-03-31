package init

import "github.com/hewo233/house-system-backend/db"

func DBInit() {
	db.ConnectDB()
	db.UpdateDB()
}

func AllInit() {
	DBInit()
}
