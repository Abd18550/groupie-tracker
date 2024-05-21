package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	
)

var groups []group

type concertlineup struct {
	Location string
	Dates    []string
}

type group struct {
	ID           int
	Image        string
	Name         string
	Members      []string
	FirstAlbum   string
	CreationDate string
	Concerts     []concertlineup
}

func main() {
	parseAPI()

	//censors
	groups[20].Image="https://upload.wikimedia.org/wikipedia/commons/7/71/Black.png"

	fmt.Println("parsed")

	http.HandleFunc("/", handleHome)
	http.Handle("/style/", http.StripPrefix("/style/", http.FileServer(http.Dir("style"))))
	http.ListenAndServe(":8080", nil)
}

func parseAPI() {
	var data map[string]interface{}

	json.Unmarshal(fetchFile("https://groupietrackers.herokuapp.com/api"), &data)

	var artists []struct {
		ID           int      `json:"id"`
		Image        string   `json:"image"`
		Name         string   `json:"name"`
		Members      []string `json:"members"`
		CreationDate int      `json:"creationDate"`
		FirstAlbum   string   `json:"firstAlbum"`
	}

	json.Unmarshal(fetchFile(data["artists"].(string)), &artists)

	var concertData map[string]interface{}

	err := json.Unmarshal(fetchFile(data["relation"].(string)), &concertData)
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(artists); i++ {
		concerts := concertData["index"].([]interface{})[i].(map[string]interface{})["datesLocations"].(map[string]interface{})
		var concertArray []concertlineup

		for location, dates := range concerts {
			var c concertlineup
			c.Location = strings.ReplaceAll(strings.ReplaceAll(location,"_"," "),"-",", ")

			var dateArray []string
			for _, date := range dates.([]interface{}) {
				dateArray = append(dateArray, strings.ReplaceAll(date.(string),"-","."))
			}
			c.Dates = dateArray
			concertArray = append(concertArray, c)
		}

	

		thing := group{
			ID:           i,
			Name:         artists[i].Name,
			Members:      artists[i].Members,
			FirstAlbum:   artists[i].FirstAlbum,
			CreationDate: strconv.Itoa(artists[i].CreationDate),
			Image:        artists[i].Image,
			Concerts:     concertArray,
		}

		groups = append(groups, thing)
	}
}

func fetchFile(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer response.Body.Close()

	copied, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return copied
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, groups); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
