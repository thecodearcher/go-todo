package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

// ResponseObj Custom type for sending response to client
type ResponseObj struct {
	Message string      `json:"message"`
	Status  bool        `json:"status"`
	Data    interface{} `json:"data"`
}

//Todo data structure for Todo model
type Todo struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// getTodos   handler
func getTodos(w http.ResponseWriter, r *http.Request) {
	rsp, err := db.Query("Select * from todos")
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	defer rsp.Close()
	res := []Todo{}
	for rsp.Next() {
		var id int
		var title string
		var body string
		var createdAt string
		err = rsp.Scan(&id, &title, &body, &createdAt)
		if err != nil {
			panic(err.Error())
		}
		res = append(res, Todo{Id: id, Title: title, Body: body, CreatedAt: createdAt})
	}
	json.NewEncoder(w).Encode(formatData(res, "OK", true))
}

func saveTodo(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

	stmt, err := db.Prepare("INSERT INTO todos(title,body,created_at) VALUES(?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(title, body, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(formatData(nil, "Todo Created!", true))
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := mux.Vars(r)["todo"]
	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

	stmt, err := db.Prepare("UPDATE todos set title = ?,body =? where id = ?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(title, body, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(formatData(nil, err.Error(), false))

	}
	json.NewEncoder(w).Encode(formatData(nil, "Todo Id: "+id+" Updated!", true))
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := mux.Vars(r)["todo"]

	stmt, err := db.Prepare("DELETE from todos where id =?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(formatData(nil, err.Error(), false))

	}
	json.NewEncoder(w).Encode(formatData(nil, "Todo Id: "+id+" Deleted!", true))
}

func getTodo(w http.ResponseWriter, r *http.Request) {
	todoID := mux.Vars(r)["todo"]
	var id int
	var title string
	var body string
	var createdAt string
	err := db.QueryRow("select * from todos where id =? ", todoID).Scan(&id, &title, &body, &createdAt)
	res := Todo{Id: id, Title: title, Body: body, CreatedAt: createdAt}
	if err != nil {
		if err == sql.ErrNoRows {
			json.NewEncoder(w).Encode(formatData(nil, "No todo added with Id: "+todoID, true))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(formatData(nil, err.Error(), false))
		}
		return
	}
	json.NewEncoder(w).Encode(formatData(res, "OK", true))
}

func main() {
	db = initializeDb()
	defer db.Close()
	router := mux.NewRouter()
	router.Use(commonMiddleware)

	// Routes consist of a path and a handler function.
	router.HandleFunc("/", saveTodo).Methods("POST")
	router.HandleFunc("/", getTodos).Methods("GET")
	router.HandleFunc("/{todo}", getTodo).Methods("GET")
	router.HandleFunc("/{todo}", updateTodo).Methods("PUT")
	router.HandleFunc("/{todo}", deleteTodo).Methods("DELETE")

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", router))

}

func initializeDb() *sql.DB {
	db, err := sql.Open("mysql", "<username>:<password>@tcp(127.0.0.1:3306)/todo")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Database connected")
	return db
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func formatData(data interface{}, message string, status bool) ResponseObj {

	return ResponseObj{
		Message: message,
		Data:    data,
		Status:  status,
	}

}
