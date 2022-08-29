package main

import (
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: './auth username|password repo|registry'")
	}
	if os.Args[1] != "username" && os.Args[1] != "password" {
		log.Fatal("field must be 'username' or 'password'")
	}
	field := os.Args[1]
	url := os.Args[2]

	// get registry hostname
	reg, err := name.NewRegistry(url)
	if err != nil {
		ref, err := name.ParseReference(url)
		if err == nil {
			log.Println("parsed as reference")
			reg = ref.Context().Registry
		} else {
			log.Fatal(err)
		}
	}
	log.Printf("registry: %s\n", reg.Name())

	// get credentials from DOCKER_CONFIG
	authenticator, err := authn.DefaultKeychain.Resolve(reg)
	if err != nil {
		log.Fatal(err)
	}
	authz, err := authenticator.Authorization()
	if err != nil {
		log.Fatal(err)
	}
	if authenticator == authn.Anonymous {
		log.Println("using anonymous authenticator")
	}

	if field == "username" {
		_, err := os.Stdout.Write([]byte(authz.Username))
		if err != nil {
			log.Fatal(err)
		}
	}
	if field == "password" {
		_, err := os.Stdout.Write([]byte(authz.Password))
		if err != nil {
			log.Fatal(err)
		}
	}
}
