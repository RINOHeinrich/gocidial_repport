package handler

import (
	"net/http"
	"text/template"
)

// Fonction de gestion de la route d'accueil (page principale).
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
	}
}

// Fonction de gestion de la route d'accueil (page principale).
func AgentReportHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("report.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
	}
}
