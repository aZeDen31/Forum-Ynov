package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Utilisateur représente un utilisateur de la base de données.
type Utilisateur struct {
	ID    int
	Nom   string
	Email string
	Mdp   string
}

type Post struct {
	ID      int
	Text    string
	like    int
	dislike int
}

// InitDB initialise la base de données et crée la table "utilisateurs" si elle n'existe pas.
func InitDB(dbPath string) (*sql.DB, error) {
	// Connexion à la base de données
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Création de la table
	createTable := `
	CREATE TABLE IF NOT EXISTS utilisateurs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nom TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		mdp VARCHAR(40) NOT NULL
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	//Creation de la table pour les posts
	createTablepost := `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		utilisateur_id INTEGER NOT NULL,
		"text" VARCHAR(256) NOT NULL,
		like INT NOT NULL,
		dislke INT NOT NULL,
		FOREIGN KEY(utilisateur_id) REFERENCES sociétés(id)
	);`
	_, err = db.Exec(createTablepost)
	if err != nil {
		return nil, err
	}

	// Vérifier la connexion
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Base de données initialisée avec succès.")
	return db, nil
}

// InsertUser insère un nouvel utilisateur dans la table "utilisateurs".
func InsertUser(db *sql.DB, nom, email, mdp string) error {
	// Hachage du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(mdp), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erreur lors du hachage du mot de passe: %v", err)
	}

	// Insertion de l'utilisateur avec le mot de passe haché
	query := "INSERT INTO utilisateurs (nom, email, mdp) VALUES (?, ?, ?)"
	_, err = db.Exec(query, nom, email, string(hashedPassword))
	return err
}

// LectureUtilisateurs récupère et affiche les utilisateurs de la table "utilisateurs".
func LectureUtilisateurs(db *sql.DB) ([]Utilisateur, error) {
	rows, err := db.Query("SELECT id, nom, email, mdp FROM utilisateurs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var utilisateurs []Utilisateur

	for rows.Next() {
		var u Utilisateur
		err = rows.Scan(&u.ID, &u.Nom, &u.Email, &u.Mdp)
		if err != nil {
			return nil, err
		}
		utilisateurs = append(utilisateurs, u)
	}

	return utilisateurs, nil
}

func FindUser(db *sql.DB, email string, mdp string) (Utilisateur, int) { //LOGIN

	var utilisateur Utilisateur

	//A noter peux pas utiliser db.Exec si il y'a une value de return
	row := db.QueryRow("SELECT id, nom, email, mdp FROM utilisateurs WHERE email = ?", email)

	err := row.Scan(&utilisateur.ID, &utilisateur.Nom, &utilisateur.Email, &utilisateur.Mdp)

	if err != nil {
		if err == sql.ErrNoRows { //erreur serveur

			return utilisateur, 1
		}
		return utilisateur, 3
	}

	err = bcrypt.CompareHashAndPassword([]byte(utilisateur.Mdp), []byte(mdp))
	if err != nil {
		return utilisateur, 2 // Mot de passe incorrect
	}

	return utilisateur, 0
}

// Insertpost insère un nouvel message dans la table "post".
func Insertpost(db *sql.DB, text string) error {

	query := "INSERT INTO post (text) VALUES (?)"
	_, err := db.Exec(query, text)
	return err
}
func Insertlike(db *sql.DB, like int, id int) error {
	query := "UPDATE posts SET like = like + ? WHERE id = ?"
	_, err := db.Exec(query,like, id)
	return err
}
func Insertdislike(db *sql.DB, dislike int, id int) error {
	query := "UPDATE posts SET dislike = dislike + ? WHERE id = ?"
	_, err := db.Exec(query,dislike, id)
	return err
}

// LecturePost récupère et affiche les messages de la table "post".
func LecturePost(db *sql.DB) ([]Post, error) {
	rows, err := db.Query("SELECT id, text, like, dislike FROM post")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post
		err = rows.Scan(&p.ID, &p.Text, &p.like, &p.dislike)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}
