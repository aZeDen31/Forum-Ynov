package main

import (
	"forum-ynov/database"
	"fmt"
	"log"
)

func main() {
	// Initialisation de la base de données
	db, err := database.InitDB("ma_base.db")
	if err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données :", err)
	}
	defer db.Close()

	fmt.Println("Base de données initialisée et prête à être utilisée !")

	// Exemple d'ajout d'un utilisateur
	err = database.InsertUser(db, "Alice", "alice@example.com", "motdepasse123")
	if err != nil {
		log.Fatal("Erreur lors de l'ajout de l'utilisateur :", err)
	}

	// Lecture des utilisateurs
	err = database.LectureUtilisateurs(db)
	if err != nil {
		log.Fatal("Erreur lors de la lecture des utilisateurs :", err)
	}
}
