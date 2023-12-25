package main

import "net/http"

func updateCounter(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/counter/`, updateCounter)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
