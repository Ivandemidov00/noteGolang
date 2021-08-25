// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.



package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
	ImageName string
}
type File struct {
	NameFile []string
	BegimFileName []string
}
type Name struct {
	FileN File
	//NameFile []string
	//BegimFileName []string
}
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}
func (p *Page) delete() error {
	filename :=p.Title+".txt"
	return os.Remove(filename)
}
func getFileName() *Name{
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	fileN:=File{NameFile: getTxt(files),BegimFileName: getBegin(files)}
	return &Name{FileN:fileN}
	//return &Name{NameFile: getTxt(files),BegimFileName: getBegin(files)}
}
func getTxt(fileInf []fs.FileInfo) []string {
	var fileTxt []string
	for _, file := range fileInf{
		if strings.HasSuffix(file.Name(),"txt"){
			fileTxt=append(fileTxt,file.Name())
		}

	}
	return fileTxt
}
func getBegin(fileInf []fs.FileInfo) []string {
	var beginFile []string
	for _, file := range fileInf{
		if strings.HasSuffix(file.Name(),"txt"){
			beginFile=append(beginFile,strings.Trim(file.Name(),".txt"))
		}

	}
	return beginFile
}
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	image:=r.FormValue("file")
	if len(image)>0 {
		fmt.Println(image)
	}
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
func deleteHandler(w http.ResponseWriter, r *http.Request,title string)  {
	body := r.FormValue("body")
	p:=&Page{Title: title,Body: []byte(body)}
	err := p.delete()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
func createHandler(w http.ResponseWriter,r *http.Request,title string)  {
	title = r.FormValue("name")
	http.Redirect(w,r,"/edit/"+title,http.StatusFound)
}
func indexHandler(w http.ResponseWriter,r *http.Request){
	p := getFileName()
	renderIndex(w,"index",p)
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html","index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderIndex(w http.ResponseWriter,tmpl string, n *Name)  {
	err:=templates.ExecuteTemplate(w,tmpl+".html",n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
var validPath = regexp.MustCompile("^/(edit|save|view|delete|create|style)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.Handle("static/", http.StripPrefix("static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/delete/",makeHandler(deleteHandler))
	http.HandleFunc("/create/",makeHandler(createHandler))
	http.HandleFunc("/",indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}