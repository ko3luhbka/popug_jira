package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"

	"github.com/ko3luhbka/task_tracker/rest/model"

)

const (
	authServerURL = "http://localhost:8080/oauth"
	roleScope = "role"
	adminRole = "admin"
	mgrRole = "manager"
)

var conf = &oauth2.Config{
	ClientID:     "000000",
	ClientSecret: "999999",
	// RedirectURL: "http://localhost:8081",
	Scopes:       []string{roleScope},
	Endpoint: oauth2.Endpoint{
		AuthURL:  authServerURL + "/authorization-grant",
		TokenURL: authServerURL + "/get-token",
	},
}

func adminOnly(c *fiber.Ctx) error {
	return oauth(c, adminRole)
}

func managerOnly(c *fiber.Ctx) error {
	return oauth(c, mgrRole)
}

func oauth(c *fiber.Ctx, requiredRole string) error {
	token := c.Get("X-Auth-Token")
	if token == "" {
		code := c.Query("code")
		if code == "" {
			conf.RedirectURL = "http://localhost:8081" + c.Path()
			// Redirect user to consent page to ask for permission
			// for the scopes specified above.
			url := conf.AuthCodeURL("state", oauth2.AccessTypeOnline)
			fmt.Printf("Visit the URL for the auth dialog: %v", url)
			return c.Redirect(url)
		}

		token, err := conf.Exchange(c.Context(), code)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
	
		fmt.Println(token)
		return c.Next()
	}

	agent := fiber.AcquireAgent()
	req := agent.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI(authServerURL + "/validate-token")

	agent.BodyString(token)
	if err := agent.Parse(); err != nil {
		log.Printf("failed to parse request to auth server: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	code, body, err := agent.Bytes()
	if err != nil {
		log.Printf("failed to send token validation request: %v", err)
		errStr := make([]string, len(err))
		for i, e := range err {
			errStr[i] = e.Error()
		}
		errs := strings.Join(errStr, ",")
		return c.Status(fiber.StatusInternalServerError).SendString(errs)
	}
	if code == fiber.StatusOK {
		var tc model.TokenClaims
		if err := json.Unmarshal(body, &tc); err != nil {
			log.Printf("failed to parse auth server response body: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if tc.Role != requiredRole {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.Next()
	}
	log.Printf("auth server returned %d code, expected HTTP 200", code)
	return c.SendStatus(fiber.StatusUnauthorized)
}
