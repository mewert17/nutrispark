
//index.html

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NutriSpark</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #553e23; //Lightgrey
            margin: 0;
            padding: 0;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
            background-color: #ddc09b; // White
            border-radius: 8px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1); // Light shadow
            overflow-x: auto; // Add overflow property

        }

        h1 {
            color: #254726; // Green
            text-align: center;
        }

        form {
            margin-bottom: 20px;
            text-align: center;
        }

        label {
            font-weight: bold;
            margin-right: 10px;
        }

        input[type="text"] {
            width: calc(70% - 10px);
            padding: 8px;
            border: 1px solid #ccc; // Light grey
            border-radius: 4px;
        }

        button[type="submit"] {
            padding: 8px 20px;
            border: none;
            background-color: #254726; // Green
            color: white;
            border-radius: 4px;
            cursor: pointer;
        }

        button[type="submit"]:hover {
            background-color: #346b36; // Darker green
        }

        #resultContainer {
            display: none;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }

        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }

        th {
            background-color: #f2f2f2; // Light grey
        }

        #sumSection {
            margin-top: 20px;
        }

        #sumTable th, #nutritionalInfo th {
            background-color: #254726; // Green
            color: white;
        }

        #sumTable td, #nutritionalInfo td {
            background-color: #f5f5f5; // Light grey
            color: #333; // Dark grey
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>NutriSpark</h1>
        <form id="queryForm">
            <label for="query">Enter Food Item:</label>
            <input type="text" id="query" required>
            <button type="submit">Get Nutrition Data</button>
        </form>
        <div id="resultContainer" style="display: none;">
            <h2>Results</h2>
            <table id="nutritionalInfo">
                <!-- Table headers and data will be dynamically added here -->
            </table>
        </div>
        <div id="sumSection">
            <h2>Sum of Each Column</h2>
            <table id="sumTable">
                <!-- Table headers and sum data will be dynamically added here -->
            </table>
        </div>
    </div>
    <script>
        var previousRequests = [];

        document.getElementById('queryForm').addEventListener('submit', function(event) {
            event.preventDefault();
            var query = document.getElementById('query').value;
            fetch("/nutrition?query=" + encodeURIComponent(query))
            .then(response => response.json())
            .then(data => {
                previousRequests.unshift(data); // Add new request to the beginning of array
                renderTable(previousRequests);
                renderSum(previousRequests);
                document.getElementById('resultContainer').style.display = 'block';
            })
            .catch(err => {
                console.error('Error fetching nutrition data:', err);
                document.getElementById('nutritionalInfo').innerHTML = "<tr><td>Error fetching nutrition data. Please try again.</td></tr>";
                document.getElementById('resultContainer').style.display = 'block';
            });
        });

        function renderTable(data) {
            var table = document.getElementById('nutritionalInfo');
            table.innerHTML = ''; // Clear previous content

            // Add table headers
            var headerRow = table.insertRow();
            for (var key in data[0][0]) { // Assuming data is an array of arrays of objects
                var th = document.createElement('th');
                th.textContent = key;
                headerRow.appendChild(th);
            }

            // Add table data
            data.forEach(request => {
                request.forEach(item => {
                    var dataRow = table.insertRow();
                    for (var key in item) {
                        var td = document.createElement('td');
                        td.textContent = item[key];
                        dataRow.appendChild(td);
                    }
                });
            });
        }

        function renderSum(data) {
    var sumTable = document.getElementById('sumTable');
    sumTable.innerHTML = ''; // Clear previous content

    // Calculate sum of each column
    var sums = {};
    data.forEach(request => {
        request.forEach(item => {
            for (var key in item) {
                // Check if the value is a number
                if (!isNaN(parseFloat(item[key]))) {
                    sums[key] = (sums[key] || 0) + parseFloat(item[key]);
                } else {
                    sums[key] = "Total";
                }
            }
        });
    });

    // Add table headers
    var headerRow = sumTable.insertRow();
    for (var key in sums) {
        var th = document.createElement('th');
        th.textContent = key;
        headerRow.appendChild(th);
    }

    // Add sum data
    var sumRow = sumTable.insertRow();
    for (var key in sums) {
        var td = document.createElement('td');
        // Check if the value is a number
        if (!isNaN(parseFloat(sums[key]))) {
            var roundedValue = parseFloat(sums[key]).toFixed(3); // Round to the thousandth place
            td.textContent = roundedValue;
        } else {
            td.textContent = sums[key];
        }
        sumRow.appendChild(td);
    }
}

    </script>
</body>
</html>

//main.go
package main

import (
	"context"
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

	// Handle requests to the root URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve the HTML page
		http.ServeFile(w, r, "index.html")
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

	// Start the server
	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

*/