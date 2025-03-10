package database

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

// Utilisateur représente un utilisateur de la base de données.
type Utilisateur struct {
	ID    int
	Nom   string
	Email string
	Mdp   string
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
		mdp TEXT NOT NULL
	);`
	_, err = db.Exec(createTable)
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
		return fmt.Errorf("Erreur lors du hachage du mot de passe: %v", err)
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

