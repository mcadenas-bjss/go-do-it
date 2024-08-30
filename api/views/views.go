package views

import (
	"embed"
	"html/template"
	"io"
	"time"

	"github.com/mcadenas-bjss/go-do-it/store"
)

var (
	//go:embed "templates/*"
	todoTemplates embed.FS
)

type TodoRenderer struct {
	templ *template.Template
}

func NewTodoRenderer() (*TodoRenderer, error) {
	templ, err := template.ParseFS(todoTemplates, "templates/*.gohtml")
	if err != nil {
		return nil, err
	}

	return &TodoRenderer{templ: templ}, nil
}

func (tr *TodoRenderer) RenderTodo(w io.Writer, todo store.Todo) error {
	todoTemplate := `<li id="todo-{{.Id}}">
  <div class="todo">
    <input id="todo-{{.Id}}-checkbox" type="checkbox" {{completed .Completed}} />
    <p>{{.Description}}</p>
    <button
      hx-delete="/api/todo/{{.Id}}"
      hx-swap="delete"
      hx-target="#todo-{{.Id}}">Delete</button
    >
    <div class="meta">
      <time datetime={{.Time}}>{{formatTime .Time}}</time>
    </div>
  </div>
</li>`

	templ, err := template.New("todo").Funcs(
		template.FuncMap{
			"completed": func(b bool) string {
				if b {
					return "checked"
				}
				return ""
			},
			"formatTime": func(t string) string {
				if len(t) == 0 {
					return ""
				}

				var layout = "2006-01-02T15:04:05Z0700" // ISO 8601 format
				var output = "Mon, 02 Jan 2006 15:04"

				check_t, _ := time.Parse(layout, t)
				checkDate_t, _ := time.Parse("2006-01-02", check_t.Format("2006-01-02"))
				nowDate_t, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))

				if len(t) == 0 {
					return ""
				}

				f := check_t.Format(output)

				if inTimeSpan(nowDate_t.AddDate(0, 0, -1), nowDate_t, checkDate_t) {
					f = "Yesterday at " + check_t.Format("3:04 PM")
				}
				if inTimeSpan(nowDate_t, nowDate_t.AddDate(0, 0, 1), checkDate_t) {
					f = "Today at " + check_t.Format("3:04 PM")
				}
				if inTimeSpan(nowDate_t.AddDate(0, 0, 1), nowDate_t.AddDate(0, 0, 2), checkDate_t) {
					f = "Tomorrow at " + check_t.Format("3:04 PM")
				}

				return f
			},
		},
	).Parse(todoTemplate)
	if err != nil {
		return err
	}

	if err := templ.Execute(w, todo); err != nil {
		return err
	}

	return nil
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}
