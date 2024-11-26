package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	models "github.com/RINOHeinrich/gocidial/Models"
)

/*
	type DisableUserRequest struct {
		User string `json:"user"`
	}
*/
func DisableUserHandler(w http.ResponseWriter, r *http.Request, servers map[string]models.ServerConfig) {
	// Vérifie que la méthode est POST
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Structure pour décoder les données reçues
	var data struct {
		User   string `json:"user"`
		Domain string `json:"domain"`
	}

	// Décodage du JSON reçu
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil || data.User == "" || data.Domain == "" {
		http.Error(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	// Recherche des informations de connexion pour le domaine spécifié
	var serverConfig models.ServerConfig
	found := false
	for _, server := range servers {
		if strings.Contains(server.URL, data.Domain) {
			serverConfig = server
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Domaine inconnu ou configuration non trouvée", http.StatusBadRequest)
		return
	}

	// URL cible pour déconnecter l'utilisateur
	requestURL := fmt.Sprintf("%svicidial/user_status.php", serverConfig.URL)
	log.Println(requestURL)

	// Génération de l'en-tête Authorization
	auth := serverConfig.Username + ":" + serverConfig.Password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	// Données du formulaire
	formData := url.Values{}
	formData.Set("DB", "0")
	formData.Set("user", data.User)
	formData.Set("stage", "log_agent_out")
	formData.Set("submit", "EMERGENCY LOG AGENT OUT")

	// Création de la requête POST
	req, err := http.NewRequest("POST", requestURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		http.Error(w, "Erreur lors de la création de la requête", http.StatusInternalServerError)
		return
	}

	// Ajout des en-têtes
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	// Client HTTP
	client := &http.Client{}

	// Exécution de la requête
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution de la requête", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Vérification du statut de la réponse
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Erreur lors de la déconnexion : %s", resp.Status), http.StatusInternalServerError)
		return
	}

	// Réponse JSON pour indiquer le succès
	//w.WriteHeader(http.StatusOK)
	ReportTableHandler(w, r, servers)

}
