package main

import (
	"flag"
	"log"
	"net/http"
)

// We will access the POST /v1/tokens/authentication endpoint
const html = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Appletree Preflight CORS</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			background-color: #f4f4f4;
			margin: 0;
			padding: 0;
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh;
		}
		.container {
			background: #fff;
			padding: 50px;
			border-radius: 8px;
			box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
			text-align: center;
		}
		.loading {
			display: none;
		}
		.error {
			color: red;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Appletree Preflight CORS</h1>
		<div id="output"></div>
		<div class="loading" id="loading">Loading...</div>
	</div>
	<script>
		document.addEventListener('DOMContentLoaded', function() {
			 output = document.getElementById("output");
			 loading = document.getElementById("loading");

			loading.style.display = "block";

			fetch("http://localhost:4000/v1/tokens/authentication", {
				method: "POST",
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					email: 'alice@example.com',
					password: 'newpassword'
				})
			}).then(function(response) {
				loading.style.display = "none";
				if (response.ok) {
					response.json().then(function(data) {
						output.innerHTML = "<strong>Authentication Token:</strong> <em>" + data.authentication_token.token + "</em>";
					});
				} else {
					response.text().then(function(text) {
						output.innerHTML = "<span class='error'>Error: " + text + "</span>";
					});
				}
			}).catch(function(err) {
				loading.style.display = "none";
				output.innerHTML = "<span class='error'>Error: " + err.message + "</span>";
			});
		});
	</script>
</body>
</html>
`

// A very simple HTTP server 
func main() {
	addr := flag.String("addr", ":9000", "Server address")
	flag.Parse()

	log.Printf("Starting server on %s", *addr)

	err := http.ListenAndServe(*addr, 
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		   w.Write([]byte(html))
	}))
 	
	log.Fatal(err)

}