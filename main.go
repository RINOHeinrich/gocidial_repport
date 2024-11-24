package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// AgentData représente les données d'un agent, avec last_call_time ajouté.
type AgentData struct {
	LiveAgentID  string `json:"live_agent_id"`
	User         string `json:"user"`
	ServerIP     string `json:"server_ip"`
	Status       string `json:"status"`
	CampaignID   string `json:"campaign_id"`
	CallsToday   string `json:"calls_today"`
	LastCallTime string `json:"last_call_time"` // Champ modifié
}

// Data contient la liste des agents.
type Data struct {
	Agents []AgentData
}

func main() {
	http.HandleFunc("/report", reportHandler)

	log.Println("Serveur démarré sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Fonction pour calculer la différence entre current_datetime et last_call_time au format "2006-01-02 15:04:05".
func calculateDuration(lastCallTime string) (string, error) {
	const layout = "2006-01-02 15:04:05"

	// Convertir lastCallTime en objet time.Time dans le fuseau horaire local
	callTime, err := time.ParseInLocation(layout, lastCallTime, time.Local) // Utilisation du fuseau horaire local
	if err != nil {
		return "", err
	}

	// Obtenir l'heure actuelle dans le fuseau horaire local
	currentTime := time.Now() // time.Now() prend automatiquement le fuseau horaire local

	// Calculer la différence de temps entre currentTime et callTime
	duration := currentTime.Sub(callTime)

	// Calculer la durée en minutes et secondes
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60

	// Retourner la durée sous la forme MM:SS
	return fmt.Sprintf("%02d:%02d", minutes, seconds), nil
}

// Gestionnaire pour la route /report.
func reportHandler(w http.ResponseWriter, r *http.Request) {
	var allAgents []AgentData
	var agents []AgentData

	// Effectuer la requête GET vers get_repport.php
	dataSource := []string{
		"https://crm.vicitelecom.fr/vicidial/get_repport.php",
		"https://axe-formation3.vicitelecom.fr/vicidial/get_repport.php",
	}

	for _, url := range dataSource {
		log.Println(url)
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des données", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Lire la réponse
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture de la réponse", http.StatusInternalServerError)
			return
		}

		// Décoder les données JSON
		err = json.Unmarshal(body, &agents)
		if err != nil {
			log.Println(err)
			http.Error(w, "Erreur lors du décodage des données JSON", http.StatusInternalServerError)
			return
		}
		allAgents = append(allAgents, agents...)
		log.Println(allAgents)
	}

	// Calculer la durée pour chaque agent entre current_datetime et last_call_time.
	for i := range allAgents {
		duration, err := calculateDuration(allAgents[i].LastCallTime)
		if err == nil {
			allAgents[i].LastCallTime = duration // Remplacer last_call_time par la durée calculée
		} else {
			allAgents[i].LastCallTime = "Erreur de calcul"
		}
	}

	// Générer la table HTML.
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Report</title>
</head>
<body>
<table>
	<tr>
		<th>Live Agent ID</th>
		<th>User</th>
		<th>Server IP</th>
		<th>Status</th>
		<th>Campaign ID</th>
		<th>Calls Today</th>
		<th>Last Call Time</th>
	</tr>
	{{range .Agents}}
	<tr>
		<td>{{.LiveAgentID}}</td>
		<td>{{.User}}</td>
		<td>{{.ServerIP}}</td>
		<td>{{.Status}}</td>
		<td>{{.CampaignID}}</td>
		<td>{{.CallsToday}}</td>
		<td>{{.LastCallTime}}</td>
		</tr>
			{{end}}
</body>
</html>
`
	// Compiler et exécuter le template.
	t := template.Must(template.New("report").Parse(tmpl))
	if err := t.Execute(w, Data{Agents: allAgents}); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
	}
}
