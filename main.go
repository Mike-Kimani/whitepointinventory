package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/mike-kimani/whitepointinventory/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	fmt.Println("Welcome to White Point inventory")

	godotenv.Load(".env")

	portstring := os.Getenv("PORT")

	if portstring == "" {
		log.Fatal("couldn't find a port in this environment")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL not found in this environment")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error: ", err)
		log.Fatal("Cannot connect to database")
	}

	db := database.New(conn)

	apiCfg := apiConfig{
		DB: db,
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	}))

	v1Router := chi.NewRouter()

	v1Router.Get("/healthz", handlerHealth)
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.handlerGetUserFromCookie)
	v1Router.Get("/user", apiCfg.handlerGetUsers)
	v1Router.Post("/login", apiCfg.handlerUserLogin)
	v1Router.Post("/logout", apiCfg.handlerUserLogout)
	v1Router.Post("/farmers", apiCfg.handlerCreateFarmer)
	v1Router.Get("/farmers", apiCfg.handlerGetFarmerByName)
	v1Router.Get("/farmer", apiCfg.handlerGetFarmers)
	v1Router.Delete("/farmers/{farmer_id}", apiCfg.handlerDeleteFarmer)
	v1Router.Post("/purchases", apiCfg.middlewareAuth(apiCfg.handerCreatePurchases))
	v1Router.Get("/purchases", apiCfg.handlerGetPurchases)
	v1Router.Get("/purchase", apiCfg.handlerGetPurchaseByID)
	v1Router.Delete("/purchases/{purchase_id}", apiCfg.middlewareAuth(apiCfg.handlerDeletePurchase))
	v1Router.Post("/payments", apiCfg.middlewareAuth(apiCfg.handlerCreatePayment))
	v1Router.Get("/payment", apiCfg.handlerGetPaymentByID)
	v1Router.Get("/payments", apiCfg.handlerGetPayments)
	v1Router.Delete("/payments/{payment_id}", apiCfg.middlewareAuth(apiCfg.handlerDeletePayment))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portstring,
	}
	log.Printf("Server starting on port %v", portstring)

	err = srv.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}
