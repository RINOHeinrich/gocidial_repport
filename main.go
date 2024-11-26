package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	handler "github.com/RINOHeinrich/gocidial/Handler"
	models "github.com/RINOHeinrich/gocidial/Models"
	"github.com/joho/godotenv"
)

// LoadServerConfigs charge les configurations des serveurs à partir des variables d'environnement
func LoadServerConfigs() map[string]models.ServerConfig {
	servers := make(map[string]models.ServerConfig)
	for _, env := range os.Environ() {
		//	log.Println(env)
		if strings.HasPrefix(env, "SERVER") {
			parts := strings.SplitN(env, "=", 2)
			key := parts[0]
			value := parts[1]
			log.Println("clé: "+key, "value: "+value)
			// Identifie l'index et le champ (URL, USERNAME, PASSWORD)
			var serverIndex int
			var field string
			fmt.Sscanf(key, "SERVER%d_%s", &serverIndex, &field)
			log.Println("field: " + field)
			// Ajoute ou met à jour la configuration du serveur
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
	log.Println(servers)
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
	http.HandleFunc("/", handler.HomeHandler) // Route pour la page d'accueil
	http.HandleFunc("/report-table", func(w http.ResponseWriter, r *http.Request) {
		handler.ReportTableHandler(w, r, serverConfigs)
	}) // Route pour récupérer la table HTML
	http.HandleFunc("/disable-users", func(w http.ResponseWriter, r *http.Request) {
		handler.DisableUserHandler(w, r, serverConfigs)
	})

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
