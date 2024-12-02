package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	handler "github.com/RINOHeinrich/gocidial/Handler"
	middleware "github.com/RINOHeinrich/gocidial/Middleware"
	models "github.com/RINOHeinrich/gocidial/Models"
	"github.com/joho/godotenv"
)

// LoadServerConfigs charge les configurations des serveurs à partir des variables d'environnement
func LoadServerConfigs() map[string]models.ServerConfig {
	servers := make(map[string]models.ServerConfig)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "SERVER") {
			parts := strings.SplitN(env, "=", 2)
			key := parts[0]
			value := parts[1]
			var serverIndex int
			var field string
			fmt.Sscanf(key, "SERVER%d_%s", &serverIndex, &field)
			if serverIndex != 0 {
				server := servers[strconv.Itoa(serverIndex)]
				switch field {
				case "URL":
					server.URL = value
				case "USERNAME":
					server.Username = value
				case "PASSWORD":
					server.Password = value
				}
				servers[strconv.Itoa(serverIndex)] = server
			}
		}
	}
	return servers
}

func main() {
	// Charger les variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Erreur lors du chargement du fichier .env: %v", err)
	}

	// Charger les configurations des serveurs
	serverConfigs := LoadServerConfigs()
	for id, config := range serverConfigs {
		log.Printf("Serveur %s: URL=%s Username=%s", id, config.URL, config.Username)
	}

	// Récupérer les paramètres de configuration généraux
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

	// Ajouter vos handlers en passant les configurations des serveurs
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Sécuriser cette route avec le middleware d'authentification
		middleware.AuthMiddleware(http.HandlerFunc(handler.HomeHandler)).ServeHTTP(w, r)
	})

	// Sécuriser la route /report-table avec le middleware d'authentification
	http.HandleFunc("/report-table", func(w http.ResponseWriter, r *http.Request) {
		middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ReportTableHandler(w, r, serverConfigs)
		})).ServeHTTP(w, r)
	})

	// Autres routes
	http.HandleFunc("/disable-users", func(w http.ResponseWriter, r *http.Request) {
		middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.DisableUserHandler(w, r, serverConfigs)
		})).ServeHTTP(w, r)
	})
	http.HandleFunc("/autodial-info", func(w http.ResponseWriter, r *http.Request) {
		middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.AutodialInfoHandler(w, r, serverConfigs)
		})).ServeHTTP(w, r)
	})
	/* 	http.HandleFunc("/get-autodial-info", func(w http.ResponseWriter, r *http.Request) {
		middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.AutodialInfoHandler(w, r, serverConfigs)
		})).ServeHTTP(w, r)
	}) */
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		handler.AuthenticationHandler(w, r, serverConfigs)
	})
	http.HandleFunc("/logout", handler.LogoutHandler())

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
