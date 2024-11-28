package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	models "github.com/RINOHeinrich/gocidial/Models"
)

func extractDomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}

// Gestionnaire pour la route /report-table, renvoie uniquement le corps de la table.
func ReportTableHandler(w http.ResponseWriter, r *http.Request, servers map[string]models.ServerConfig) {
	var allAgents []models.AgentData
	var agents []models.AgentData

	for _, server := range servers {
		url := server.URL + "/vicidial/get_repport.php"
		//log.Println(url)
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

		var anyFetchError models.FetchError
		err = json.Unmarshal(body, &anyFetchError)
		if err == nil {
			continue
		}

		if err := json.Unmarshal(body, &agents); err != nil {
			http.Error(w, "Erreur lors du décodage des données JSON", http.StatusInternalServerError)
			return
		}
		// Ajouter le domaine source pour chaque agent
		domain := extractDomain(url)
		for i := range agents {
			agents[i].SourceDomain = domain
		}

		allAgents = append(allAgents, agents...)
	}

	for i := range allAgents {
		//	log.Println(allAgents[i].LastStateChange)
		//	log.Println(allAgents[i].LastCallTime)
		duration, err := calculateDuration(allAgents[i].LastStateChange)
		if err == nil {
			allAgents[i].LastCallTime = duration
		} else {
			allAgents[i].LastCallTime = "Erreur de calcul"
		}

		minutes, err := extractMinutes(allAgents[i].LastCallTime)
		if err == nil && allAgents[i].Status == "READY" {
			if minutes < 1 {
				allAgents[i].RowClass = "bg-blue-200" // Bleu ciel
			} else if minutes >= 1 && minutes <= 5 {
				allAgents[i].RowClass = "bg-blue-500" // Bleu roi
			} else {
				allAgents[i].RowClass = "bg-blue-800" // Bleu marine
			}
		} else {
			allAgents[i].RowClass = ""
		}
		seconds, err := extractSeconds(allAgents[i].LastCallTime)
		if err == nil && allAgents[i].Status == "PAUSED" {
			if seconds < 10 && minutes == 0 {
				allAgents[i].RowClass = "bg-slate-50"
			} else if minutes < 1 {
				allAgents[i].RowClass = "bg-yellow-100"
			} else if minutes < 5 {
				allAgents[i].RowClass = "bg-yellow-400"
			} else if minutes < 10 {
				allAgents[i].RowClass = "bg-lime-600"
			} else if minutes < 15 {
				allAgents[i].RowClass = "bg-lime-700"
			} else {
				allAgents[i].RowClass = "bg-yellow-900"
			}
		}
	}

	tmpl := `
{{range .Agents}}
<tr class="{{.RowClass}}">
    <td class="px-4 py-2 border border-gray-300">{{.User}}</td>
    <td class="px-4 py-2 border border-gray-300">{{.ServerIP}}</td>
    <td class="px-4 py-2 border border-gray-300">{{.Status}}</td>
    <td class="px-4 py-2 border border-gray-300">{{.LastCallTime}}</td>
    <td class="px-4 py-2 border border-gray-300">{{.CampaignID}}</td>
    <td class="px-4 py-2 border border-gray-300">{{.CallsToday}}</td>
    <td class="px-4 py-2 border border-gray-300">
		<button 
   		 class="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded flex items-center gap-2"
   		 hx-post="/disable-users"
		 hx-ext="json-enc" 
    	 hx-vals='{"user": "{{.User}}", "domain": "{{.SourceDomain}}"}'
    	 hx-target="#agent-table"
		 hx-swapp="innerHTML"
		 hx-confirm="Êtes vous sure de vouloir déconnecter l'agent {{.User}}?"
		 >
    	 Déconnecter
		</button>

    </td>
</tr>
{{end}}


    `

	t := template.Must(template.New("table").Parse(tmpl))
	if err := t.Execute(w, models.Data{Agents: allAgents}); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
	}
}

// Fonction pour calculer la durée depuis la dernière appel.
func calculateDuration(timeForCalcul string) (string, error) {
	const layout = "2006-01-02 15:04:05"
	callTime, err := time.ParseInLocation(layout, timeForCalcul, time.Local)
	if err != nil {
		return "", err
	}

	callTimeUTC := callTime.UTC()
	currentTimeUTC := time.Now().UTC()
	duration := currentTimeUTC.Sub(callTimeUTC)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds), nil
}

// Fonction pour extraire les secondes d'une chaîne au format MM:SS.
func extractSeconds(mmss string) (int, error) {
	parts := strings.Split(mmss, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("format invalide")
	}

	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return seconds, nil
}

func extractMinutes(mmss string) (int, error) {
	parts := strings.Split(mmss, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("format invalide")
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	return minutes, nil
}
