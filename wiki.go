package main

import (
	"html/template"
	
	"net/http"
	"regexp"
	"errors"
	"labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"

)


var (
	session, _ = mgo.Dial("localhost")
   database = session.DB("wiki").C("pages")

)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
type Page struct {
    Title string
    Body  []byte
}

func (p *Page) save() error {
    //return database.Insert(&Page{p.Title, p.Body})
	_, err := database.Upsert(bson.M{"title": p.Title}, &Page{p.Title, p.Body})
	return err
}

func loadPage(title string) (*Page, error) {
	page  := Page{}
    err := database.Find(bson.M{"title":title}).One(&page)
    if err != nil {
        return nil, err
    }
    return &page, nil
}


func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
        return
    }
	p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
	
}

func editHandler(w http.ResponseWriter, r *http.Request) {
    
	title, err := getTitle(w, r)
    if err != nil {
		 return
        
    }
	p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
	
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
    if err != nil {
        return
    }
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err = p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil // The title is the second subexpression.

}
func main() {

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	

	
	
	
	
	http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler)
    http.HandleFunc("/save/", saveHandler)
	http.ListenAndServe(":8080", nil)
}