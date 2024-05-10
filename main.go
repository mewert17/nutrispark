package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoURI = "mongodb://localhost:27017"

type Query struct {
	Query     string    `bson:"query"`
	CreatedAt time.Time `bson:"created_at"`
}

func main() {
	// Set up MongoDB connection
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access a MongoDB collection
	database := client.Database("local")
	collection := database.Collection("queries")

	// Modify the existing "/" route to serve index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Serve other HTML pages for specific routes
	http.HandleFunc("/results", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "results.html")
	})

	http.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "sum.html")
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "login.html")
	})

	http.HandleFunc("/recommendations", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "recommendations.html")
	})

	// Handle requests to the /nutrition endpoint
	http.HandleFunc("/nutrition", func(w http.ResponseWriter, r *http.Request) {
		// Get the query parameter from the request URL
		query := r.URL.Query().Get("query")

		// Make a request to the API with the query
		url := fmt.Sprintf("https://nutrition-by-api-ninjas.p.rapidapi.com/v1/nutrition?query=%s", query)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			log.Println("Failed to create request:", err)
			return
		}
		req.Header.Add("X-RapidAPI-Key", "367e572633mshc2ec0281249b7a9p143421jsn8bd3269e2aad")
		req.Header.Add("X-RapidAPI-Host", "nutrition-by-api-ninjas.p.rapidapi.com")

		// Send the request
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch data from API", http.StatusInternalServerError)
			log.Println("Failed to fetch data from API:", err)
			return
		}
		defer res.Body.Close()

		// Copy the response from the API to the response writer
		io.Copy(w, res.Body)

		// Save the query to MongoDB
		_, err = collection.InsertOne(context.Background(), bson.M{"query": query, "created_at": time.Now()})
		if err != nil {
			log.Println("Failed to insert query into MongoDB:", err)
		}
	})

	// Handle requests to fetch past queries
	http.HandleFunc("/pastqueries", func(w http.ResponseWriter, r *http.Request) {
		// Define a slice to store the retrieved queries
		var queries []Query

		// Query MongoDB for past queries
		cursor, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			log.Println("Failed to query MongoDB:", err)
			http.Error(w, "Failed to retrieve past queries", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Iterate over the cursor and decode each document into a Query struct
		for cursor.Next(context.Background()) {
			var query Query
			if err := cursor.Decode(&query); err != nil {
				log.Println("Failed to decode query:", err)
				continue
			}
			queries = append(queries, query)
		}
		if err := cursor.Err(); err != nil {
			log.Println("Cursor error:", err)
			http.Error(w, "Failed to retrieve past queries", http.StatusInternalServerError)
			return
		}

		// Encode the retrieved queries into JSON and write it to the response writer
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(queries); err != nil {
			log.Println("Failed to encode queries:", err)
			http.Error(w, "Failed to encode past queries", http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
