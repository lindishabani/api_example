package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	DBHost  = "127.0.0.1"
	DBPort  = ":3306"
	DBUser  = "root"
	DBPass  = ""
	DBDbase = "api"
)

type event struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

type allEvents []event

var events = allEvents{}

var database *sql.DB

func main() {
	dbConn := fmt.Sprintf("%s:%s@tcp(%s)/%s", DBUser, DBPass, DBHost, DBDbase)
	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		panic("Couldn't connect!")
	}

	database = db

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink).Methods("GET")
	router.HandleFunc("/event", createEvent).Methods("POST")
	router.HandleFunc("/events", getAllEvents).Methods("GET")
	router.HandleFunc("/events/{id:[0-9]+}", getOneEvent).Methods("GET")
	router.HandleFunc("/events/{id:[0-9]+}", updateEvent).Methods("PATCH")
	router.HandleFunc("/events/{id:[0-9]+}", deleteEvent).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	var newEvent event
	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &newEvent)
	stmt, err := database.Prepare("INSERT INTO event SET title = ?, description = ?")
	checkErr(err)
	_, err = stmt.Exec(newEvent.Title, newEvent.Description)
	checkErr(err)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newEvent)
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]

	e := event{}

	err := database.QueryRow("SELECT id, title, description FROM event WHERE id=?", eventID).Scan(&e.ID, &e.Title, &e.Description)
	checkErr(err)

	json.NewEncoder(w).Encode(e)
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	e := event{}
	rows, err := database.Query("SELECT id, title, description FROM event")
	checkErr(err)

	for rows.Next() {
		rows.Scan(&e.ID, &e.Title, &e.Description)
		events = append(events, e)
	}

	json.NewEncoder(w).Encode(events)
}

func updateEvent(w http.ResponseWriter, r *http.Request) {

	eventID := mux.Vars(r)["id"]
	e := event{}

	reqBody, err := ioutil.ReadAll(r.Body)
	checkErr(err)

	json.Unmarshal(reqBody, &e)

	stmt, err := database.Prepare("UPDATE event SET title = ?, description = ? WHERE id=?")
	checkErr(err)
	_, err = stmt.Exec(e.Title, e.Description, eventID)
	checkErr(err)

}

func deleteEvent(w http.ResponseWriter, r *http.Request) {

	eventID := mux.Vars(r)["id"]

	stmt, err := database.Prepare("DELETE FROM event WHERE id=?")
	checkErr(err)

	_, err = stmt.Exec(eventID)
	checkErr(err)
}

func checkErr(err error) {

	if err != nil {
		panic(err)
	}
}
