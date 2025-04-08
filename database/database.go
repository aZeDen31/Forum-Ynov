package database

import (
	"database/sql"
	"encoding/base64"
	"fmt"

	"strings"

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
	Posts []Post
}

type Post struct {
	ID         int
	Titre      string
	Text       string
	Thread     string
	Like       int
	Dislike    int
	Image      []byte
	AuthorName string //j'ai ajouter des champs chef il en manquait
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

	//je vérifie la si les tables que j'ai ajouter elles existent sinon je  les adds
	_, err = db.Exec("PRAGMA table_info(posts)")
	if err != nil {
		return nil, err
	}

	db.Exec("ALTER TABLE posts ADD COLUMN titre VARCHAR(256)")

	db.Exec("ALTER TABLE posts ADD COLUMN thread VARCHAR(50)")

	db.Exec("ALTER TABLE posts ADD COLUMN liked_by TEXT DEFAULT ''")

	db.Exec("ALTER TABLE posts ADD COLUMN disliked_by TEXT DEFAULT ''")

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
/*func LectureUtilisateurs(db *sql.DB) ([]Utilisateur, error) {
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
}*/

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
func FindUserByNom(db *sql.DB, nom string) (Utilisateur, int) {
	var utilisateur Utilisateur

	row := db.QueryRow("SELECT id, nom, email, mdp, desc FROM utilisateurs WHERE nom = ?", nom)
	err := row.Scan(&utilisateur.ID, &utilisateur.Nom, &utilisateur.Email, &utilisateur.Mdp, &utilisateur.Desc)

	if err != nil {
		if err == sql.ErrNoRows {
			// Aucun utilisateur trouvé avec ce nom
			return utilisateur, 1
		}
		// Une autre erreur s'est produite
		return utilisateur, 2
	}

	// Utilisateur trouvé
	return utilisateur, 0
}

func UpdateUserInfo(db *sql.DB, id int, desc string, name string) error {
	var query string

	if name == "" && desc != "" {
		query = "UPDATE utilisateurs SET desc = ? WHERE id = ?"

		_, err := db.Exec(query, desc, id)
		return err
	}
	if name != "" && desc == "" {
		query = "UPDATE utilisateurs SET nom = ? WHERE id = ?"

		_, err := db.Exec(query, name, id)
		return err
	}
	if name != "" && desc != "" {
		query = "UPDATE utilisateurs SET nom = ?, desc = ? WHERE id = ?"

		_, err := db.Exec(query, name, desc, id)
		return err
	}
	return nil
}

// Insertpost insère un nouvel message dans la table "post".
func InsertPost(db *sql.DB, utilisateurID int, titre string, text string, thread string) error {
	query := "INSERT INTO posts (utilisateur_id, titre, text, thread, like, dislike) VALUES (?, ?, ?, ?, 0, 0)"
	_, err := db.Exec(query, utilisateurID, titre, text, thread)
	return err
}

// Insertlike inserts a like for the given post from the given user.
// If the user has already liked the post, remove the like.
func Insertlike(db *sql.DB, postID int, username string) error {
	var likedBy string
	if err := db.QueryRow("SELECT liked_by FROM posts WHERE id = ?", postID).Scan(&likedBy); err != nil {
		return err
	}
	if strings.Contains(likedBy, username) {
		updated := strings.Trim(strings.ReplaceAll(","+likedBy+",", ","+username+",", ","), ",")
		_, err := db.Exec("UPDATE posts SET like = like - 1, liked_by = ? WHERE id = ?", updated, postID)
		return err
	}
	_, err := db.Exec(`
		UPDATE posts 
		SET like = like + 1,
			liked_by = CASE 
				WHEN liked_by = '' THEN ? 
				ELSE liked_by || ',' || ? 
			END 
		WHERE id = ?`, username, username, postID)
	return err
}

// Insertdislike inserts a dislike for the given post from the given user.
// If the user has already disliked the post, remove the dislike.
func Insertdislike(db *sql.DB, postID int, username string) error {
	var dislikedBy string
	if err := db.QueryRow("SELECT disliked_by FROM posts WHERE id = ?", postID).Scan(&dislikedBy); err != nil {
		return err
	}
	if strings.Contains(dislikedBy, username) {
		updated := strings.Trim(strings.ReplaceAll(","+dislikedBy+",", ","+username+",", ","), ",")
		_, err := db.Exec("UPDATE posts SET dislike = dislike - 1, disliked_by = ? WHERE id = ?", updated, postID)
		return err
	}
	_, err := db.Exec(`
		UPDATE posts 
		SET dislike = dislike + 1,
			disliked_by = CASE 
				WHEN disliked_by = '' THEN ? 
				ELSE disliked_by || ',' || ? 
			END 
		WHERE id = ?`, username, username, postID)
	return err
}

// InsertPostWithImage insère une image dans la table "post".
func InsertPostWithImage(db *sql.DB, utilisateurID int, imageData []byte) error {
	query := "INSERT INTO posts (utilisateur_id, image) VALUES (?, ?, ?)"
	_, err := db.Exec(query, utilisateurID, imageData)
	return err
}

func LecturePost(db *sql.DB) ([]Post, error) {
	query := `
    SELECT p.id, p.titre, p.text, p.thread, p.like, p.dislike, p.image, u.nom 
    FROM posts p
    JOIN utilisateurs u ON p.utilisateur_id = u.id
    ORDER BY p.id DESC
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post
		err = rows.Scan(&p.ID, &p.Titre, &p.Text, &p.Thread, &p.Like, &p.Dislike, &p.Image, &p.AuthorName)
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

func GetUserDescription(db *sql.DB, username string) (string, error) {
	var description string

	query := `SELECT desc FROM utilisateurs WHERE nom = ?`
	err := db.QueryRow(query, username).Scan(&description)
	if err != nil {
		return "", err
	}
	return description, nil
}

// la fonction c pour tes images de con en b64
func ImageToBase64(imageData []byte) string {
	return base64.StdEncoding.EncodeToString(imageData)
}

func LecturePostThread(thread string, db *sql.DB) ([]Post, error) {
	query := `
        SELECT p.id, p.titre, p.text, p.thread, p.like, p.dislike, p.image, u.nom 
        FROM posts p
        JOIN utilisateurs u ON p.utilisateur_id = u.id
        WHERE p.thread = ?
        ORDER BY p.id DESC
    `

	rows, err := db.Query(query, thread)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post
		err = rows.Scan(&p.ID, &p.Titre, &p.Text, &p.Thread, &p.Like, &p.Dislike, &p.Image, &p.AuthorName)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func LecturePostAuthor(Author string, db *sql.DB) ([]Post, error) {
	query := `
        SELECT p.id, p.titre, p.text, p.thread, p.like, p.dislike, p.image, u.nom 
        FROM posts p
        JOIN utilisateurs u ON p.utilisateur_id = u.id
        WHERE u.nom = ?
        ORDER BY p.id DESC
    `

	rows, err := db.Query(query, Author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post
		err = rows.Scan(&p.ID, &p.Titre, &p.Text, &p.Thread, &p.Like, &p.Dislike, &p.Image, &p.AuthorName)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}
