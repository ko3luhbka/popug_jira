package rest

import (
	// "github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

var conf = &oauth2.Config{
	ClientID:     "000000",
	ClientSecret: "999999",
	Scopes:       []string{"SCOPE1", "SCOPE2"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "http://localhost:8080/authorize",
		TokenURL: "http://localhost:8080/token",
	},
}

// func simpleAuth(c *fiber.Ctx) error {

// }

// 	// Redirect user to consent page to ask for permission
// 	// for the scopes specified above.
// 	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
// 	fmt.Printf("Visit the URL for the auth dialog: %v", url)

// 	// Use the authorization code that is pushed to the redirect
// 	// URL. Exchange will do the handshake to retrieve the
// 	// initial access token. The HTTP Client returned by
// 	// conf.Client will refresh the token as necessary.
// 	var code string
// 	if _, err := fmt.Scan(&code); err != nil {
// 		log.Fatal(err)
// 	}
// 	tok, err := conf.Exchange(ctx, code)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	client := conf.Client(ctx, tok)
// 	client.Get("...")
