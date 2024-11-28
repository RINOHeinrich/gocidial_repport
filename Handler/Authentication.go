package handler

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"

	models "github.com/RINOHeinrich/gocidial/Models"
)

// isAuthenticated vérifie si l'utilisateur est connecté en recherchant un jeton d'authentification (par exemple un cookie "auth_token")
func IsAuthenticated(r *http.Request) bool {
	// Vérifier si le cookie "auth_token" existe
	cookie, err := r.Cookie("auth_token")
	if err != nil || cookie == nil {
		return false
	}

	// Ici, vous pouvez valider le JWT. Pour l'instant, on suppose qu'un jeton est suffisant pour l'authentification.
	return true
}

// AuthenticationHandler gère l'authentification auprès de plusieurs serveurs
func AuthenticationHandler(w http.ResponseWriter, r *http.Request, servers map[string]models.ServerConfig) {
	if r.Method == http.MethodGet {
		// Si l'utilisateur n'est pas connecté, retourner auth.html
		http.ServeFile(w, r, "auth.html")
		return
	}

	if r.Method == http.MethodPost {
		// Parse les données du formulaire
		if err := r.ParseForm(); err != nil {
			log.Printf("Erreur lors de l'analyse des données postées: %v\n", err)
			http.Error(w, "Requête invalide", http.StatusBadRequest)
			return
		}

		// Récupérer les informations du formulaire
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "" || password == "" {
			http.Error(w, "Username et password sont requis", http.StatusBadRequest)
			return
		}

		// Préparer les données pour la requête POST
		authData := map[string]string{
			"username": username,
			"password": password,
		}
		authJSON, err := json.Marshal(authData)
		if err != nil {
			log.Printf("Erreur lors de la sérialisation des données: %v\n", err)
			http.Error(w, "Erreur interne", http.StatusInternalServerError)
			return
		}

		// Parcourir tous les serveurs pour tenter l'authentification
		for _, server := range servers {
			url := server.URL + ":8099/login"
			log.Printf("Tentative d'authentification auprès de %s\n", url)
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}

			// Crée un client HTTP avec le transport personnalisé
			client := &http.Client{Transport: tr}

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(authJSON))
			if err != nil {
				log.Printf("Erreur lors de la création de la requête: %v\n", err)
				continue
			}

			req.Header.Set("Content-Type", "application/json")

			// Envoyer la requête
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Erreur lors de l'envoi de la requête: %v\n", err)
				continue
			}
			defer resp.Body.Close()

			// Vérifier le code de réponse
			if resp.StatusCode == http.StatusOK {
				// Lire le body pour récupérer le token
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Erreur lors de la lecture de la réponse: %v\n", err)
					continue
				}

				// Supposons que le token est dans le body sous forme brute
				token := string(body)

				// Ajouter le token au cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    token,
					Path:     "/",
					HttpOnly: true, // Empêche l'accès via JavaScript
					Secure:   true, // Nécessite HTTPS
				})

				// Rediriger vers la page des rapports
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			log.Printf("Échec d'authentification auprès de %s: %s\n", url, resp.Status)
		}

		// Si aucun serveur n'a validé l'authentification
		log.Println("Échec d'authentification auprès de tous les serveurs")
		http.Error(w, "Erreur d'authentification", http.StatusUnauthorized)
	}
}
