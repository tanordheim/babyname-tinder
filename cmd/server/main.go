package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/tanordheim/babyname-tinder/http"
	"github.com/tanordheim/babyname-tinder/psql"
)

func getConfig(name string) string {
	val := os.Getenv(name)
	if val == "" {
		panic(fmt.Errorf("Environment variable %s is not set", name))
	}

	return val
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("Unable to parse port number '%s'", port)))
	}
	siteURL := getConfig("SITE_URL")
	cookieSecret := getConfig("COOKIE_SECRET")
	oauthClientID := getConfig("OAUTH_CLIENT_ID")
	oauthClientSecret := getConfig("OAUTH_CLIENT_SECRET")

	repo := psql.NewRepository(os.Getenv("DATABASE_URL"))
	server := http.NewServer(
		portNumber,
		siteURL,
		cookieSecret,
		oauthClientID,
		oauthClientSecret,
		repo,
	)
	server.Run()
}
