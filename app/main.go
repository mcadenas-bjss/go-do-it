package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/mcadenas-bjss/go-do-it/app/utils"
)

const (
	baseURL    = "http://localhost:8000/api"
	WEBAPP_URL = "https://cricket-rational-pika.ngrok-free.app"
)

func main() {
	fmt.Println("Running Fyne app")

	app := NewApp()
	app.Window.Resize(fyne.NewSize(700, 500))

	todos := app.newTodoList()

	buttonBox := container.New(layout.NewVBoxLayout())
	form := app.newTodoForm()
	app.form = form
	syncButton := widget.NewButton("Sync", func() {
		log.Println("Synchronizing list")
		go func() {
			app.fetchAll()
			todos.Refresh()
		}()
	})
	app.Synchronize = syncButton
	buttonBox.Add(utils.NewSectionLabel("Add a new todo:"))
	buttonBox.Add(form)
	buttonBox.Add(syncButton)

	top := canvas.NewText("To Do:", color.White)
	top.Alignment = fyne.TextAlignLeading
	top.TextSize = 24.0
	top.TextStyle = fyne.TextStyle{Bold: true}
	appLayout := container.NewBorder(top, buttonBox, nil, nil, todos)
	app.Window.SetContent(appLayout)

	app.Synchronize.OnTapped()
	app.Window.ShowAndRun()
}

type Todo struct {
	Id          int
	Time        string
	Description string
	Completed   bool
}

type Store struct {
	data           map[int]Todo
	RequestChannel chan<- Command
}

type CommandType int

const (
	GetCommand = iota
	GetAllCommand
	InsertCommand
	UpdateCommand
	DeleteCommand
	ToggleCommand
)

type Command struct {
	Cmd     CommandType
	Payload interface{}
	Reply   chan interface{}
	Err     chan error
}

type App struct {
	App         fyne.App
	Window      fyne.Window
	Store       Store
	Synchronize *widget.Button
	form        *widget.Form
}

func NewApp() *App {
	a := app.New()
	w := a.NewWindow("Go Do It")
	store := Store{
		data: make(map[int]Todo),
	}
	store.StartManager()
	return &App{
		App:    a,
		Window: w,
		Store:  store,
	}
}

func (a *App) newTodoList() fyne.CanvasObject {
	length := func() int {
		return len(a.Store.data)
	}
	create := func() fyne.CanvasObject {
		return a.NewTodoListItem()
	}
	selected := func(id widget.ListItemID) {
		fmt.Println("Selected", id)
	}
	updateItem := func(id widget.ListItemID, obj fyne.CanvasObject) {
		todo := a.Store.data[id+1]

		checkbox := obj.(*fyne.Container).Objects[0].(*widget.Check)
		checkbox.SetChecked(todo.Completed)
		checkbox.OnChanged = func(value bool) {
			log.Printf("checkbox %d clicked", id)
			go func() {
				a.toggle(id + 1)
			}()
		}

		description := obj.(*fyne.Container).Objects[1].(*widget.Label)
		description.SetText(todo.Description)

		dueText := obj.(*fyne.Container).Objects[3].(*canvas.Text)
		if len(todo.Time) > 0 {
			dueText.Text = "Due:"
			dueText.Show()
		} else {
			dueText.Hide()
		}

		dateTime := obj.(*fyne.Container).Objects[4].(*widget.Label)
		dateTime.SetText(utils.FormatDueDateTime(todo.Time))
	}
	t := widget.NewList(length, create, updateItem)
	t.OnSelected = selected
	return t
}

func (a *App) NewTodoListItem() fyne.CanvasObject {
	checkbox := widget.NewCheck("", nil)
	description := widget.NewLabel("")
	dateTime := widget.NewLabel("")
	content := container.New(layout.NewHBoxLayout(), checkbox, description, layout.NewSpacer(), canvas.NewText("Due:", color.White), dateTime)
	return content
}

func (a *App) newTodoForm() *widget.Form {
	description := widget.NewEntry()
	day := widget.NewSelectEntry(getRange(1, 31))
	month := widget.NewSelectEntry(getRange(1, 12))
	year := widget.NewSelectEntry(getRange(2023, 2030))
	hours := widget.NewSelectEntry(getRange(0, 23))
	minutes := widget.NewSelectEntry(getRange(0, 59))
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Description", Widget: description},
			{Text: "Day:", Widget: day},
			{Text: "Month:", Widget: month},
			{Text: "Year:", Widget: year},
			{Text: "Hour:", Widget: hours},
			{Text: "Minutes:", Widget: minutes},
		},
		OnSubmit: func() {
			log.Println("Submitting todo")
			var timeString = ""
			if len(day.Text) > 0 {
				timeString = fmt.Sprintf("%s-%s-%sT%s:%s:00Z", year.Text, month.Text, day.Text, hours.Text, minutes.Text)
			}
			a.insert(Todo{Id: 0, Description: description.Text, Time: timeString, Completed: false})
		},
	}
	form.Orientation = widget.Horizontal
	return form
}

