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
	// Effectuer la requête GET vers get_repport.php
	resp, err := http.Get("https://crm.vicitelecom.fr/vicidial/get_repport.php")
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
	var agents []AgentData
	err = json.Unmarshal(body, &agents)
	if err != nil {
		log.Println(err)
		http.Error(w, "Erreur lors du décodage des données JSON", http.StatusInternalServerError)
		return
	}

	// Calculer la durée pour chaque agent entre current_datetime et last_call_time.
	for i := range agents {
		duration, err := calculateDuration(agents[i].LastCallTime)
		if err == nil {
			agents[i].LastCallTime = duration // Remplacer last_call_time par la durée calculée
		} else {
			agents[i].LastCallTime = "Erreur de calcul"
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
		<style>
			table {
				width: 100%;
				border-collapse: collapse;
			}
			th, td {
				border: 1px solid #ccc;
				padding: 8px;
				text-align: left;
			}
			th {
				background-color: #f4f4f4;
			}
		</style>
	</head>
	<body>
		<h1>Agent Report</h1>
		<table>
			<thead>
				<tr>
					<th>Live Agent ID</th>
					<th>User</th>
					<th>Server IP</th>
					<th>Status</th>
					<th>Campaign ID</th>
					<th>Calls Today</th>
					<th>Duration (MM:SS)</th>
				</tr>
			</thead>
			<tbody>
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
			</tbody>
		</table>
	</body>
	</html>
	`

	// Compiler et exécuter le template.
	t := template.Must(template.New("report").Parse(tmpl))
	if err := t.Execute(w, Data{Agents: agents}); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
	}
}
