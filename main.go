package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//Using struct tags to provide transformation info on how a struct field is encoded to or decoded from another format (or stored/retrieved from a database)
type Movie struct {
	MovieID string `"json:movieid"`
	MovieName string `"json:moviename"`
}

type JsonResponse struct {
	Type string `json:"type"`
	Data []Movie `json:"data"`
	Message string `json:"message"`	
}

func getEnvVars(){
	err := godotenv.Load("credentials.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
  

//DB Setup Function
func setupDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", dbinfo)

	checkErr(err)

	return db
}

func main(){
	getEnvVars()
	router := mux.NewRouter()

	router.HandleFunc("/movies/", GetMovies).Methods("GET")

	router.HandleFunc("/movies/", CreateMovie).Methods("POST")

	router.HandleFunc("/movies/{movieID}", DeleteMovie).Methods("DELETE")

	router.HandleFunc("/movies/", DeleteMovies).Methods("DELETE")

	fmt.Println("Listening at port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

//Check Error Function
func checkErr(err error){
	if err != nil{
		panic(err)
	}
}

//Print Message Function
func printMessage(message string){
	fmt.Println("")
	fmt.Println(message)
	fmt.Println("")
}

//Get All Movies
func GetMovies(w http.ResponseWriter, r *http.Request){
	db := setupDB()

	printMessage("Getting Movies...")

	rows, err := db.Query("SELECT * FROM movies")

	checkErr(err)

	var movies []Movie

	for rows.Next() {
		var id int
		var movieID string
		var movieName string
		
		err = rows.Scan(&id, &movieID, &movieName)

		checkErr(err)

		movies = append(movies, Movie{MovieID: movieID, MovieName: movieName})
	}

	var response = JsonResponse{Type: "success", Data: movies}

	json.NewEncoder(w).Encode(response)
}

//Insert Single Movie
func CreateMovie(w http.ResponseWriter, r *http.Request){
	movieID := r.FormValue("movieid")
	movieName := r.FormValue("moviename")

	var response = JsonResponse{}

	if movieID == "" || movieName == "" {
		response = JsonResponse{Type: "error", Message: "You are missing MovieID or MovieName"}
	} else {
		db := setupDB()

		printMessage("Inserting Movie into DB...")

		fmt.Println("Inserting new movie with ID: " + movieID + " and name: " + movieName)

		var lastInsertID int
		err := db.QueryRow("INSERT INTO movies(movie_id, movie_name) VALUES($1, $2) returning id;", movieID, movieName).Scan(&lastInsertID)
		
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "Movie has been successfully inserted"}
	}

	json.NewEncoder(w).Encode(response)
}

//Delete Single Movie
func DeleteMovie(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)

	movieID := params["movieID"]

	var response = JsonResponse{}

	if movieID == "" {
		response = JsonResponse{Type: "error", Message: "Missing Movie Id"}
	} else {
		db := setupDB()

		printMessage("Deleting movie from DB...")

		_, err := db.Exec("DELETE FROM movies WHERE movie_id = $1", movieID)
		
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "Successfully deleted movie from DB"}
	}
	json.NewEncoder(w).Encode(response)
}

//Delete All Movies
func DeleteMovies(w http.ResponseWriter, r *http.Request){
	db := setupDB()

	printMessage("Deleting all movies...")

	_, err := db.Exec("DELETE FROM movies")

	checkErr(err)

	printMessage("All movies have been deleted successfully")

	var response = JsonResponse{Type: "success", Message: "All movies have been deleted successfully"}
	json.NewEncoder(w).Encode(response)
}