func (a *App) resetForm() {
	for i, f := range a.form.Items {
		switch i {
		case 0:
			f.Widget.(*widget.Entry).SetText("")
		case 1:
		case 2:
		case 3:
		case 4:
		case 5:
			f.Widget.(*widget.SelectEntry).SetText("")
		}
	}
}

func (s *Store) StartManager() {
	cmds := make(chan Command)

	go func() {
		for cmd := range cmds {
			switch cmd.Cmd {
			case GetAllCommand:
				log.Println("GetAllCommand")
				if todos, err := s.all(); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- todos
				}
			case InsertCommand:
				log.Println("InsertCommand")
				if err := s.insert(cmd.Payload.(Todo)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- true
				}
			case ToggleCommand:
				log.Println("ToggleCommand")
				if err := s.toggle(cmd.Payload.(int)); err != nil {
					cmd.Err <- err
				} else {
					cmd.Reply <- true
				}
			default:
				log.Fatal("unknown command type", cmd.Cmd)
			}
		}
	}()
	s.RequestChannel = cmds
}

func (s *Store) all() ([]Todo, error) {
	url := fmt.Sprintf("%s/todos", baseURL)

	request, _ := http.NewRequest(http.MethodGet, url, nil)
	request.Header.Add("Accept", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)

	if response.StatusCode != 200 {
		return []Todo{}, err
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return []Todo{}, err
	}

	v := make([]Todo, 0)

	err = json.Unmarshal(content, &v)
	if err != nil {
		return []Todo{}, err
	}

	return v, nil
}

func (s *Store) insert(todo Todo) error {
	url := fmt.Sprintf("%s/todo", baseURL)

	buff := bytes.Buffer{}
	json.NewEncoder(&buff).Encode(todo)
	request, _ := http.NewRequest(http.MethodPost, url, &buff)
	request.Header.Add("Accept", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)

	if response.StatusCode != 200 {
		return err
	}

	return nil
}

func (s *Store) toggle(id int) error {
	url := fmt.Sprintf("%s/todo/toggle/%d", baseURL, id)

	request, _ := http.NewRequest(http.MethodPost, url, nil)
	request.Header.Add("Accept", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)

	if response.StatusCode != 200 {
		return err
	}

	return nil
}

func (a *App) fetchAll() {
	log.Println("Fetching all todos")

	errChan := make(chan error)
	defer close(errChan)
	replyChan := make(chan interface{})
	defer close(replyChan)

	cmd := Command{
		Cmd:     GetAllCommand,
		Payload: nil,
		Reply:   replyChan,
		Err:     errChan,
	}

	a.Store.RequestChannel <- cmd

	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("%v", err)
			return
		}
	case reply := <-replyChan:
		log.Printf("Received reply from fetchAll. Count: %d", len(reply.([]Todo)))
		m := make(map[int]Todo)
		for _, t := range reply.([]Todo) {
			m[t.Id] = t
		}
		a.Store.data = m
	}
}

func (a *App) insert(todo Todo) {
	log.Println("Inserting todo")
	errChan := make(chan error)
	defer close(errChan)
	replyChan := make(chan interface{})
	defer close(replyChan)

	cmd := Command{
		Cmd:     InsertCommand,
		Payload: todo,
		Reply:   replyChan,
		Err:     errChan,
	}

	a.Store.RequestChannel <- cmd

	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("%v", err)
			return
		}
	case reply := <-replyChan:
		log.Println("Received reply from insert", reply)
		a.Synchronize.OnTapped()
		a.resetForm()
	}
}

func (a *App) toggle(id int) {
	log.Printf("Toggling todo %d", id)
	errChan := make(chan error)
	defer close(errChan)
	replyChan := make(chan interface{})
	defer close(replyChan)

	cmd := Command{
		Cmd:     ToggleCommand,
		Payload: id,
		Reply:   replyChan,
		Err:     errChan,
	}

	a.Store.RequestChannel <- cmd

	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("%v", err)
			return
		}
	case reply := <-replyChan:
		log.Println("Received reply from toggle", reply)
		a.Synchronize.OnTapped()
	}
}

func getRange(start, end int) []string {
	rangeList := []string{}
	for i := start; i <= end; i++ {
		if i < 10 {
			rangeList = append(rangeList, fmt.Sprintf("0%d", i))
		} else {
			rangeList = append(rangeList, fmt.Sprintf("%d", i))
		}
	}
	return rangeList
}
