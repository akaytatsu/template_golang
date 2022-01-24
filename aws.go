package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
)

var AWSConfig aws.Config = aws.Config{}

func getAWSConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func listBuckets(w http.ResponseWriter, r *http.Request) {
	svc := s3.NewFromConfig(AWSConfig)

	result, err := svc.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func YourHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func main() {

	AWSConfig = getAWSConfig()

	r := mux.NewRouter()

	// Routes consist of a path and a handler function.
	r.HandleFunc("/asd", YourHandler)
	r.HandleFunc("/l", listBuckets)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8080", r))
}
