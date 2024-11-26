package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)
type DisableUserRequest struct {
	User string `json:"user"`
}
func DisableUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var req DisableUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erreur de décodage JSON", http.StatusBadRequest)
		return
	}

	// Logique pour déconnecter l'utilisateur (exemple)
	fmt.Printf("Déconnexion de l'utilisateur : %s\n", req.User)

	// Réponse de confirmation
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Utilisateur déconnecté avec succès"))
}


	

