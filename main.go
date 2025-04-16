package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"forum-ynov/database"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// variable global sinon c chiant
var DB *sql.DB
var err error

type UserData struct {
	Username    string
	Error       string
	ID          int
	Email       string
	Description string
	//faudra ajouter des trucs c ce que je passe a la template profile
}

var userdata UserData

func main() {
	os.MkdirAll("uploads", os.ModePerm) //On fait le fichier uploads
	// Initialisation de la base de données
	userdata.Username = "Non connecté"
	userdata.ID = 0 //si ID = 0 user non connécté
	DB, err = database.InitDB("ma_base.db")
	if err != nil {
		log.Fatal("Erreur d'initialisation de la base de données:", err)
	}
	defer DB.Close()

	// Lecture des utilisateurs
	/*utilisateurs, err := database.LectureUtilisateurs(DB)
	if err != nil {
		log.Fatal("Erreur lors de la récupération des utilisateurs:", err)
	}

	for _, u := range utilisateurs {
		fmt.Printf("ID: %d | Nom: %s | Email: %s | MDP: %s\n", u.ID, u.Nom, u.Email, u.Mdp)
	}
	*/
	server()
}

func server() {

	fs := http.FileServer(http.Dir("./CSS"))
	http.Handle("/CSS/", http.StripPrefix("/CSS/", fs))

	fd := http.FileServer(http.Dir("./img"))
	http.Handle("/img/", http.StripPrefix("/img/", fd))

	uploads := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", uploads))

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/profileModif", modifProfileHandler)
	http.HandleFunc("/login", LoginPage)
	http.HandleFunc("/register", Signup)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/threads", threadsHandler)
	http.HandleFunc("/createpost", createpostHandler)
	http.HandleFunc("/like/", LikeHandler)
	http.HandleFunc("/dislike/", DislikeHandler)
	http.HandleFunc("/upload", uploadHandler)

	fmt.Println("clique sur le lien http://localhost:5400/")
	if err := http.ListenAndServe(":5400", nil); err != nil {
		panic(err)
	}
}

