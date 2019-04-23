package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"text/template"
)

type User struct {
	ID    int
	Uname string
	Pass  string
}

type Users struct {
	Users []User
}

type Home struct {
	article string
	articles []string
}

type Contact struct {
	name string
	email string
	phonenumber string
	comments string
}

var session map[string]bool

func main() {
	session = make(map[string]bool)

	// frontend
	http.HandleFunc("/login", formhtml)
	http.HandleFunc("/about", abouthtml)
	http.HandleFunc("/contactus", contactushtml)
	http.HandleFunc("/home", homehtml)

	// backend
	http.HandleFunc("/api/register", register)
	http.HandleFunc("/api/login", login)
	http.HandleFunc("/api/home", home)
	http.HandleFunc("/api/contactus", contactus)

	err := http.ListenAndServe(":8080", nil) // set listen port to 8080
	if err != nil {
		log.Fatal("Error running service: ", err)
	}
}

// frontend
func formhtml(w http.ResponseWriter, r *http.Request) {
	const tpl = `
	<!DOCTYPE html>
	<div class="log-form">
  		<h2>Login to your account</h2>
    	<input id="uname" type="text" title="username" placeholder="username" />
    	<input id="pass" type="password" title="username" placeholder="password" />
		<button type="submit" class="btn">Login</button>
		<button onclick="httpGetAsync()" id="registerbutton" type="submit" class="btn">Register</button>
    	<a class="forgot" href="#">Forgot Username?</a>
	</div>
	<script>
	function httpGetAsync() {
		var uname = document.getElementById("uname").value;
		var pass = document.getElementById("pass").value;
		var param = new XMLHttpRequest("uname","pass").value;
		alert(uname,pass);
		var xmlHttp = new XMLHttpRequest();
		xmlHttp.onreadystatechange = function() { 
			if (xmlHttp.readyState == XMLHttpRequest.DONE)
				alert(xmlHttp.responseText);
		}
		var url = "http://localhost:8080/api/register?" + "uname=" + uname + "&pass=" +  pass;
		alert(url);
		xmlHttp.open("GET", url, true); // true for asynchronous 
		xmlHttp.send(null);
	}
	</script>`

	t, err := template.New("login").Parse(tpl)
	check(err)

	data := struct {
		User  string
		Users []string
	}{
		User: "Login Form",
		Users: []string{
			"Your Username",
			"Your Password",
		},
	}

	err = t.Execute(w, data)
	check(err)
}

func abouthtml(w http.ResponseWriter, r *http.Request) {
	const tpl = `
	<!DOCTYPE html>
	<html>
		<head>
			<meta charset="UTF-8">
			<title>About Us</title>
		</head>
		<body>
			<h1>About Us</h1>
			<p>
				Halo, Selamat Datang di Website Perusahaan Kami.
				Perusahaan Kami Bergerak di Bidang IT.
				Perusahaan ini telah berdiri dari tahun 2000
			</p>
		</body>
	<html>
	`

	t, err := template.New("about").Parse(tpl)
	check(err)

	err = t.Execute(w, nil)
	check(err)
}

func homehtml(w http.ResponseWriter, r *http.Request) {
	const tpl = `
	<!DOCTYPE HTML>
	<html>
	<head>
	<title>Retrieve data from database </title>
	</head>
	<body>

	<?php
	// Connect to database server
	mysql_connect("mysql.myhost.com", "test", "articles") or die (mysql_error ());

	// Select database
	mysql_select_db("test") or die(mysql_error());

	// SQL query
	$strSQL = "SELECT * FROM articles";

	// Execute the query (the recordset $rs contains the result)
	$rs = mysql_query($strSQL);
	
	// Loop the recordset $rs
	// Each row will be made into an array ($row) using mysql_fetch_array
	while($row = mysql_fetch_array($rs)) {

	   // Write the value of the column FirstName (which is now in the array $row)
	  echo $row['articles'] . "<br />";

		}
	?>
	</body>
	</?php>
	</html>
`
	t, err := template.New("home").Parse(tpl)
	check(err)

	err = t.Execute(w, nil)
	check(err)
}

func contactushtml(w http.ResponseWriter, r *http.Request) {
	const tpl = `
	<!DOCTYPE html>
	<div class="log-form">
  		<h2>Write Your Comments Below!</h2>
    	<input id="Name" type="text" title="Name" placeholder="Name" />
		<input id="PhoneNumber" type="text" title="PhoneNumber" placeholder="PhoneNumber" />
		<input id="FromEmailAddress" type="text" title="FromEmailAddress" placeholder="FromEmailAddress" />
		<input id="Comments" type="text" title="Comments" placeholder="Comments" />
		<button onclick="httpGetAsync()" id="submitbutton" type="submit" class="btn">Submit</button>
	</div>
	<script>
	function httpGetAsync() {
		var Name = document.getElementById("Name").value;
		var PhoneNumber = document.getElementById("PhoneNumber").value;
		var FromEmailAddress = document.getElementById("FromEmailAddress").value;
		var Comments = document.getElementById("Comments").value;
		var param = new XMLHttpRequest("Name","PhoneNumber","FromEmailAddress","Comments").value;
		alert(Name,PhoneNumber,"FromEmailAddress","Comments");
		var xmlHttp = new XMLHttpRequest();
		xmlHttp.onreadystatechange = function() { 
			if (xmlHttp.readyState == XMLHttpRequest.DONE)
				alert(xmlHttp.responseText);
		}
		var url = "http://localhost:8080/api/contactus?" + "Name=" + Name + "&PhoneNumber=" +  PhoneNumber 
		+ "FromEmailAddress=" + FromEmailAddress + "Comments=" + Comments;
		alert(url);
		xmlHttp.open("GET", url, true); // true for asynchronous 
		xmlHttp.send(null);
	}
	</script>`

	t, err := template.New("contactus").Parse(tpl)
	check(err)

	err = t.Execute(w, nil)
	check(err)
}


