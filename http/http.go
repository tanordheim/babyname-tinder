package http

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/tanordheim/babyname-tinder"
	"golang.org/x/oauth2"
)

var (
	auth0Domain = os.Getenv("AUTH0_DOMAIN")
)

var (
	babyImages = []string{
		"1up-gopher.png",
		"bowtie-gopher.png",
		"bunny-gopher.png",
		"caffeine-gopher.png",
		"cranky-gopher.png",
		"flower-gopher.png",
		"gamer-gopher.png",
		"hearts-gopher.png",
		"latte-gopher.png",
		"lollipop-gopher.png",
		"mohawk-gopher.png",
		"party-gopher.png",
		"pirate-gopher.png",
		"pirikiikari-gopher.png",
		"popcorn-gopher.png",
		"princess-gopher.png",
		"smrt-gopher.png",
		"soda-gopher.png",
		"supermohawk-gopher.png",
		"superparty-gopher.png",
		"superunicorn-gopher.png",
		"sushi-gopher.png",
		"unicorn-gopher.png",
		"viking-gopher.png",
	}
)

func init() {
	gob.Register(&user{})
}

// NewServer creates a new HTTP server.
func NewServer(
	port int,
	siteURL string,
	cookieSecret string,
	oauthClientID string,
	oauthClientSecret string,
	repo babynames.Repository,
) *Server {

	sessionStore := sessions.NewCookieStore([]byte(cookieSecret))
	router := mux.NewRouter()
	oauthConfig := newOAuthConfig(siteURL, oauthClientID, oauthClientSecret)

	// Static content
	cwd, err := os.Getwd()
	if err != nil {
		panic(errors.Wrap(err, "Unable to get working directory"))
	}
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(path.Join(cwd, "static")))))

	// Authentication routes
	router.Handle("/login", newLoginHandler(oauthConfig, sessionStore)).Methods("GET")
	router.Handle("/callback", newCallbackHandler(oauthConfig, sessionStore)).Methods("GET")

	// App routes
	router.Handle("/", withAuth(sessionStore, newQueueHandler(repo))).Methods("GET")
	router.Handle("/like", withAuth(sessionStore, newLikeHandler(repo))).Methods("POST")
	router.Handle("/like/undo", withAuth(sessionStore, newUndoLikeHandler(repo))).Methods("POST")
	router.Handle("/superlike", withAuth(sessionStore, newSuperlikeHandler(repo))).Methods("POST")
	router.Handle("/dislike", withAuth(sessionStore, newDislikeHandler(repo))).Methods("POST")
	router.Handle("/dislike/undo", withAuth(sessionStore, newUndoDislikeHandler(repo))).Methods("POST")
	router.Handle("/liked", withAuth(sessionStore, newLikedHandler(repo))).Methods("GET")
	router.Handle("/liked/export_csv", withAuth(sessionStore, newExportLikedHandler(repo))).Methods("GET")
	router.Handle("/disliked", withAuth(sessionStore, newDislikedHandler(repo))).Methods("GET")
	router.Handle("/disliked/export_csv", withAuth(sessionStore, newExportDislikedHandler(repo))).Methods("GET")
	router.Handle("/matches", withAuth(sessionStore, newMatchesHandler(repo))).Methods("GET")
	router.Handle("/matches/export_csv", withAuth(sessionStore, newExportMatchesHandler(repo))).Methods("GET")
	router.Handle("/stats", withAuth(sessionStore, newStatsHandler(repo))).Methods("GET")

	// Admin routes
	router.Handle("/import", withAuth(sessionStore, newImportFormHandler())).Methods("GET")
	router.Handle("/import", withAuth(sessionStore, newImportHandler(repo))).Methods("POST")

	return &Server{
		port:   port,
		router: router,
	}

}

func withAuth(session sessions.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := session.Get(r, "auth")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := sess.Values["user"]; !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			ctx := setCurrentUser(r.Context(), sess.Values["user"].(*user))
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func newOAuthConfig(siteURL, clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("%s/callback", siteURL),
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s/authorize", auth0Domain),
			TokenURL: fmt.Sprintf("https://%s/oauth/token", auth0Domain),
		},
	}
}

func getRoleForEmail(email string) (babynames.Role, error) {
	switch email {
	case os.Getenv("DAD_EMAIL"):
		return babynames.DadRole, nil
	case os.Getenv("MOM_EMAIL"):
		return babynames.MomRole, nil
	default:
		return babynames.MomRole, fmt.Errorf("unknown email address")
	}
}

func parseTemplate(name string) *template.Template {
	template, err := template.ParseFiles("templates/layout.html", fmt.Sprintf("templates/%s.html", name))
	if err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("Unable to parse template '%s'", name)))
	}
	return template
}

func renderTemplate(w http.ResponseWriter, template *template.Template, data interface{}) {
	template.ExecuteTemplate(w, "layout", data)
}

func getRandomImage() string {
	img := babyImages[rand.Intn(len(babyImages))]
	return img
}
