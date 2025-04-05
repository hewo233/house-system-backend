package db

import (
	"fmt"
	"github.com/hewo233/house-system-backend/models"
	"github.com/hewo233/house-system-backend/shared/consts"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func UpdateDB() {
	err := DB.Table(consts.UserTable).AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal(err)
	}
	err = DB.Table(consts.PropertyTable).AutoMigrate(&models.Property{})
	if err != nil {
		log.Fatal(err)
	}
	err = DB.Table(consts.PropertyImageTable).AutoMigrate(&models.PropertyImage{})
	log.Println("\033[32mAutoMigrate success\033[0m")
}

func ConnectDB() {

	if err := godotenv.Load(consts.DBEnvFile); err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", host, user, password, dbname, port)

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}
}