// backend
func register(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10)
	log.Println(r.FormValue("uname"))

	var user User

	if r.FormValue("uname") == "" {
		// code if field is empty
		fmt.Println("Username Tidak Boleh Kosong")
		return
	}
	user.Uname = r.FormValue("uname")

	if r.FormValue("pass") == "" {
		// code if field is empty
		fmt.Println("Password Tidak Boleh Kosong")
		return
	}
	user.Pass = r.FormValue("pass")

	db, err := sql.Open("mysql", "root:smanlibogor@/test?charset=utf8")
	checkErr(err)

	// insert
	stmt, err := db.Prepare("insert into `course` values (null, ?, ?)")
	checkErr(err)

	res, err := stmt.Exec(user.Uname, user.Pass)
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(id)

	w.Write([]byte("string"))
}

func login(w http.ResponseWriter, r *http.Request) {

	var user User

	r.ParseForm()
	if r.Form["uname"][0] == "" {
		// code if field is empty
		fmt.Println("Username Tidak Boleh Kosong")
		return
	}
	user.Uname = r.Form["uname"][0]

	if r.Form["Pass"][0] == "" {
		// code if field is empty
		fmt.Println("Password Tidak Boleh Kosong")
		return
	}
	user.Pass = r.Form["Pass"][0]

	if m, _ := regexp.MatchString("^[a-zA-Z]+$", r.Form.Get("username")); !m {
		return
	}

	session[user.Uname] = true

	db, err := sql.Open("mysql", "root:smanlibogor@/course?charset=utf8")
	checkErr(err)

	rows, err := db.Query("select * from course")
    checkErr(err)

    for rows.Next() {
        var ID int
        var Uname string
        var Pass string
        err = rows.Scan(&ID, &Uname, &Pass)
        checkErr(err)
        fmt.Println(ID)
        fmt.Println(Uname)
        fmt.Println(Pass)
    }
}

func home(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("mysql", "root:smanlibogor@/course?charset=utf8")
	checkErr(err)

	rows, err := db.Query("select article from course")
	checkErr(err)

	for rows.Next() {
		var article string
		err = rows.Scan(&article)
		checkErr(err)
		fmt.Println(article)
	}
}

func islogin(uname string)bool {
	islogin := session[uname]
    return islogin
}

func contactus(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10)
	log.Println(r.FormValue("Contact"))

	var contact Contact

	if r.FormValue("Name") == "" {
		// code if field is empty
		fmt.Println("Nama Tidak Boleh Kosong")
		return
	}
	contact.name = r.FormValue("Name")

	if r.FormValue("PhoneNumber") == "" {
		// code if field is empty
		fmt.Println("Nomor Telfon Tidak Boleh Kosong")
		return
	}
	contact.phonenumber = r.FormValue("PhoneNumber")

	if r.FormValue("FromEmailAddress") == "" {
		// code if field is empty
		fmt.Println("Email Tidak Boleh Kosong")
		return
	}
	contact.email = r.FormValue("FromEmailAddress")

	if r.FormValue("Comments") == "" {
		// code if field is empty
		fmt.Println("Pesan Tidak Boleh Kosong")
		return
	}
	contact.comments = r.FormValue("Comments")

	db, err := sql.Open("mysql", "root:smanlibogor@/test?charset=utf8")
	checkErr(err)

	// insert
	stmt, err := db.Prepare("insert into `course` values (null, ?, ?, ?, ?)")
	checkErr(err)

	res, err := stmt.Exec(contact.name, contact.phonenumber, contact.email, contact.comments)
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(id)

	w.Write([]byte("id"))
}

func article(w http.ResponseWriter, r *http.Request) {

	var article Home

	db, err := sql.Open("mysql", "root:smanlibogor@/test?charset=utf8")
	checkErr(err)

	//insert
	stmt, err := db.Prepare("insert into `articles` values (null, ?)")
	checkErr(err)

	res, err := stmt.Exec(article.articles)
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(id)

	w.Write([]byte("id"))

	// update
    stmt, err = db.Prepare("update `userinfo` set articles=? where uid=?")
    checkErr(err)

    res, err = stmt.Exec("adminupdate", id)
    checkErr(err)

    affect, err := res.RowsAffected()
    checkErr(err)

    fmt.Println(affect)

    // query
    rows, err := db.Query("select * from articles")
    checkErr(err)

    for rows.Next() {
        var article string
        err = rows.Scan(&article)
        checkErr(err)
        fmt.Println(article)
    }

    // delete
    stmt, err = db.Prepare("delete from `articles` where uid=?")
    checkErr(err)

    res, err = stmt.Exec(id)
    checkErr(err)

    affect, err = res.RowsAffected()
    checkErr(err)

    fmt.Println(affect)

    db.Close()
}

func islogout(uname string)bool {
	islogout := session[uname]
	islogout = false
	return islogout
}

func koneksi(w http.ResponseWriter, r *http.Request) {
	jsonFile, err := os.Open("user.json")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Success open user.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var users Users

	err = json.Unmarshal([]byte(byteValue), &users)
	fmt.Println(string(byteValue[:]))
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(users.Users); i++ {
		fmt.Println("User User Name: " + users.Users[i].Uname)
		fmt.Println("User Password: " + users.Users[i].Pass)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}