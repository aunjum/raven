package main

import (
	"GOLANG/news"
	_ "GOLANG/news"
	"bytes"
	_ "bytes"
	"context"
	_ "encoding/json"
	"fmt"
	_ "fmt"
	_ "github.com/boltdb/bolt"
	_ "github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	"html/template"
	_ "io/ioutil"
	"log"
	_ "log"
	"math"
	"net/http"
	_ "net/http"
	"net/url"
	_ "net/url"
	"os"
	_ "os"
	"strconv"
	_ "strconv"
	"time"
	_ "time"
)

//1
/*
func main() {
	http.HandleFunc("/", handler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}

	// Open my.db data file in your current directory.
	// It will be created if it doesn't exist.
	// Opening an already open Bolt database will cause it to hang until the other process closes it.
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	// Read-write transactions
	err = db.Update(func(tx *bolt.Tx) error {
		//
		return nil
	})

	// Read-only transactions
	err = db.View(func(tx *bolt.Tx) error {
		//
		return nil
	})

	// Batch read-write transactions
	err = db.Batch(func(tx *bolt.Tx) error {
		//
		return nil
	})
}

func handler(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "Hello World!")
	if err != nil {
		return
	}
}
*/

//2
/*
func main() {
	r := mux.NewRouter()

	r.HandleFunc("/hello", handler).Methods("GET")

	err := http.ListenAndServe("8080", r)
	if err != nil {
		return
	}
}

func handler(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "Hello World!")
	if err != nil {
		return
	}
}
*/

//3

type Search struct {
	Query      string
	NextPage   int
	TotalPages int
	Results    *news.Results
}

var newsApi *news.Client

// This is a user defined method to close resources.
// This method closes mongoDB connection and cancel context.
func disconnect(client *mongo.Client, ctx context.Context,
	cancel context.CancelFunc) {

	// CancelFunc to cancel to context
	defer cancel()

	// client provides a method to close
	// a mongoDB connection.
	defer func() {

		// client.Disconnect method also has deadline.
		// returns error if any,
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

// This is a user defined method that returns mongo.Client,
// context.Context, context.CancelFunc and error.
// mongo.Client will be used for further database operation.
// context.Context will be used set deadlines for process.
// context.CancelFunc will be used to cancel context and
// resource associated with it.

func connect(uri string) (*mongo.Client, context.Context,
	context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

// This is a user defined method that accepts
// mongo.Client and context.Context
// This method used to ping the mongoDB, return error if any.
func ping(client *mongo.Client, ctx context.Context) error {

	// mongo.Client has Ping to ping mongoDB, deadline of
	// the Ping method will be determined by cxt
	// Ping method return error if any occurred, then
	// the error can be handled.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	fmt.Println("connected successfully")
	return nil
}

func main() {
	// Get Client, Context, CalcelFunc and
	// err from connect method.
	uri := os.Getenv("MONGO_URL")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	client, ctx, cancel, err := connect(uri)
	if err != nil {
		panic(err)
	}

	// Release resource when the main
	// function is returned.
	defer disconnect(client, ctx, cancel)

	// Ping mongoDB with Ping method
	err = ping(client, ctx)
	if err != nil {
		return
	}

	//var _ *news.Client
	err = godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	apiKey := os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		log.Fatal("Env: apiKey must be set")
	}

	myClient := &http.Client{Timeout: 10 * time.Second}
	_ = news.NewClient(myClient, apiKey, 20)

	fs := http.FileServer(http.Dir("assets"))

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/", indexHandler)
	//mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/search", searchHandler(newsApi))
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		return
	}
	if err != nil {
		return
	}
}

var tpl = template.Must(template.ParseFiles("tpl/index2.html"))

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	buf := &bytes.Buffer{}
	err := tpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		return
	}
}

/*
func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := u.Query()
	searchQuery := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	fmt.Println("Search Query is: ", searchQuery)
	fmt.Println("Page is: ", page)
}
*/

func searchHandler(newsApi *news.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		params := u.Query()
		searchQuery := params.Get("q")
		page := params.Get("page")
		if page == "" {
			page = "1"
		}

		results, err := newsApi.FetchEverything(searchQuery, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextPage, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		search := &Search{
			Query:      searchQuery,
			NextPage:   nextPage,
			TotalPages: int(math.Ceil(float64(results.TotalResults) / float64(newsApi.PageSize))),
			Results:    results,
		}

		buf := &bytes.Buffer{}
		err = tpl.Execute(buf, search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = buf.WriteTo(w)
		if err != nil {
			return
		}

		fmt.Printf("%+v", results)

		fmt.Println("Search Query is: ", searchQuery)
		fmt.Println("Page is: ", page)
	}
}