func modifProfileHandler(w http.ResponseWriter, r *http.Request) {
	if checkCookie(w, r) != 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("HTML/profilModif.html"))

	tmpl.Execute(w, userdata)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si l'utilisateur est connecté
	if checkCookie(w, r) != 0 {

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Traiter la déconnexion et les modifications de profil
	if r.Method == "POST" {
		if r.FormValue("action") == "Deconnexion" {
			deleteCookie(w, r)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Parser le formulaire multipart pour pouvoir récupérer l'image
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil && !errors.Is(err, http.ErrNotMultipart) {
			log.Println("Erreur lors du parsing du formulaire:", err)
		}

		// Traitement des modifications de profil
		newPseudo := r.FormValue("Pseudo")
		newDescription := r.FormValue("Description")

		// Variables pour gérer l'image de profil
		var profileImagePath string
		var profileImageUpdated bool

		// Récupérer l'image de profil s'il y en a une
		file, handler, err := r.FormFile("profileImage")
		if err == nil {
			defer file.Close()

			// Créer le dossier profiles s'il n'existe pas
			os.MkdirAll("uploads/profiles", os.ModePerm)

			// Générer un nom unique pour l'image
			fileExt := filepath.Ext(handler.Filename)
			newFilename := fmt.Sprintf("%d_%s%s", userdata.ID, time.Now().Format("20060102150405"), fileExt)
			profileImagePath = filepath.Join("uploads/profiles", newFilename)

			dst, err := os.Create(profileImagePath)
			if err == nil {
				defer dst.Close()
				if _, err := io.Copy(dst, file); err == nil {
					profileImageUpdated = true
				} else {
					log.Println("Erreur lors de la copie de l'image:", err)
				}
			} else {
				log.Println("Erreur lors de la création du fichier:", err)
			}
		}

		// Vérifications pour le nouveau pseudo
		if newPseudo != "" && newPseudo != userdata.Username {
			_, verif := database.FindUserByNom(DB, newPseudo)
			if verif == 0 {
				userdata.Error = "Nom déjà utilisé"
				tmpl := template.Must(template.ParseFiles("HTML/profilModif.html"))
				tmpl.Execute(w, userdata)
				userdata.Error = ""
				return
			}
		}

		var imagePathToUpdate string
		if profileImageUpdated {
			imagePathToUpdate = profileImagePath
		} else {
			imagePathToUpdate = ""
		}

		err = database.UpdateUserInfo(DB, userdata.ID, newDescription, newPseudo, imagePathToUpdate)

		if err != nil {
			log.Println("Erreur lors de la mise à jour du profil:", err)
			http.Error(w, "Erreur lors de la mise à jour du profil", http.StatusInternalServerError)
			return
		}

		// Mettre à jour les données utilisateur si le pseudo a changé
		if newPseudo != "" {
			userdata.Username = newPseudo
			// Recréer le cookie avec le nouveau nom
			deleteCookie(w, r)
			user, _ := database.FindUserByNom(DB, newPseudo)
			setUserCookie(user, w, r)
		}

		// Mettre à jour la description localement
		if newDescription != "" {
			userdata.Description = newDescription
		}

		// Rediriger vers la page de profil mise à jour
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Récupérer la description et l'image de profil de l'utilisateur
	user, err := database.GetUserProfile(DB, userdata.Username)
	if err != nil {
		log.Println("Erreur lors de la récupération du profil:", err)
		// On continue avec des valeurs par défaut
		userdata.Description = ""
	} else {
		userdata.Description = user.Desc
		// On pourra ajouter l'image de profil aussi ici
	}

	// Récupérer les posts de l'utilisateur
	posts, err := database.LecturePostAuthor(userdata.Username, DB)
	if err != nil {
		log.Println("Erreur lors de la récupération des posts:", err)
		// On continue avec une liste vide
		posts = []database.Post{}
	}

	// Structure de données pour le template
	Data := struct {
		Username     string
		Email        string
		Description  string
		ProfileImage string
		Posts        []database.Post
	}{
		Username:     userdata.Username,
		Email:        userdata.Email,
		Description:  userdata.Description,
		ProfileImage: user.ProfileImage, // Nouveau champ
		Posts:        posts,
	}

	// Fonctions personnalisées pour le template
	funcMap := template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
	}

	// Créer et exécuter le template
	tmpl, err := template.New("profile.html").Funcs(funcMap).ParseFiles("HTML/profile.html")
	if err != nil {
		log.Println("Erreur lors de l'analyse du template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, Data)
	if err != nil {
		log.Println("Erreur lors de l'exécution du template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	checkCookie(w, r)

	posts, err := database.LecturePost(DB)
	if err != nil {
		log.Println("Erreur lors de la récupération des posts:", err)
	}

	// j'envoi ça dans index
	Data := struct {
		Username string
		Posts    []database.Post
	}{
		Username: userdata.Username,
		Posts:    posts,
	}

	// Créer un FuncMap avec la fonction subtract
	funcMap := template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
	}

	// Créer le template avec les fonctions personnalisées
	tmpl := template.New("index.html").Funcs(funcMap)

	// Analyser le fichier de template
	tmpl, err = tmpl.ParseFiles("HTML/index.html")
	if err != nil {
		log.Println("Erreur lors de l'analyse du template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	// Exécuter le template
	tmpl.Execute(w, Data)
}

func threadsHandler(w http.ResponseWriter, r *http.Request) {
	checkCookie(w, r)
	threadName := r.URL.Query().Get("name")

	posts, err := database.LecturePostThread(threadName, DB)
	if err != nil {
		log.Println("Erreur lors de la récupération des posts:", err)
	}

	// j'envoi ça dans index
	data := struct {
		Username string
		Thread   string
		Posts    []database.Post
	}{
		Username: userdata.Username,
		Thread:   posts[0].Thread,
		Posts:    posts,
	}

	// Créer un FuncMap avec la fonction subtract
	funcMap := template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
	}

	// Créer le template avec les fonctions personnalisées
	tmpl := template.New("thread.html").Funcs(funcMap)

	// Analyser le fichier de template
	tmpl, err = tmpl.ParseFiles("HTML/thread.html")
	if err != nil {
		log.Println("Erreur lors de l'analyse du template:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	// Exécuter le template
	tmpl.Execute(w, data)
}

func createpostHandler(w http.ResponseWriter, r *http.Request) {
	if checkCookie(w, r) != 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		titre := r.FormValue("Titre")
		contenu := r.FormValue("Contenu")
		thread := r.FormValue("pets")

		// Vérifier que les champs requis sont remplis
		if titre == "" || contenu == "" || thread == "" {
			// Rediriger avec un message d'erreur
			http.Redirect(w, r, "/createpost?error=Tous les champs sont obligatoires", http.StatusSeeOther)
			return
		}

		err := database.InsertPost(DB, userdata.ID, titre, contenu, thread)
		if err != nil {
			log.Println("Erreur lors de l'insertion du post:", err)
			http.Redirect(w, r, "/createpost?error=Erreur lors de la création du post", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/threads?name="+thread, http.StatusSeeOther) // Rediriger vers la page des threads
		return
	}

	tmpl := template.Must(template.ParseFiles("HTML/createpost.html"))
	data := struct {
		Username string
		Error    string
	}{
		Username: userdata.Username,
		Error:    r.URL.Query().Get("error"),
	}
	tmpl.Execute(w, data)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Erreur lors du parsing du formulaire", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Fichier non trouvé", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(handler.Filename)
	savePath := filepath.Join("uploads", filename)

	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Erreur lors de la sauvegarde du fichier", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Erreur lors de la copie du fichier", http.StatusInternalServerError)
		return
	}

	err = saveImagePathToDB(savePath)
	if err != nil {
		http.Error(w, "Erreur BDD", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Image reçue et enregistrée !"))
}

func saveImagePathToDB(path string) error {
	_, err := DB.Exec(`INSERT INTO images (path) VALUES (?)`, path)
	return err
}

func Signup(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("HTML/inscription.html")

	if r.Method == "POST" {
		username := r.FormValue("username")
		mail := r.FormValue("mail")
		password := r.FormValue("mdp")
		password2 := r.FormValue("mdp2")

		if mail == "" || password == "" || password2 == "" || username == "" {
			userdata.Error = "Veuillez remplir tous les champs"
			tmpl.Execute(w, userdata)
			return
		}

		if password != password2 {
			userdata.Error = "Les mots de passe ne correspondent pas"
			tmpl.Execute(w, userdata)
		} else {

			err := database.InsertUser(DB, username, mail, password)
			if err != nil {
				userdata.Error = "Email / username déjà utilisé" //oui c pas safe mais ntm je fait pas une double auth
				tmpl.Execute(w, userdata)
			}
			userdata.Error = ""
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return

		}
	} else {
		// Si ce n'est pas une requête POST, afficher le formulaire
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl, _ := template.ParseFiles("HTML/inscription.html")
		tmpl.Execute(w, nil)
	}
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("HTML/connexion.html")
	if r.Method == "POST" {
		mail := r.FormValue("mail")
		password := r.FormValue("mdp")

		fmt.Printf("Username: %s, Password: %s\n", mail, password)

		if mail == "" || password == "" {
			userdata.Error = "Veuillez remplir tous les champs"
			tmpl.Execute(w, userdata)
			return
		}

		user, err := database.FindUser(DB, mail, password)
		if err != 0 {
			userdata.Error = "email / mdp invalide "
			tmpl.Execute(w, userdata)
			return
		}
		setUserCookie(user, w, r)
		userdata.Error = ""
		userdata.Email = mail
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, userdata)
}

func setUserCookie(user database.Utilisateur, w http.ResponseWriter, r *http.Request) { //je peux pas appeller la fonction setCOokie car setCokkie existe déjà

	cookie := http.Cookie{
		Name:     "user",
		Value:    user.Nom,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600, //ttl du cookie
	}
	fmt.Println("USERNAME DANS COOKIE:", user.Nom)
	Write64(w, cookie) //donne le cookie qu'on vient de faire au client
}

func deleteCookie(w http.ResponseWriter, r *http.Request) {
	c := &http.Cookie{
		Name:    "user",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, c)
}
func checkCookie(w http.ResponseWriter, r *http.Request) int {
	_, err := r.Cookie("user")

	if err != nil {
		userdata.Username = "Non connecté"
		userdata.Email = "" // Réinitialiser l'email aussi
		switch {
		case errors.Is(err, http.ErrNoCookie):
			return 1
		default:
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return 2
		}
	}

	username, _ := Read64(r, "user")
	userdata.Username = username
	userdata.ID = database.GetId(DB, userdata.Username)

	// Récupérer les informations complètes de l'utilisateur, y compris l'email
	user, status := database.FindUserByNom(DB, username)
	if status == 0 {
		userdata.Email = user.Email
	} else {
		// En cas d'erreur, on pourrait logger mais on continue
		log.Println("Impossible de récupérer l'email pour", username)
	}

	return 0
}

func Write64(w http.ResponseWriter, cookie http.Cookie) error { //transforme le cookie en b64
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	http.SetCookie(w, &cookie) //pointeur pour changer idrectement la variable car elle n'est pas global
	return nil
}

func Read64(r *http.Request, name string) (string, error) { //lit le cookie en b6
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

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if checkCookie(w, r) != 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	idStr := r.URL.Path[len("/like/"):]
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	username, err := Read64(r, "user")

	if err := database.Insertlike(DB, postID, username); err != nil {
		http.Error(w, "Erreur lors de l'ajout du like", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	if checkCookie(w, r) != 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	idStr := r.URL.Path[len("/dislike/"):]
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}
	username, err := Read64(r, "user")
	if err != nil {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}
	if err := database.Insertdislike(DB, postID, username); err != nil {
		http.Error(w, "Erreur lors de l'ajout du dislike", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
