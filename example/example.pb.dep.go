package example

import (
	"database/sql"
	_ "github.com/lib/pq"
)

// Lets start by creating a Model and Handler for our flow
type Handler struct {
	m *model
}

type model struct {
	DB    *sql.DB
	Table string
}

func NewHandler(db *sql.DB) *Handler {
	model := New(model)
	model.DB = db
	model.Table = "hello"
}

// List function should return a list of these objects
func (m *model) List(tenant string) (map[int]Hello, error) {
	ret := make(map[int]Hello)

	rows, err := m.DB.Query("SELECT id, data FROM list_data($1, $2)", tenant, m.Table)
	if err != nil {
		return ret, err
	}

	defer rows.Close()

	for rows.Next() {
		var row Hello
		var id int

		err := rows.Scan(&id, &row)
		if err != nil {
			return ret, err
		}

		ret[id] = row
	}

	return ret, nil
}

// Create function will create a new object of this type
func (m *model) Create(tenant string, data *Hello) error {
	if Hello.Email != "" {
		_, err := m.DB.Exec("CALL insert_data($1, $2, $3)", tenant, m.Table, contact)

		if err != nil {
			return err
		}
	} else {
		return errors.New("Email Was not set")
	}

	return nil
}

// Update function will replace the object stored at the given ID
func (m *model) Update(tenant string, id string, data *Hello) error {
	_, err := m.DB.Exec("CALL update_data($1, $2, $3, $4)",
		tenant, m.Table, id, data)

	return err
}

// Delete function will... well delete the object at given ID
func (m *model) Delete(tenant string, id string) error {
	_, err := m.DB.Exec("CALL delete_data_by_id($1, $2, $3)",
		tenant, m.Table, id)

	return err
}

// A simple function to handle a htmx form and populate the struct
func (x *Hello) HandleForm(req *http.Request) error {
	x.Email = req.FormValue("Hello__Email")
	x.Name = req.FormValue("Hello__Name")
	return x.Validate()
}

// Deps function returns a static string for the time being, needs dev
func (t *Hello) Deps() string {
	return "Hello"
}
