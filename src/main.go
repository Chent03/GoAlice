package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nlopes/slack"
)

type nobleEmployee struct {
	ID      string
	Profile slack.UserProfile
}

type serverResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type vistor struct {
	First    string `json:"firstName"`
	Last     string `json:"lastName"`
	Purpose  string `json:"purpose"`
	NobleEmp string `json:"nobleEmp"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(run())
}

func run() error {
	mux := makeMuxRouter()
	log.Println("Listening on port 3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal("Failed to start server", err)
		return err
	}
	return nil
}

func makeMuxRouter() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc("/noble", handleNoble).Methods("GET")
	m.HandleFunc("/noble/{nobleID}", handleNoble).Methods("POST")
	return m
}

func handleNoble(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getNobleEmployees(w, r)
	case "POST":
		msgNobleEmployee(w, r)
	default:
		w.Write([]byte("Method does not exist"))
	}
}

func getNobleEmployees(w http.ResponseWriter, r *http.Request) {
	e := []nobleEmployee{}

	api := slack.New(os.Getenv("BOT_TOKEN"))
	users, err := api.GetUsers()
	if err != nil {
		fmt.Printf("There was an error getting the list of Noble Employees %s\n", err)
		respondWithJSON(w, r, http.StatusInternalServerError, e)
	}
	for _, user := range users {
		e = append(e, nobleEmployee{user.ID, user.Profile})
	}
	spew.Dump(users)
	respondWithJSON(w, r, http.StatusCreated, e)
}

func msgNobleEmployee(w http.ResponseWriter, r *http.Request) {
	reply := serverResponse{}
	v := vistor{}
	vars := mux.Vars(r)
	nobleID := vars["nobleID"]
	api := slack.New(os.Getenv("BOT_TOKEN"))

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Panic("Couldn't not read incoming JSON")
	}

	if err := json.Unmarshal(body, &v); err != nil {
		log.Panic("Error is json parsing")
	}

	spew.Dump(v)

	sendMessage := fmt.Sprintf("Hey %s, %s %s is here for you at the front desk.", v.NobleEmp, v.First, v.Last)

	params := slack.PostMessageParameters{AsUser: true}
	attachment := slack.Attachment{
		Color:   "#36a64f",
		Pretext: "*Purpose*",
		Text:    v.Purpose,
	}

	params.Attachments = []slack.Attachment{attachment}

	_, _, er := api.PostMessage(nobleID, sendMessage, params)

	if er != nil {
		reply.Success = false
		reply.Message = "Failed to send message"
		fmt.Printf("There was an error getting the list of Noble Employees %s\n", err)
		respondWithJSON(w, r, http.StatusInternalServerError, reply)
	}
	reply.Success = true
	reply.Message = "Message successfully sent"
	respondWithJSON(w, r, http.StatusCreated, reply)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTPP 500: Internal Server Error"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
