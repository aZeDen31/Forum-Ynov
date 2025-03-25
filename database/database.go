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
	Desc  string
}

type Post struct {
	ID      int
	Text    string
	Like    int
	Dislike int
	Image   []byte
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
		mdp VARCHAR(40) NOT NULL,
		desc TEXT
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
		dislike INT NOT NULL,
		image BLOB,
		FOREIGN KEY(utilisateur_id) REFERENCES utilisateurs(id)
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
	rows, err := db.Query("SELECT id, nom, email, mdp, desc FROM utilisateurs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var utilisateurs []Utilisateur

	for rows.Next() {
		var u Utilisateur
		err = rows.Scan(&u.ID, &u.Nom, &u.Email, &u.Mdp, &u.Desc)
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
func FindUserByNom(db *sql.DB, nom string) (Utilisateur, error) {
	var utilisateur Utilisateur

	row := db.QueryRow("SELECT id, nom, email, mdp, desc FROM utilisateurs WHERE nom = ?", nom)

	err := row.Scan(&utilisateur.ID, &utilisateur.Nom, &utilisateur.Email, &utilisateur.Mdp, &utilisateur.Desc)

	return utilisateur, err
}

func UpdateDesc(db *sql.DB, userID int, desc string) error {
	query := "UPDATE utilisateurs SET desc = ? WHERE id = ?"
	_, err := db.Exec(query, desc, userID)
	return err
}

// Insertpost insère un nouvel message dans la table "post".
func InsertPost(db *sql.DB, utilisateurID int, text string) error {
	query := "INSERT INTO posts (utilisateur_id, text) VALUES (?, ?)"
	_, err := db.Exec(query, utilisateurID, text)
	return err
}

// Insertlike insère un nouveau like dans le table "like".
func Insertlike(db *sql.DB, id int) error {
	query := "UPDATE posts SET like = like + 1 WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}

// Insertdislike insère un nouveau dislike dans le table "dislike".
func Insertdislike(db *sql.DB, id int) error {
	query := "UPDATE posts SET dislike = dislike + 1 WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}

// InsertPostWithImage insère une image dans la table "post".
func InsertPostWithImage(db *sql.DB, utilisateurID int, imageData []byte) error {
	query := "INSERT INTO posts (utilisateur_id, image) VALUES (?, ?, ?)"
	_, err := db.Exec(query, utilisateurID, imageData)
	return err
}

// LecturePost récupère et affiche les messages de la table "post" ainsi que les likes et dislikes et l'image du poste
//
//	( l'image doit etre convertie en base64 pour l'afficher sur le site web).
func LecturePost(db *sql.DB) ([]Post, error) {
	rows, err := db.Query("SELECT id, text, like_count, dislike_count, image FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post
		err = rows.Scan(&p.ID, &p.Text, &p.Like, &p.Dislike, &p.Image)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func GetId(db *sql.DB, name string) int {

	query := "SELECT id FROM utilisateurs WHERE nom = ?"
	row := db.QueryRow(query, name)
	var id int
	row.Scan(&id)
	return id
}
