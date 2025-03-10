package main

import (
	"fmt"
	"forum-ynov/database"
	"html/template"
	"log"
	"net/http"
)

func main() {
	initdatabase()
	server()
}

func server() {
	fileServer := http.FileServer(http.Dir("./HTML"))

	fs := http.FileServer(http.Dir("./CSS"))
	http.Handle("/CSS/", http.StripPrefix("/CSS/", fs))

	fd := http.FileServer(http.Dir("./img"))
	http.Handle("/img/", http.StripPrefix("/img/", fd))

	http.Handle("/", fileServer)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/profile", profileHandler)

	fmt.Println("clique sur le lien http://localhost:5500/")
	if err := http.ListenAndServe(":5500", nil); err != nil {
		panic(err)
	}

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("HTML/connexion.html"))
	tmpl.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("HTML/inscription.html"))
	tmpl.Execute(w, nil)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("HTML/profile.html"))
	tmpl.Execute(w, nil)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		mail := r.FormValue("mail") //note a moi même FormValue récupere la categorie name
		password := r.FormValue("mdp")
		password2 := r.FormValue("mdp2")

		if password != password2 {
			fmt.Fprintf(w, "Not the same password ")
		} else {
			// faut insérer ici la query vers la db
			// penser a check si user existe déjà
			//si non l'add a la db
			fmt.Printf("Username: %s, Password: %s\n", mail, password)
			if true {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			} else {
				fmt.Fprintf(w, "Error user is already registered")
			}

		}
	}

	tmpl, err := template.ParseFiles("html/inscription.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

func LoginPage(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		//Check si les identifiants sont corrects
		fmt.Printf("Username: %s, Password: %s\n", username, password)
		if true {
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		} else {
			fmt.Fprintf(w, "Error user is not registered")
		}
	}
	tmpl, err := template.ParseFiles("HTML/connexion.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func initdatabase(){
	// Initialisation de la base de données
	db, err := database.InitDB("ma_base.db")
	if err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données :", err)
	}
	defer db.Close()

	fmt.Println("Base de données initialisée et prête à être utilisée !")

	// Exemple d'ajout d'un utilisateur
	err = database.InsertUser(db, "Alice", "caca3@example.com", "motdepasse123")
	if err != nil {
		log.Fatal("Erreur lors de l'ajout de l'utilisateur :", err)
	}

	// Lecture des utilisateurs
	utilisateurs, err := database.LectureUtilisateurs(db)
	if err != nil {
		log.Fatal("Erreur lors de la récupération des utilisateurs:", err)
	}

	for _, u := range utilisateurs {
		fmt.Printf("ID: %d | Nom: %s | Email: %s | MDP: %s\n", u.ID, u.Nom, u.Email, u.Mdp)
	}
}

