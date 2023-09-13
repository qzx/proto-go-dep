package example

import (
	"database/sql"
	_ "github.com/lib/pq"
	"net/http"
	"github.com/go-chi/chi/v5"
)

// ListHandler is our http handler that acquires and renders a list of objects
func (x *Hello) ListHandler(w http.ResponseWriter, req *http.Request) {
	db, err := r.Context().Value("db").(*sql.DB)
	if err != nil {
		return
	}

	tenant := chi.URLParam(req, "id")
	ret, err := x.List(db, tenant)
	if err != nil {
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// List function should return a list of these objects
func (x *Hello) List(db *sql.DB, tenant string) (map[int]Hello, error) {
	ret := make(map[int]Hello)

	rows, err := db.Query("SELECT id, data FROM list_data($1, $2)", tenant, x.TableName())
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

// Get function acquires a single record based on ID in database
func (x *Hello) Get(db *sql.DB, tenant string, id string) error {

	return db.QueryRow("SELECT data FROM list_data($1, $2) WHERE id = $3",
		tenant, x.TableName(), id).Scan(x)

}

// Create function will create a new object of this type
func (x *Hello) Create(db *sql.DB, tenant string, data *Hello) error {
	if Hello.Email != "" {
		_, err := db.Exec("CALL insert_data($1, $2, $3)", tenant, x.TableName(), contact)

		if err != nil {
			return err
		}
	} else {
		return errors.New("Email Was not set")
	}

	return nil
}

// Update function will replace the object stored at the given ID
func (x *Hello) Update(db *sql.DB, tenant string, id string, data *Hello) error {
	_, err := db.Exec("CALL update_data($1, $2, $3, $4)",
		tenant, x.TableName(), id, data)

	return err
}

// Delete function will... well delete the object at given ID
func (x *Hello) Delete(db *sql.DB, tenant string, id string) error {
	_, err := db.Exec("CALL delete_data_by_id($1, $2, $3)",
		tenant, x.TableName(), id)

	return err
}

// A simple function to handle a htmx form and populate the struct
func (x *Hello) HandleForm(req *http.Request) error {
	x.Email = req.FormValue("Hello__Email")
	x.Name = req.FormValue("Hello__Name")
	return x.Validate()
}

// Deps function returns a static string for the time being, needs dev
func (*Hello) TableName() string {
	return "hello"
}
