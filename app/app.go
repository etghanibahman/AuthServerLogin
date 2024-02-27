package app

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bankingAuth/domain"
	"bankingAuth/logger"
	"bankingAuth/service"
	"bankingAuth/storage"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func Start() {
	sanityCheck()
	router := mux.NewRouter()
	authRepository := domain.NewAuthRepository(getDbClient())
	ah := AuthHandler{service.NewLoginService(authRepository, domain.GetRolePermissions())}

	router.HandleFunc("/auth/login", ah.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/register", ah.NotImplementedHandler).Methods(http.MethodPost)
	router.HandleFunc("/auth/refresh", ah.Refresh).Methods(http.MethodPost)
	router.HandleFunc("/auth/verify", ah.Verify).Methods(http.MethodGet)

	address := os.Getenv("SERVER_ADDRESS")
	port := os.Getenv("SERVER_PORT")
	logger.Info(fmt.Sprintf("Starting OAuth server on %s:%s ...", address, port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), router))
}

func getDbClient() *gorm.DB {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Fatal(err.Error())
	}
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	client, err := storage.NewConnection(config)
	if err != nil {
		logger.Fatal("could not load the database")
	}
	return client
}

// func getDbClient() *sqlx.DB {
// 	dbUser := os.Getenv("DB_USER")
// 	dbPasswd := os.Getenv("DB_PASSWD")
// 	dbAddr := os.Getenv("DB_ADDR")
// 	dbPort := os.Getenv("DB_PORT")
// 	dbName := os.Getenv("DB_NAME")

// 	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPasswd, dbAddr, dbPort, dbName)
// 	client, err := sqlx.Open("mysql", dataSource)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// See "Important settings" section.
// 	client.SetConnMaxLifetime(time.Minute * 3)
// 	client.SetMaxOpenConns(10)
// 	client.SetMaxIdleConns(10)
// 	return client
// }

func sanityCheck() {
	envProps := []string{
		"SERVER_ADDRESS",
		"SERVER_PORT",
		"DB_USER",
		"DB_PASSWD",
		"DB_ADDR",
		"DB_PORT",
		"DB_NAME",
	}
	for _, k := range envProps {
		if os.Getenv(k) == "" {
			logger.Error(fmt.Sprintf("Environment variable %s not defined. Terminating application...", k))
		}
	}
}
