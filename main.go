package main

import (
	"encoding/base64"
	"net/http"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Session struct {
	Username string
	Password string
}

func main() {
	sessions := make(map[string]Session)

	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)

	// Serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/LoginForm.html")
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		// Check if the submitted credentials are valid
		if r.FormValue("username") == "test" && r.FormValue("password") == "test" {
			// Create a new session and add it to the sessions map
			session := Session{
				Username: r.FormValue("username"),
				Password: r.FormValue("password"),
			}
			sessionID := base64.StdEncoding.EncodeToString([]byte(session.Username + ":" + session.Password))
			sessions[sessionID] = session

			// Set the session ID as a cookie on the response
			http.SetCookie(w, &http.Cookie{
				Name:  "sessionID",
				Value: sessionID,
				Path:  "/",
			})

			// Redirect to the home page
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)   // returns 401 Unauthorized telling the user they are not authorized to access the requested resource (in this case, the login-protected page) due to invalid credentials.
			w.Write([]byte("invalid login details")) // returns an error message
		}
	})

	r.Get("/home", func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is authenticated by checking if the session ID is present in the cookies
		sessionID, err := r.Cookie("sessionID")
		if err != nil || sessionID.Value == "" || sessions[sessionID.Value].Username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get the username from the session and render the home page
		username := sessions[sessionID.Value].Username
		tmpl, err := template.ParseFiles("static/HomePage.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, struct {
			Username string
		}{
			Username: username,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		// Delete the session cookie from the user's browser
		http.SetCookie(w, &http.Cookie{
			Name:   "sessionID",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})

		// Remove the session from the sessions map
		sessionID, err := r.Cookie("sessionID")
		if err == nil {
			delete(sessions, sessionID.Value)
		}

		// Redirect the user back to the login page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Start the server
	http.ListenAndServe(":8080", r)
}
