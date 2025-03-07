package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" 
)

func InitDB(dbPath string) (*sql.DB,error) {
	// Connexion à la base de données (créée si elle n'existe pas) pour que ca ajoute directement a ma_base.db toutes les valeurs crées de la table
	db, err := sql.Open("sqlite3", "ma_base.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Création d'une table , cela permet de créer une table avec id , nom , email et mdp
	createTable := `
	CREATE TABLE IF NOT EXISTS utilisateurs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nom TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		mdp TEXT UNIQUE NOT NULL
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }
    return db
}

func InsertUser(db *sql.DB, nom, mail, mdp string)error {

	// Insertion d'un utilisateur , permet donc de inserer les valeurs nom email et mdp
	query := "INSERT INTO users (name, email, password) VALUES (?, ?, ?)"
	_, err := db.Exec(query, name, email, password) // We are ignoring the Result returned
	return err
}

	// Lecture des utilisateurs
	rows, err := db.Query("SELECT id, nom, email, mdp FROM utilisateurs")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Utilisateurs :")
	for rows.Next() {
		var id int
		var nom, email, mdp string
		err = rows.Scan(&id, &nom, &email, &mdp)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d : %s (%s) mdp: %s\n", id, nom, email, mdp)
	}
