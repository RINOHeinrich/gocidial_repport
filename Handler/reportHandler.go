package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"
	"time"

	models "github.com/RINOHeinrich/gocidial/Models"
)

// Gestionnaire pour la route /report-table, renvoie uniquement le corps de la table.
func ReportTableHandler(w http.ResponseWriter, r *http.Request) {
	var allAgents []models.AgentData
	var agents []models.AgentData

	dataSource := []string{
		"https://crm.vicitelecom.fr/vicidial/get_repport.php",
		"https://axe-formation3.vicitelecom.fr/vicidial/get_repport.php",
	}

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

		var anyFetchError models.FetchError
		err = json.Unmarshal(body, &anyFetchError)
		if err == nil {
			continue
		}

		if err := json.Unmarshal(body, &agents); err != nil {
			http.Error(w, "Erreur lors du décodage des données JSON", http.StatusInternalServerError)
			return
		}

		allAgents = append(allAgents, agents...)
	}

	for i := range allAgents {
		duration, err := calculateDuration(allAgents[i].LastCallTime)
		if err == nil {
			allAgents[i].LastCallTime = duration
		} else {
			allAgents[i].LastCallTime = "Erreur de calcul"
		}
	}

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
	if err := t.Execute(w, models.Data{Agents: allAgents}); err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
	}
}

// Fonction pour calculer la durée depuis la dernière appel.
func calculateDuration(lastCallTime string) (string, error) {
	const layout = "2006-01-02 15:04:05"
	callTime, err := time.ParseInLocation(layout, lastCallTime, time.Local)
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
