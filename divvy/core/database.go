package core

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB obj
var DB *gorm.DB

func ConnectDB() {
	// db connection with gorm
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	// "user:pass@tcp(127.0.0.1:5432)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// sslmode := os.Getenv("SSLMODE")

	// username := os.Getenv("USERNAME")
	// password := os.Getenv("PASSWORD")
	// host := os.Getenv("HOST")
	// port := os.Getenv("PORT")
	// database := os.Getenv("DATABASE")
	// dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"

	dsn := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	// db.LogMode(true)
	if err != nil {
		fmt.Println("DB ERROR:")
		panic(err)
	}

	fmt.Println("CONNECTED TO DB!")

	DB = db
}
