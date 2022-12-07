package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"cloud.google.com/go/datastore"
)

var validPath = regexp.MustCompile("^/(edit|save|view|list|delete)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body  []byte
}
type User struct {
	Full_name string
	Email     string
	Dob       time.Time
}
type AllUsers struct {
	users []User
}
type Counter struct {
	Count int
}

// [START datastore_add_entity]
// Task is the model used to store tasks in the datastore.
type Task struct {
	Desc    string    `datastore:"description"`
	Created time.Time `datastore:"created"`
	Done    bool      `datastore:"done"`
	id      int64     // The integer ID used in the datastore.
	// PrimeKey int64          `datastore:"prime_key"`
	K *datastore.Key `datastore:"__key__"`
}

type AllData struct {
	Tasks []*Task
	Dummy string
}

// AddTask adds a task with the given description to the datastore,
// returning the key of the newly created entity.
func AddTask(ctx context.Context, client *datastore.Client, desc string) (*datastore.Key, error) {
	task := &Task{
		Desc:    desc,
		Created: time.Now(),
	}
	key := datastore.IncompleteKey("Task", nil)
	return client.Put(ctx, key, task)
}

// [END datastore_add_entity]

// [START datastore_update_entity]
// MarkDone marks the task done with the given ID.
func MarkDone(ctx context.Context, client *datastore.Client, taskID int64) error {
	// Create a key using the given integer ID.
	key := datastore.IDKey("Task", taskID, nil)

	// In a transaction load each task, set done to true and store.
	_, err := client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var task Task
		if err := tx.Get(key, &task); err != nil {
			return err
		}
		task.Done = true
		_, err := tx.Put(key, &task)
		return err
	})
	return err
}
func GetTask(ctx context.Context, client *datastore.Client, taskID int64) *Task {
	// Create a key using the given integer ID.
	key := datastore.IDKey("Task", taskID, nil)
	var task Task

	// In a transaction load each task, set done to true and store.

	client.Get(ctx, key, &task)

	return &task
}

// [END datastore_update_entity]

// [START datastore_retrieve_entities]
// ListTasks returns all the tasks in ascending order of creation time.
func ListTasks(ctx context.Context, client *datastore.Client) ([]*Task, error) {
	var tasks []*Task

	// Create a query to fetch all Task entities, ordered by "created".
	query := datastore.NewQuery("Task").Order("created")
	keys, err := client.GetAll(ctx, query, &tasks)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// Set the id field on each Task from the corresponding key.
	for i, key := range keys {
		tasks[i].id = key.ID
	}

	return tasks, nil
}

// [END datastore_retrieve_entities]

// [START datastore_delete_entity]
// DeleteTask deletes the task with the given ID.
func DeleteTask(ctx context.Context, client *datastore.Client, taskID int64) error {
	return client.Delete(ctx, datastore.IDKey("Task", taskID, nil))
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Task) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)

}
func renderList(w http.ResponseWriter, tmpl string, p []*Task) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, &AllData{Tasks: p, Dummy: "pass something"})

}
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	ctx := context.Background()

	// Create a datastore client. In a typical application, you would create
	// a single client which is reused for every datastore operation.
	dsClient, err := datastore.NewClient(ctx, "golang-370407")
	if err != nil {
		// Handle error.
	}
	defer dsClient.Close()
	id, _ := strconv.ParseInt(title, 10, 64)
	p := GetTask(ctx, dsClient, id)

	renderTemplate(w, "view", p)
}
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	if title != "0" {
		ctx := context.Background()

		// Create a datastore client. In a typical application, you would create
		// a single client which is reused for every datastore operation.
		dsClient, err := datastore.NewClient(ctx, "golang-370407")
		if err != nil {
			// Handle error.
		}
		defer dsClient.Close()
		id, _ := strconv.ParseInt(title, 10, 64)
		p := GetTask(ctx, dsClient, id)
		renderTemplate(w, "edit", p)

	} else {
		renderTemplate(w, "edit", &Task{K: &datastore.Key{ID: 0}})

	}
}
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	ctx := context.Background()

	// Create a datastore client. In a typical application, you would create
	// a single client which is reused for every datastore operation.
	dsClient, err := datastore.NewClient(ctx, "golang-370407")
	if err != nil {
		// Handle error.
	}
	defer dsClient.Close()
	if title == "0" {
		task, _ := AddTask(ctx, dsClient, r.FormValue("body"))
		new_task := GetTask(ctx, dsClient, task.ID)
		key := datastore.IDKey("Task", task.ID, nil)
		if r.FormValue("done") == "on" {
			fmt.Println(r.FormValue("done"))
			new_task.Done = true
		} else {
			fmt.Println(r.FormValue("done"))

			new_task.Done = false
		}

		_, _ = dsClient.Put(ctx, key, new_task)
		http.Redirect(w, r, "/view/"+fmt.Sprint(task.ID), http.StatusFound)
	} else {
		id, _ := strconv.ParseInt(title, 10, 64)
		p := GetTask(ctx, dsClient, id)
		key := datastore.IDKey("Task", id, nil)
		p.Desc = r.FormValue("body")
		if r.FormValue("done") == "on" {
			fmt.Println(r.FormValue("done"))
			p.Done = true
		} else {
			fmt.Println(r.FormValue("done"))

			p.Done = false
		}
		_, _ = dsClient.Put(ctx, key, p)
		http.Redirect(w, r, "/view/"+fmt.Sprint(id), http.StatusFound)

	}

}
func listHandler(w http.ResponseWriter, r *http.Request, title string) {
	ctx := context.Background()

	// Create a datastore client. In a typical application, you would create
	// a single client which is reused for every datastore operation.
	dsClient, err := datastore.NewClient(ctx, "golang-370407")
	if err != nil {
		// Handle error.
	}
	defer dsClient.Close()
	tasks, _ := ListTasks(ctx, dsClient)

	renderList(w, "list", tasks)
}
func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	ctx := context.Background()

	// Create a datastore client. In a typical application, you would create
	// a single client which is reused for every datastore operation.
	dsClient, err := datastore.NewClient(ctx, "golang-370407")
	if err != nil {
		// Handle error.
	}
	defer dsClient.Close()
	id, _ := strconv.ParseInt(title, 10, 64)
	DeleteTask(ctx, dsClient, id)

	tasks, _ := ListTasks(ctx, dsClient)

	renderList(w, "list", tasks)
}

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
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/list/", makeHandler(listHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
