package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

func DisconnectUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		requestURL := "https://axe.com/vicidial/user_status.php"

		// Données du formulaire
		data := url.Values{}
		data.Set("DB", "0")
		data.Set("user", "11180")
		data.Set("stage", "log_agent_out")
		data.Set("submit", "EMERGENCY LOG AGENT OUT")

		// Crée une requête POST
		req, err := http.NewRequest("POST", requestURL, bytes.NewBufferString(data.Encode()))
		if err != nil {
			fmt.Println("Erreur lors de la création de la requête:", err)
			return
		}

		// Ajoute les en-têtes nécessaires
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Client HTTP
		client := &http.Client{}

		// Exécute la requête
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Erreur lors de l'exécution de la requête:", err)
			return
		}
		defer resp.Body.Close()

		// Affiche le statut de la réponse
		fmt.Println("Statut de la réponse:", resp.Status)
	}
}
