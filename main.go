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

// AgentData représente les données d'un agent.
type AgentData struct {
	LiveAgentID  string `json:"live_agent_id"`
	User         string `json:"user"`
	ServerIP     string `json:"server_ip"`
	Status       string `json:"status"`
	CampaignID   string `json:"campaign_id"`
	CallsToday   string `json:"calls_today"`
	LastCallTime string `json:"last_call_time"`
}
type fetchError struct {
	Message string `json:"message"`
}

// Data contient la liste des agents.
type Data struct {
	Agents []AgentData
}

func main() {
	http.HandleFunc("/", homeHandler)                    // Route pour la page d'accueil
	http.HandleFunc("/report-table", reportTableHandler) // Route pour récupérer la table HTML

	log.Println("Serveur démarré sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Fonction pour calculer la durée depuis la dernière appel.
func calculateDuration(lastCallTime string) (string, error) {
	const layout = "2006-01-02 15:04:05"
	callTime, err := time.ParseInLocation(layout, lastCallTime, time.Local)
	if err != nil {
		return "", err
	}
	// Convertir callTime en UTC
	callTimeUTC := callTime.UTC()

	// Utiliser l'heure actuelle en UTC
	currentTimeUTC := time.Now().UTC()

	log.Println("Current time in UTC:", currentTimeUTC)
	log.Println("Call time in UTC:", callTimeUTC)

	// Calculer la différence
	duration := currentTimeUTC.Sub(callTimeUTC)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds), nil
}

// Fonction de gestion de la route d'accueil (page principale).
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("report.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement du template", http.StatusInternalServerError)
		return
	}

	// Afficher la page principale
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
	}
}

// Gestionnaire pour la route /report-table, renvoie uniquement le corps de la table.
func reportTableHandler(w http.ResponseWriter, r *http.Request) {
	var allAgents []AgentData
	var agents []AgentData

	dataSource := []string{
		"https://crm.vicitelecom.fr/vicidial/get_repport.php",
		"https://axe-formation3.vicitelecom.fr/vicidial/get_repport.php",
	}

	// Récupérer les données des sources
	for _, url := range dataSource {
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des données", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture de la réponse", http.StatusInternalServerError)
			return
		}

		// Vérifier si la réponse contient un message d'erreur ("No records found")
		var anyFetchError fetchError
		err = json.Unmarshal(body, &anyFetchError)
		//log.Println(anyFetchError)
		if err == nil {
			log.Println(err)
			//log.Println("Aucune donnée trouvée pour l'URL:", url)
			continue
		}

		/* 	// Si le message est "No records found", ne pas essayer de désérialiser les agents
		if message, exists := response["message"]; exists && message == "No records found" {
			log.Println("Aucune donnée trouvée pour l'URL:", url)
			continue
		} */

		// Décoder les données JSON dans le tableau d'agents
		if err := json.Unmarshal(body, &agents); err != nil {
			log.Println("Erreur lors du décodage des agents:", err)
			http.Error(w, "Erreur lors du décodage des données JSON", http.StatusInternalServerError)
			return
		}

		// Ajouter les agents récupérés à la liste globale
		allAgents = append(allAgents, agents...)
	}

	// Calculer la durée pour chaque agent.
	for i := range allAgents {
		duration, err := calculateDuration(allAgents[i].LastCallTime)
		if err == nil {
			allAgents[i].LastCallTime = duration
		} else {
			allAgents[i].LastCallTime = "Erreur de calcul"
		}
	}

	// Générer la table HTML avec les données mises à jour
	tmpl := `
    {{range .Agents}}
    <tr>
        <td class="px-4 py-2 border border-gray-300">{{.User}}</td>
        <td class="px-4 py-2 border border-gray-300">{{.ServerIP}}</td>
        <td class="px-4 py-2 border border-gray-300">{{.Status}}</td>
		<td class="px-4 py-2 border border-gray-300">{{.LastCallTime}}</td>
        <td class="px-4 py-2 border border-gray-300">{{.CampaignID}}</td>
        <td class="px-4 py-2 border border-gray-300">{{.CallsToday}}</td>
    </tr>
    {{end}}
    `

	t := template.Must(template.New("table").Parse(tmpl))
	if err := t.Execute(w, Data{Agents: allAgents}); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
	}
}
