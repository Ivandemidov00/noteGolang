//Редактирование, добавление, удаление и просмотр заметок
//происходит при помощи создания тестового файла название которого является название заметки
//так же как и название файла с изображением является название заметки

//РАЗМЕТКА
//Для поддержки адаптивности используются единицы измерения как hv,hw и %, большего не требуется т к вёрстка относительно проста
//
///usr/bin/true; exec /usr/bin/env go run "$0" "$@"
package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Page struct {
	Title     string
	Body      []byte
	ImageName string
}
type File struct {
	NameFile      []string
	BegimFileName []string
}
type Name struct {
	FileN File
}

func (p *Page) save() error { //запись и сохранение файла
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}
func (p *Page) delete() error { //удаление файла
	filename := p.Title + ".txt"
	return os.Remove(filename)
}
func getFileName() *Name { //
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	fileN := File{NameFile: getTxt(files), BegimFileName: getBegin(files)}
	return &Name{FileN: fileN}
}
func getTxt(fileInf []fs.FileInfo) []string { //
	var fileTxt []string
	for _, file := range fileInf {
		if strings.HasSuffix(file.Name(), "txt") {
			fileTxt = append(fileTxt, file.Name())
		}

	}
	return fileTxt
}
func getBegin(fileInf []fs.FileInfo) []string {
	var beginFile []string
	for _, file := range fileInf {
		if strings.HasSuffix(file.Name(), "txt") {
			beginFile = append(beginFile, strings.Trim(file.Name(), ".txt"))
		}

	}
	return beginFile
}
func initImage(_nameFile string, title string) { //добавление изображения
	var path, er = os.Getwd() //получаем путь где расположен main.go
	if er != nil {
		fmt.Println("Absolute:", path)
	}
	var nameFile string
	var err error
	nameFile, err = filepath.Abs(_nameFile) //получаем путь изображениия
	if err != nil {
		fmt.Println("Absolute:", nameFile)
	}
	if checkPath(_nameFile, title) && !strings.Contains(nameFile, "/static/images/") { //содержит ли static/images файлы равные текщему названию изображения либо заметки
		nameFile = path + "/static/images/" + _nameFile
	}
	var end = filepath.Ext(nameFile)
	var s = path + "/static/images/" + title + end
	os.Rename(nameFile, s)
}
func checkPath(_path string, title string) bool {
	files, err := ioutil.ReadDir("./static/images")
	if err != nil {
		log.Fatal(err)
	}
	var end = filepath.Ext(_path)
	var path, er = os.Getwd()
	if er != nil {
		fmt.Println("Absolute:", path)
	}
	for _, file := range files {
		if strings.TrimSuffix(file.Name(), end) == title || file.Name() == _path {

			var endTitle = filepath.Ext(_path)
			err = os.Rename(path+"/static/images/"+title+endTitle, path+"/static/images/"+strconv.Itoa(rand.Int())+end)
			if err != nil {
				return false
			} //если у редактируемой заметки уже было изображение то его название смениться на набор случайных чисел
			return true
		}
	}
	return false
}
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	files, err := ioutil.ReadDir("./static/images")
	var im string
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.Name()[:strings.IndexByte(file.Name(), '.')] == title {
			im = file.Name()
		}
	}
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body, ImageName: im}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, _ *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	image := r.FormValue("file")
	if len(image) > 0 {
		initImage(image, title)
	}
	p := &Page{Title: title, Body: []byte(body), ImageName: image}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.delete()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
func createHandler(w http.ResponseWriter, r *http.Request, title string) {
	title = r.FormValue("name")
	http.Redirect(w, r, "/edit/"+title, http.StatusFound)
}
func indexHandler(w http.ResponseWriter, _ *http.Request) {
	p := getFileName()
	renderIndex(w, "index", p)
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderIndex(w http.ResponseWriter, tmpl string, n *Name) {
	err := templates.ExecuteTemplate(w, tmpl+".html", n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view|delete|create|style|static)/([a-zA-Z0-9]+)$")

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
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))
	http.HandleFunc("/create/", makeHandler(createHandler))
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8098", nil))
}
