package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	handler "github.com/RINOHeinrich/gocidial/Handler"
	"github.com/joho/godotenv"
)

func main() {
	// Charger les variables d'environnement à partir du fichier .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Erreur lors du chargement du fichier .env: %v", err)
	}

	// Récupérer les paramètres de configuration
	tlsEnable := os.Getenv("TLS_ENABLE")
	tlsCert := os.Getenv("TLS_CERT")
	tlsPrivateKey := os.Getenv("TLS_PRIVATE_KEY")
	bindAddress := os.Getenv("BIND_ADDRESS")
	port := os.Getenv("PORT")

	if bindAddress == "" {
		bindAddress = "127.0.0.1"
	}

	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf("%s:%s", bindAddress, port)

	http.HandleFunc("/", handler.HomeHandler)                    // Route pour la page d'accueil
	http.HandleFunc("/report-table", handler.ReportTableHandler) // Route pour récupérer la table HTML

	log.Printf("Serveur démarré sur : http://%s", address)

	// Activer HTTPS si nécessaire
	if tlsEnable == "YES" {
		if tlsCert == "" || tlsPrivateKey == "" {
			log.Fatal("TLS_ENABLE est activé mais TLS_CERT ou TLS_PRIVATE_KEY est manquant")
		}
		log.Println("Démarrage du serveur en mode HTTPS")
		log.Fatal(http.ListenAndServeTLS(address, tlsCert, tlsPrivateKey, nil))
	} else {
		log.Println("Démarrage du serveur en mode HTTP")
		log.Fatal(http.ListenAndServe(address, nil))
	}
}
