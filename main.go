package main

import (
	"database/sql"
	"fmt"
	"forum-ynov/database"
	"html/template"
	"log"
	"net/http"
	"errors"
	"encoding/base64"
)

// variable global sinon c chiant
var DB *sql.DB
var err error

func main() {
	// Initialisation de la base de données
	DB, err = database.InitDB("ma_base.db")
	if err != nil {
		log.Fatal("Erreur d'initialisation de la base de données:", err)
	}
	defer DB.Close()

	// Lecture des utilisateurs
	utilisateurs, err := database.LectureUtilisateurs(DB)
	if err != nil {
		log.Fatal("Erreur lors de la récupération des utilisateurs:", err)
	}

	for _, u := range utilisateurs {
		fmt.Printf("ID: %d | Nom: %s | Email: %s | MDP: %s\n", u.ID, u.Nom, u.Email, u.Mdp)
	}

	server()
}

func server() {
	fileServer := http.FileServer(http.Dir("./HTML"))

	fs := http.FileServer(http.Dir("./CSS"))
	http.Handle("/CSS/", http.StripPrefix("/CSS/", fs))

	fd := http.FileServer(http.Dir("./img"))
	http.Handle("/img/", http.StripPrefix("/img/", fd))

	http.Handle("/", fileServer)
	http.HandleFunc("/login", LoginPage)
	http.HandleFunc("/register", Signup)
	http.HandleFunc("/profile", profileHandler)

	fmt.Println("clique sur le lien http://localhost:5500/")
	if err := http.ListenAndServe(":5500", nil); err != nil {
		panic(err)
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	if checkCookie(w, r) != 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	tmpl := template.Must(template.ParseFiles("HTML/profile.html"))
	tmpl.Execute(w, nil)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		mail := r.FormValue("mail")
		password := r.FormValue("mdp")
		password2 := r.FormValue("mdp2")

		if mail == "" || password == "" || password2 == "" || username == "" {
			fmt.Fprintf(w, "Please fill in all fields")
			return
		}

		if password != password2 {
			fmt.Fprintf(w, "Not the same password ")
		} else {

			err := database.InsertUser(DB, username, mail, password)
			if err != nil {
				fmt.Fprintf(w, "Error registering user: %v", err)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return

		}
	} else {
		// Si ce n'est pas une requête POST, afficher le formulaire
		tmpl, err := template.ParseFiles("HTML/inscription.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	}
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		mail := r.FormValue("mail")
		password := r.FormValue("mdp")


		fmt.Printf("Username: %s, Password: %s\n", mail, password)

		if mail == "" || password == "" {
			fmt.Fprintf(w, "Please fill in all fields")
			return
		}

		user, err := database.FindUser(DB, mail, password)
		if err != 0 {
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		setUserCookie(user,w ,r)
				

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("HTML/connexion.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func setUserCookie(user database.Utilisateur , w http.ResponseWriter, r *http.Request) { //je peux pas appeller la fonction setCOokie car setCokkie existe déjà

	cookie := http.Cookie{
		Name:     "user",
		Value:    user.Nom,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600, //ttl du cookie 
	}
		Write64(w, cookie) //donne le cookie qu'on vient de faire au client  
	
}

func checkCookie(w http.ResponseWriter, r *http.Request) (int){

	_, err := r.Cookie("user")

	if err != nil {
        switch {
        case errors.Is(err, http.ErrNoCookie):
            return 1
        default:
            log.Println(err)
            http.Error(w, "server error", http.StatusInternalServerError)
			return 2
        }
    }
		return 0
} 
	
func Write64(w http.ResponseWriter, cookie http.Cookie) error { //transforme le cookie en b64 
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	http.SetCookie(w, &cookie) //pointeur pour changer idrectement la variable car elle n'est pas global 
	return nil 
}

func Read64(r *http.Request, name string ) (string, error){ //lit le cookie en b6
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	decoded, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
	
}