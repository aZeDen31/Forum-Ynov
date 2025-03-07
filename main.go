package main

import (
	"fmt"
	"html/template"
	"net/http"
)



func main() {
	server()
}

func server(){
	fileServer := http.FileServer(http.Dir("./html"))

	fs := http.FileServer(http.Dir("./styles"))
	http.Handle("/assets/", http.StripPrefix("/styles/", fs)) 

	http.Handle("/", fileServer)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler) 
	http.HandleFunc("/profile", profileHandler)

	fmt.Println("clique sur le lien http://localhost:7000/")
	if err := http.ListenAndServe(":7000", nil); err != nil {
		panic(err)
	}

}

func loginHandler(w http.ResponseWriter, r *http.Request){
	tmpl := template.Must(template.ParseFiles("html/login.html"))
	tmpl.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request){
	tmpl := template.Must(template.ParseFiles("html/register.html"))

	tmpl.Execute(w, nil)
}

func profileHandler(w http.ResponseWriter, r *http.Request){
	tmpl := template.Must(template.ParseFiles("html/profile.html"))
	tmpl.Execute(w, nil)
} 

func Signup (w http.ResponseWriter, r *http.Request){
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// faut insérer ici la query vers la db 
		// penser a check si user existe déjà
		//si non l'add a la db
		fmt.Printf("Username: %s, Password: %s\n", username, password)
		if true {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}else{
			fmt.Fprintf(w,"Error user is already registered")
		}
	}

	tmpl, err := template.ParseFiles("html/register")
	if err != nil {
		 http.Error(w, err.Error(), http.StatusInternalServerError)
		  return
	}
	tmpl.Execute(w, nil)

}

func LoginPage(w http.ResponseWriter, r *http.Request){

	if r.Method == "POST"{
		username := r.FormValue("username")
		password := r.FormValue("password")

		//Check si les identifiants sont corrects
		fmt.Printf("Username: %s, Password: %s\n", username, password)
		if true {
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return 
		}else{
			fmt.Fprintf(w,"Error user is not registered")
		}
	}
	tmpl, err := template.ParseFiles("templates/login.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, nil)
}