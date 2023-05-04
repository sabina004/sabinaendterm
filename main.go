package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type server struct {
	db *sql.DB
}

type Order struct {
	OrderId   int
	UserId    int
	DesignId  int
	CreatedAt time.Time
}

type Design struct {
	DesignId   int
	Name       string
	DesignType string
	Price      int
}

const (
	host     = "localhost"
	port     = 5433
	user     = "postgres"
	password = "12345"
	dbname   = "homedesign"
)

func dbConnect() *server {
	dbconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", dbconn)
	fmt.Println("Opening database")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to database")

	return &server{db: db}

}

func (s *server) catalogPage(w http.ResponseWriter, r *http.Request) {
	var designs []Design
	res, _ := s.db.Query("select * from designs")
	for res.Next() {
		var design Design
		res.Scan(&design.DesignId, &design.Name, &design.DesignType, &design.Price)
		designs = append(designs, design)
	}
	t, _ := template.ParseFiles("static/html/catalog.html")
	t.Execute(w, designs)
}

func (s *server) registerPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}
		username := r.FormValue("username")
		first_name := r.FormValue("first_name")
		last_name := r.FormValue("last_name")
		phone := r.FormValue("phone")
		pass := r.FormValue("password")

		if _, err := s.db.Query("select * from users where username=$1", username); err == nil {
			fmt.Print("User with this username is already exists!")
			t, _ := template.ParseFiles("static/html/register.html")
			t.Execute(w, nil)
			return
		}

		if _, err := s.db.Exec("insert into users(username, firstname, lastname, phone, password) values($1, $2, $3, $4, $5)", username, first_name, last_name, phone, pass); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/register.html")
	t.Execute(w, nil)
}

func (s *server) authPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}
		username := r.FormValue("username")
		pass := r.FormValue("password")
		var passRight string
		if err := s.db.QueryRow("select password from users where username=$1", username).Scan(&passRight); err != nil {
			fmt.Print("No such user!")
			t, _ := template.ParseFiles("static/html/auth.html")
			t.Execute(w, nil)
			return
		}
		if pass != passRight {
			fmt.Print("Password is incorrect")
			t, _ := template.ParseFiles("static/html/auth.html")
			t.Execute(w, nil)
			return
		}
		fmt.Print("You are logged in!")
		if username == "sabina" {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return
	}
	t, _ := template.ParseFiles("static/html/auth.html")
	t.Execute(w, nil)
}

func (s *server) addDesignPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}

		name := r.FormValue("name")
		designType := r.FormValue("type")
		price := r.FormValue("price")
		if _, err := s.db.Exec("insert into designs(name, type, price) values($1, $2, $3)", name, designType, price); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/addDesign.html")
	t.Execute(w, nil)
}

func (s *server) deleteDesignPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}
		id := r.FormValue("id")
		if _, err := s.db.Exec("delete from designs where designid=$1", id); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/deleteDesign.html")
	t.Execute(w, nil)
}

func (s *server) updateDesignPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}

		id := r.FormValue("id")
		name := r.FormValue("name")
		designType := r.FormValue("type")
		price := r.FormValue("price")
		if _, err := s.db.Exec("update designs set name=$1, type=$2, price=$3 where designid=$4", name, designType, price, id); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/updateDesign.html")
	t.Execute(w, nil)
}

func adminPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/html/admin.html")
	t.Execute(w, nil)
}

func (s *server) ordersPage(w http.ResponseWriter, r *http.Request) {
	var orders []Order
	res, _ := s.db.Query("select * from orders")
	for res.Next() {
		var order Order
		res.Scan(&order.OrderId, &order.UserId, &order.DesignId, &order.CreatedAt)
		orders = append(orders, order)
	}
	t, _ := template.ParseFiles("static/html/orders.html")
	t.Execute(w, orders)
}

func (s *server) addOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}

		user := r.FormValue("user")
		design := r.FormValue("design")
		createdAt := time.Now().UTC()
		if _, err := s.db.Exec("insert into orders(userid, designid, createdat) values($1, $2, $3)", user, design, createdAt); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/addOrder.html")
	t.Execute(w, nil)
}

func (s *server) updateOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}

		id := r.FormValue("id")
		userid := r.FormValue("user")
		design := r.FormValue("design")
		if _, err := s.db.Exec("update orders set userid=$1, designid=$2 where orderid=$3", userid, design, id); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/updateOrder.html")
	t.Execute(w, nil)
}

func (s *server) deleteOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}
		id := r.FormValue("id")
		if _, err := s.db.Exec("delete from orders where orderid=$1", id); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/deleteOrder.html")
	t.Execute(w, nil)
}

func (s *server) deleteUserPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}
		id := r.FormValue("id")
		if _, err := s.db.Exec("delete from users where userid=$1", id); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/deleteUser.html")
	t.Execute(w, nil)
}

func (s *server) updateUserPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Fatal(err)
		}

		id := r.FormValue("id")
		username := r.FormValue("username")
		first_name := r.FormValue("first_name")
		last_name := r.FormValue("last_name")
		phone := r.FormValue("phone")
		pass := r.FormValue("password")

		if _, err := s.db.Query("select * from users where username=$1 and where id not in($2)", username, id); err == nil {
			fmt.Print("User with this username is already exists!")
			t, _ := template.ParseFiles("static/html/updateUser.html")
			t.Execute(w, nil)
			return
		}

		if _, err := s.db.Exec("update users set username=$1, firstname=$2, lastname=$3, phone=$4, password=$5 where userid=$6", username, first_name, last_name, phone, pass, id); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseFiles("static/html/updateUser.html")
	t.Execute(w, nil)
}

func main() {
	s := dbConnect()
	defer s.db.Close()

	fileServer := http.FileServer(http.Dir("./static/"))
	http.Handle("/", fileServer)
	http.HandleFunc("/catalog", s.catalogPage)
	http.HandleFunc("/register", s.registerPage)
	http.HandleFunc("/orders", s.ordersPage)
	http.HandleFunc("/auth", s.authPage)

	http.HandleFunc("/admin", adminPage)

	http.HandleFunc("/adddesign", s.addDesignPage)
	http.HandleFunc("/deletedesign", s.deleteDesignPage)
	http.HandleFunc("/updatedesign", s.updateDesignPage)

	http.HandleFunc("/addorder", s.addOrderPage)
	http.HandleFunc("/updateorder", s.updateOrderPage)
	http.HandleFunc("/deleteorder", s.deleteOrderPage)

	http.HandleFunc("/deleteuser", s.deleteUserPage)
	http.HandleFunc("/updateuser", s.updateUserPage)
	http.ListenAndServe(":8080", nil)
}
