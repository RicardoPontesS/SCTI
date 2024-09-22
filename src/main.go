package main

import (
    "SCTI/database"
    "SCTI/fileserver"
    "SCTI/middleware"
    "fmt"
    "log"
    "net/http"

    "github.com/joho/godotenv"
)

func main() {
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    err = database.OpenDatabase()
    if err != nil {
        log.Printf("Error connecting to postgres database\n%v", err)
    }
    defer database.CloseDatabase()

    fileserver.RunFileServer()
    mux := http.NewServeMux()
    LoadRoutes(mux)

    certFile := "/etc/letsencrypt/live/sctiuenf.com.br/fullchain.pem"
    keyFile := "/etc/letsencrypt/live/sctiuenf.com.br/privkey.pem"

    server := http.Server{
        Addr:    ":443",
        Handler: middleware.EndpointLogging(mux),
    }

    fmt.Println("Server started at: https://sctiuenf.com.br")
    
    log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}

