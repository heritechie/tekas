package main

import (
	"log"
	"mnctech-restapi/cmd/rest-api/handlers"
	"mnctech-restapi/cmd/rest-api/middlewares"
	"mnctech-restapi/cmd/rest-api/models"
	"net/http"
	"os"

	"github.com/go-gormigrate/gormigrate/v2"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Connect to the database and perform migrations
	db := ConnectDB()
	accessTokenKey := []byte(os.Getenv("ACCESS_TOKEN_KEY"))   // or replace with your actual key
	refreshTokenKey := []byte(os.Getenv("REFRESH_TOKEN_KEY")) // or replace with your actual key

	// Call the migration function
	if err := MigrateDatabase(db); err != nil {
		log.Fatalf("could not migrate: %v", err)
	}

	// Set up the router using the NewRouter function
	r := NewRouter(db, accessTokenKey, refreshTokenKey)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

var DB *gorm.DB

// ConnectDB establishes a connection to the PostgreSQL database.
func ConnectDB() *gorm.DB {
	var err error

	dsn := "host=db user=mnctech password=mnctechpass dbname=mnctechdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// MigrateDatabase initializes the database migrations
func MigrateDatabase(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "20241020_05",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(
					&models.User{},
					&models.UserAccount{},
					&models.AccountTransactionLog{},
					&models.TopUpTransaction{},
					&models.PaymentTransaction{},
					&models.TransferTransaction{},
				); err != nil {
					log.Printf("Error during migration: %v", err)
					return err
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				// Define rollback logic here
				return tx.Migrator().DropTable("users")
			},
		},
	})

	// Execute migrations
	if err := m.Migrate(); err != nil {
		log.Printf("Error running migrations: %v", err) // Log any errors
		return err
	}
	log.Println("Migrations ran successfully!")
	return nil
}

// NewRouter initializes and returns a new mux.Router with the defined routes.
func NewRouter(db *gorm.DB, accessTokenKey, refreshTokenKey []byte) *mux.Router {
	appHandler := &handlers.AppHandler{
		DB: db,
	}
	authHandler := &handlers.AuthHandler{
		AppHandler:      appHandler,
		AccessTokenKey:  accessTokenKey,
		RefreshTokenKey: refreshTokenKey,
	}

	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")
	r.Handle("/topup", middlewares.JWTMiddleware(accessTokenKey)(http.HandlerFunc(appHandler.HandleTopUp))).Methods("POST")
	r.Handle("/pay", middlewares.JWTMiddleware(accessTokenKey)(http.HandlerFunc(appHandler.HandlePayment))).Methods("POST")
	r.Handle("/transfer", middlewares.JWTMiddleware(accessTokenKey)(http.HandlerFunc(appHandler.HandleTransfer))).Methods("POST")

	return r
}
