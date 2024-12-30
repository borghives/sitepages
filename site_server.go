package sitepages

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/borghives/websession"
)

func GetAllowedHosts() map[string]bool {
	// Determine allowed hosts for HTTP service.
	var allowedHosts = map[string]bool{}
	envAllowHosts := os.Getenv("ALLOW_HOSTS")
	for _, host := range strings.Split(envAllowHosts, " ") {
		allowedHosts[host] = true
		log.Printf("Has allow host: %s", host)
	}
	return allowedHosts
}

func RunListenAndServer(handler http.Handler) {
	log.Print("starting server...")

	hostInfo := GetHostInfo()
	log.Printf("START New Host Instance@%s Build:%s Image:%s ", hostInfo.Id, hostInfo.BuildId, hostInfo.ImageId)

	// Initialize the session logic.  Will exit with fatal if cannot securely generate session
	websession.SessionInitCheck()

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
