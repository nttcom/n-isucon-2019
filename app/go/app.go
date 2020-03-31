package main

import (
	"./utils"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	dbDriver   = "mysql"
	dataSource string
	store      *sessions.CookieStore
	users      map[string]string
)

func initialize(c web.C, w http.ResponseWriter, r *http.Request) {
	_, err := exec.Command("sh", "-c", "../common/db/init.sh").Output()
	if err != nil {
		panic("DB init failed.")
	}

	return
}

func signin(c web.C, w http.ResponseWriter, r *http.Request) {
	var user *User
	w.Header().Set("Content-Type", "application/json")

	contentType := r.Header.Get("Content-type")
	if utils.ContentTypeEqualJSON(contentType) == false {
		utils.SetStatus(w, 400)
		return
	}

	jsonReq := utils.JSONToMap(w, r)
	if jsonReq == nil {
		utils.SetStatus(w, 400)
		return
	}

	username, _ := jsonReq["username"].(string)

	// is username empty?
	if utf8.RuneCountInString(username) <= 0 {
		utils.SetStatus(w, 400)
		return
	}

	password, _ := jsonReq["password"].(string)
	// is password empty?
	if utf8.RuneCountInString(password) <= 0 {
		utils.SetStatus(w, 400)
		return
	}

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		panic("Unable to connect the DB.")
	}
	user, err = SelectUserByUsername(db, username)

	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}

	fmt.Printf("salt:%s hash:%s", user.salt, user.passwordHash)

	if utils.GetPasswordHash(user.salt, password) != user.passwordHash {
		utils.SetStatus(w, 401)
		return
	}

	session, err := store.Get(r, "session")
	if err != nil {
		// recreate session
		session, err = store.New(r, "session")
	}
	session.Values["username"] = username
	session.Save(r, w)

	// Build a response
	result, err := json.Marshal(user)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}
}

func signout(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		utils.SetStatus(w, 401)
		return
	}
	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	username, _ := session.Values["username"].(string)
	fmt.Printf("%s\n", username)

	//delete(session.Values, "username")
	session.Options = &sessions.Options{MaxAge: -1, Path: "/"}
	session.Save(r, w)
	utils.SetStatus(w, 204)
}

func usersGet(c web.C, w http.ResponseWriter, r *http.Request) {
	var user *User
	searchName := c.URLParams["username"]

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		panic("Unable to connect the DB.")
	}
	user, err = SelectUserByUsername(db, searchName)

	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	utils.SetStatus(w, 200)

	// Build a response
	result, err := json.Marshal(user)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}

}

func usersPost(c web.C, w http.ResponseWriter, r *http.Request) {
	jsonReq := utils.JSONToMap(w, r)

	// Is the request format correct?
	contentType := r.Header.Get("Content-type")
	if utils.ContentTypeEqualJSON(contentType) == false {
		utils.SetStatus(w, 400)
		return
	}

	if jsonReq == nil {
		utils.SetStatus(w, 400)
		return
	}

	// Is username empty?

	newusername, _ := jsonReq["username"].(string)
	if utf8.RuneCountInString(newusername) <= 0 {
		utils.SetStatus(w, 400)
		return
	}

	// Is password empty?

	password, _ := jsonReq["password"].(string)
	if utf8.RuneCountInString(password) <= 0 {
		utils.SetStatus(w, 400)
		return
	}

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		panic("Unable to connect the DB.")
	}

	// if username exists

	if UsernameExists(newusername, db) {
		utils.SetStatus(w, 409)
		return
	}

	// create new user.

	salt := utils.GetSalt()
	passwordHash := utils.GetPasswordHash(salt, password)
	const layout = "2006-01-02 15:04:05"
	t := time.Now()
	query := "INSERT INTO users (username, password_hash, salt, created_at, updated_at) VALUES ((?), (?), (?), (?), (?));"
	_, err = db.Exec(query, newusername, passwordHash, salt, t.Format(layout), t.Format(layout))

	if err != nil {
		panic("Unable to insert to the users table.")
		return
	}

	// search new user.

	var newuser *User
	newuser, err = SelectUserByUsername(db, newusername)

	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}

	result, err := json.Marshal(newuser)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	utils.SetStatus(w, 201)
	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}

}

func usersPatch(c web.C, w http.ResponseWriter, r *http.Request) {
	var user *User
	targetName := c.URLParams["username"]
	session, err := store.Get(r, "session")

	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	jsonReq := utils.JSONToMap(w, r)

	if jsonReq == nil {
		return
	}

	newusername, _ := jsonReq["username"].(string)
	newpassword, _ := jsonReq["password"].(string)
	newpasswordHash := ""
	newsalt := ""

	fmt.Printf("target: %s\n", targetName)
	fmt.Printf("newusername: %s\n", newusername)
	fmt.Printf("newpass: %s\n", newpassword)

	// When there is no required value
	if utf8.RuneCountInString(newusername) <= 0 &&
		utf8.RuneCountInString(newpassword) <= 0 {
		utils.SetStatus(w, 400)
		return
	}

	// search target user
	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		panic("Unable to connect the DB.")
	}

	user, err = SelectUserByUsername(db, targetName)

	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}

	//If the request is not yours
	if session.Values["username"] != targetName {
		utils.SetStatus(w, 403)
		return
	}

	if utf8.RuneCountInString(newusername) <= 0 {
		newusername = user.Username
	} else {
		// if exist
		if UsernameExists(newusername, db) {
			utils.SetStatus(w, 409)
			return
		}
	}

	if utf8.RuneCountInString(newpassword) <= 0 {
		newpasswordHash = user.passwordHash
		newsalt = user.salt
	} else {
		newsalt = utils.GetSalt()
		newpasswordHash = utils.GetPasswordHash(newsalt, newpassword)
	}
	fmt.Printf("-----------------------\n")
	fmt.Printf("newsalt: %s\n", newsalt)
	fmt.Printf("newpasswordHash: %s\n", newpasswordHash)
	fmt.Printf("newusername: %s\n", newusername)
	fmt.Printf("newpass: %s\n", newpassword)

	const layout = "2006-01-02 15:04:05"
	t := time.Now()
	query := "UPDATE users SET username=(?), password_hash=(?), salt=(?), updated_at=(?) WHERE username=(?);"
	_, err = db.Exec(query, newusername, newpasswordHash, newsalt, t.Format(layout), targetName)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to update the users table.")
		return
	}

	user, err = SelectUserByUsername(db, newusername)

	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}

	session.Values["username"] = newusername
	session.Save(r, w)

	result, err := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}

	return

}

func usersDelete(c web.C, w http.ResponseWriter, r *http.Request) {
	username := c.URLParams["username"]
	session, err := store.Get(r, "session")

	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		panic("Unable to connect the DB.")
	}

	if UsernameExists(username, db) == false {
		utils.SetStatus(w, 404)
		return
	}

	//If the request is not yours
	if session.Values["username"] != username {
		utils.SetStatus(w, 403)
		return
	}

	query := "DELETE FROM users WHERE username=(?)"
	res, err := db.Exec(query, username)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to delete from the users table.")
		return
	}

	num, _ := res.RowsAffected()

	if num <= 0 {
		utils.SetStatus(w, 500)
		return
	}

	session.Options = &sessions.Options{MaxAge: -1, Path: "/"}
	session.Save(r, w)

	utils.SetStatus(w, 204)
}

func itemsGet(c web.C, w http.ResponseWriter, r *http.Request) {
	const ItemLimit int = 10
	var offset = 0
	var sortLike = false
	var items Items

	queryStr := r.URL.Query()
	fmt.Printf("Quey str: %v\n", queryStr)
	fmt.Printf("OFFSET: %d\n", offset)

	if queryStr["page"] != nil {
		page := queryStr.Get("page")
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			utils.SetStatus(w, 500)
			return
		}
		offset = pageInt * ItemLimit
	}

	if queryStr["sort"] != nil {
		sort := queryStr.Get("sort")
		if sort == "like" {
			sortLike = true
		}
	}

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic("Unable to connect the DB.")
	}

	defer db.Close()

	fmt.Printf("limit: %v\n", ItemLimit)

	if sortLike == false {

		query := "SELECT id, title, user_id, created_at from items ORDER BY created_at DESC"
		rows, err := db.Query(query)
		if err != nil {
			utils.SetStatus(w, 500)
			panic("Unable to get the query results.")
			return
		}

		for rows.Next() {
			var userID int
			result := Item{}
			err := rows.Scan(&result.ID, &result.Title, &userID, &result.CreatedAt)
			if err != nil {
				panic(err)
			}
			user, err := SelectUserByUserID(db, userID)
			if err != nil {
				panic("Unexpected err.")
			}
			result.Username = user.Username
			fmt.Println("res: %v", result)
			items.Items = append(items.Items, result)
		}

	} else {

		query := "SELECT id, title, created_at, likes from items"
		rows, err := db.Query(query)
		if err != nil {
			utils.SetStatus(w, 500)
			panic("Unable to get the query results.")
			return
		}

		for rows.Next() {
			result := Item{}

			err := rows.Scan(&result.ID, &result.Title, &result.CreatedAt, &result.likes)
			if err != nil {
				utils.SetStatus(w, 500)
				panic("Unable to scan from the result.")
				return
			}

			if result.likes.Valid == false {
				result.likeCount = 0
			} else {
				result.likeCount = len(strings.Split(result.likes.String, ","))
			}
			//fmt.Printf("likes = %s  item: %#v\n", result.likes.String, result)
			items.Items = append(items.Items, result)
		}

		sort.Slice(items.Items, func(i, j int) bool {
			return items.Items[i].likeCount > items.Items[j].likeCount
		})

	}

	w.Header().Set("Content-Type", "application/json")

	start := offset
	end := offset + ItemLimit

	if len(items.Items) > end {
		items.Items = items.Items[start:end]
	} else if len(items.Items) > start {
		items.Items = items.Items[start:len(items.Items)]
	} else {
		items.Items = []Item{}
	}

	if items.Items == nil {
		_, err = fmt.Fprintf(w, "{\"items\": [], \"count\": 0}")
		return
	}

	query := "SELECT id from items"
	rows, err := db.Query(query)
	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to get the query results.")
		return
	}

	items.Count = 0
	for rows.Next() {
		_ = rows.Scan()
		items.Count++
	}

	result, err := json.Marshal(items)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}
	return
}
func itemsGetByID(c web.C, w http.ResponseWriter, r *http.Request) {

	var item *DetailedItem
	itemID := c.URLParams["item_id"]

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic("Unable to connect the DB.")
	}
	defer db.Close()

	item, err = SelectItemByID(db, itemID)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(*item)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}
	return

}

func itemsPost(c web.C, w http.ResponseWriter, r *http.Request) {

	var user *User

	contentType := r.Header.Get("Content-type")
	if utils.ContentTypeEqualJSON(contentType) == false {
		utils.SetStatus(w, 400)
		return
	}

	jsonReq := utils.JSONToMap(w, r)
	if jsonReq == nil {
		utils.SetStatus(w, 400)
		return
	}

	title, _ := jsonReq["title"].(string)

	// Is title empty?
	if utf8.RuneCountInString(title) <= 0 {
		utils.SetStatus(w, 400)
		return
	}

	body, _ := jsonReq["body"].(string)
	// Is body empty?
	if utf8.RuneCountInString(body) <= 0 {
		utils.SetStatus(w, 400)
		return
	}

	session, err := store.Get(r, "session")
	if err != nil {
		utils.SetStatus(w, 401)
		return
	}
	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to connect the DB.")
	}

	// get user ID

	username, _ := session.Values["username"].(string)
	user, err = SelectUserByUsername(db, username)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}
	userID := user.ID

	// create new item.

	const layout = "2006-01-02 15:04:05"
	t := time.Now()
	query := "INSERT INTO items (user_id, title, body, created_at, updated_at) VALUES ((?), (?), (?), (?), (?));"
	_, err = db.Exec(query, userID, title, body, t.Format(layout), t.Format(layout))

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to insert to items table.")
		return
	}

	// get new item

	var item DetailedItem
	query = "SELECT id, user_id, title, body, created_at, updated_at, likes from items WHERE user_id=(?) AND created_at=(?)"
	rows, err := db.Query(query, userID, t.Format(layout))
	if err != nil || rows.Next() == false {
		utils.SetStatus(w, 500)
		panic("Unable to get the query results.")
		return
	}
	defer rows.Close()
	err = rows.Scan(&item.ID, &item.userID, &item.Title, &item.Body, &item.CreatedAt, &item.UpdatedAt, &item.likes)
	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to scan from the result.")
		return
	}

	item.Username = username

	if item.likes.Valid {
		item.Likes = item.likes.String
	}

	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(item)
	utils.SetStatus(w, 201)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	return
}

func itemsDelete(c web.C, w http.ResponseWriter, r *http.Request) {
	var user *User
	itemID := c.URLParams["item_id"]
	session, err := store.Get(r, "session")

	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		panic("Unable to connect the DB.")
	}

	var item *DetailedItem
	item, err = SelectItemByID(db, itemID)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
		return
	}

	username, _ := session.Values["username"].(string)
	user, err = SelectUserByUsername(db, username)
	if item.userID != user.ID {
		utils.SetStatus(w, 403)
		return
	}

	fmt.Println(item.ID)

	query := "DELETE FROM items WHERE id=(?)"
	res, err := db.Exec(query, itemID)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to delete item.")
		return
	}

	num, _ := res.RowsAffected()

	if num <= 0 {
		utils.SetStatus(w, 500)
		panic("Unable to delete item.")
		return
	}

	utils.SetStatus(w, 204)
}

func itemsPatch(c web.C, w http.ResponseWriter, r *http.Request) {
	var user *User
	var item *DetailedItem
	itemID := c.URLParams["item_id"]

	var titleFlg, bodyFlg bool
	jsonReq := utils.JSONToMap(w, r)

	if jsonReq == nil {
		utils.SetStatus(w, 400)
		return
	}

	title, _ := jsonReq["title"].(string)

	// Want to change the title?

	if utf8.RuneCountInString(title) > 0 {
		titleFlg = true
	}

	// Want to change the body?

	body, _ := jsonReq["body"].(string)
	if utf8.RuneCountInString(body) > 0 {
		bodyFlg = true
	}

	if titleFlg == false && bodyFlg == false {
		utils.SetStatus(w, 400)
		return
	}

	// If doesn't have session

	session, err := store.Get(r, "session")
	if err != nil || session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to connect the DB.")
	}

	// get user ID

	username, _ := session.Values["username"].(string)
	user, err = SelectUserByUsername(db, username)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}
	userID := user.ID

	//get current item

	item, err = SelectItemByID(db, itemID)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}

	if userID != item.userID {
		utils.SetStatus(w, 403)
		return
	}

	// update item.

	if titleFlg {
		item.Title = title
	}

	if bodyFlg {
		item.Body = body
	}

	const layout = "2006-01-02 15:04:05"
	t := time.Now()
	query := "UPDATE items SET  title=(?), body=(?), updated_at=(?) WHERE id=(?);"
	_, err = db.Exec(query, item.Title, item.Body, t.Format(layout), itemID)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to insert to items table.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.SetStatus(w, 200)
	result, err := json.Marshal(item)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}
	return

}

func commentsGet(c web.C, w http.ResponseWriter, r *http.Request) {
	var comments *Comments
	itemID := c.URLParams["item_id"]

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to connect the DB.")
	}

	if ItemExists(itemID, db) == false {
		utils.SetStatus(w, 404)
		return
	}

	comments, err = SelectCommentsByID(db, itemID)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			// If there are no comments.
			w.Header().Set("Content-Type", "application/json")

			id, _ := strconv.Atoi(itemID)
			// create empty response.
			comments = &Comments{[]Comment{}, [100]sql.NullString{}, id}

			result, err := json.Marshal(comments)

			if err != nil {
				utils.SetStatus(w, 500)
				panic("Failed to convert to JSON.")
				return
			}

			_, err = w.Write(result)

			return
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		case DBCONVERR:
			panic("Failed to convert to struct.")
		default:
			panic(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	utils.SetStatus(w, 200)
	result, err := json.Marshal(*comments)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}
	return

}
func commentsPost(c web.C, w http.ResponseWriter, r *http.Request) {
	var comments *Comments
	itemID := c.URLParams["item_id"]

	jsonReq := utils.JSONToMap(w, r)

	if jsonReq == nil {
		utils.SetStatus(w, 400)
		return
	}

	commentStr, _ := jsonReq["comment"].(string)

	// Is commentStr empty?
	if utf8.RuneCountInString(commentStr) <= 0 {
		utils.SetStatus(w, 400)
		return
	}

	session, err := store.Get(r, "session")
	if err != nil {
		utils.SetStatus(w, 401)
		return
	}

	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to connect the DB.")
	}

	if ItemExists(itemID, db) == false {
		utils.SetStatus(w, 404)
		return
	}

	// get user ID
	username, _ := session.Values["username"].(string)
	user, err := SelectUserByUsername(db, username)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 500)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}
	userID := user.ID

	var emptyFlg = false
	comments, err = SelectCommentsByID(db, itemID)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			emptyFlg = true
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		case DBCONVERR:
			panic("Failed to convert to struct.")
		default:
			panic(err)
		}
	}

	var newComment Comment
	id, _ := strconv.Atoi(itemID)
	if emptyFlg { //insert new comment

		newComment = Comment{username, id, userID, 1, commentStr}
		jsonByte, _ := json.Marshal(newComment)
		jsonStr := string(jsonByte)
		query := "INSERT into comments(comment_001, id) value  (?, ?)"
		_, err = db.Exec(query, jsonStr, itemID)
		if err != nil {
			panic(err)
		}
	} else { //update empty comment_xxx column

		// search empty column
		for i := 0; i < 100; i++ {
			if comments.jsonRAW[i].Valid == true {
				continue
			}
			newComment = Comment{username, id, userID, i + 1, commentStr}
			comments.Comments = append(comments.Comments, newComment)
			break
		}
		// update colmun
		for _, x := range comments.Comments {
			if x.CommentID != newComment.CommentID {
				continue
			}
			jsonByte, _ := json.Marshal(x)
			jsonStr := string(jsonByte)
			fmt.Printf("string(jsonBytes): %s\n", jsonStr)
			query := fmt.Sprintf("UPDATE comments set comment_%03d=(?) WHERE id=(?)", x.CommentID)
			_, err = db.Exec(query, jsonStr, itemID)
			if err != nil {
				panic(err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	utils.SetStatus(w, 201)
	result, err := json.Marshal(newComment)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}
	return

}
func commentsDelete(c web.C, w http.ResponseWriter, r *http.Request) {
	var comments *Comments
	itemID := c.URLParams["item_id"]
	commentID, _ := strconv.Atoi(c.URLParams["comment_id"])

	session, err := store.Get(r, "session")
	if err != nil {
		utils.SetStatus(w, 401)
		return
	}

	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	db, err := sql.Open(dbDriver, dataSource)
	defer db.Close()
	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to connect the DB.")
	}

	if ItemExists(itemID, db) == false {
		utils.SetStatus(w, 404)
		return
	}

	// get user ID

	username, _ := session.Values["username"].(string)
	user, err := SelectUserByUsername(db, username)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 500)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}
	userID := user.ID

	var emptyFlg = false
	comments, err = SelectCommentsByID(db, itemID)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			emptyFlg = true
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		case DBCONVERR:
			panic("Failed to convert to struct.")
		default:
			panic(err)
		}
	}

	if emptyFlg { // there is no comments here
		utils.SetStatus(w, 404)
		return
	}

	// search target column
	deleted := false
	for _, x := range comments.Comments {
		log.Println("comments:", x)
		if x.CommentID != commentID {
			continue
		}
		log.Println("x.userID", x.userID)
		log.Println("userID", userID)
		if x.userID != userID {
			utils.SetStatus(w, 403)
			return
		}
		var query string
		if len(comments.Comments) == 1 { // if this comments is last.
			query = "DELETE from comments WHERE id=(?)"
		} else {
			query = fmt.Sprintf("UPDATE comments set comment_%03d=NULL WHERE id=(?)", x.CommentID)
		}
		_, err = db.Exec(query, itemID)
		if err != nil {
			panic(err)
		}
		deleted = true
		break
	}

	if deleted == false {
		utils.SetStatus(w, 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.SetStatus(w, 204)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}
	return

}
func likeGet(c web.C, w http.ResponseWriter, r *http.Request) {
	var like Likes
	itemID := c.URLParams["item_id"]

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic("Unable to connect the DB.")
	}
	defer db.Close()

	query := "SELECT likes from items WHERE id=(?)"
	rows, err := db.Query(query, itemID)
	if err != nil {
		panic("Unable to get the query results.")
		return
	}
	defer rows.Close()
	if rows.Next() == false {
		utils.SetStatus(w, 404)
		return
	}

	rows.Scan(&like.likes)

	if like.likes.Valid == false {
		like.Likes = ""
	} else {
		like.Likes = like.likes.String
		like.LikeCount = len(strings.Split(like.Likes, ","))
	}

	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(like)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	_, err = w.Write(result)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}
	return

}

func likePost(c web.C, w http.ResponseWriter, r *http.Request) {

	var like Likes
	var user *User

	itemID := c.URLParams["item_id"]
	session, err := store.Get(r, "session")
	if err != nil {
		utils.SetStatus(w, 401)
		return
	}
	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic("Unable to connect the DB.")
	}
	defer db.Close()

	// get user ID

	username, _ := session.Values["username"].(string)
	user, err = SelectUserByUsername(db, username)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
		return
	}

	query := "SELECT likes from items WHERE id=(?)"
	rows, err := db.Query(query, itemID)
	if err != nil {
		panic("Unable to get the query results.")
		return
	}
	defer rows.Close()
	if rows.Next() == false {
		utils.SetStatus(w, 404)
		return
	}

	rows.Scan(&like.likes)

	if like.likes.Valid == false {
		like.Likes = "" + fmt.Sprint(user.ID)
	} else {
		like.Likes = ""
		likeStrs := strings.Split(like.likes.String, ",")

		fmt.Printf("like.likes.String:%v\n", like.likes.String)
		fmt.Printf("likeStrs:%v\n", likeStrs)
		fmt.Printf("user.ID:%v\n", user.ID)

		userID := fmt.Sprint(user.ID)
		// search loop
		for _, s := range likeStrs {
			fmt.Println(s)

			if s == userID {
				continue
			}

			if like.Likes == "" {
				like.Likes = s
			} else {
				like.Likes += "," + s
			}
		}
		// append
		if like.Likes == "" {
			like.Likes = userID
		} else {
			like.Likes += "," + userID
		}
		like.LikeCount = len(likeStrs)
	}

	fmt.Printf("final likes: %s\n", like.Likes)

	query = "UPDATE items set likes=(?) WHERE id=(?)"
	_, err = db.Exec(query, like.Likes, itemID)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to update likes.")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to convert to JSON.")
		return
	}

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to write responses.")
		return
	}

	utils.SetStatus(w, 204)
	return

}

func likeDelete(c web.C, w http.ResponseWriter, r *http.Request) {

	var like Likes
	var user *User

	itemID := c.URLParams["item_id"]

	session, err := store.Get(r, "session")
	if err != nil {
		utils.SetStatus(w, 401)
		return
	}
	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic("Unable to connect the DB.")
	}
	defer db.Close()

	// get user ID

	username, _ := session.Values["username"].(string)
	user, err = SelectUserByUsername(db, username)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
		return
	}

	query := "SELECT likes from items WHERE id=(?)"
	rows, err := db.Query(query, itemID)
	if err != nil {
		panic("Unable to get the query results.")
		return
	}
	defer rows.Close()
	if rows.Next() == false {
		utils.SetStatus(w, 404)
		return
	}

	rows.Scan(&like.likes)

	var result interface{}
	if like.likes.Valid == false {
		utils.SetStatus(w, 404)
		return
	}

	userIDExists := false
	like.Likes = ""
	likeStrs := strings.Split(like.likes.String, ",")
	userID := fmt.Sprint(user.ID)

	// search loop
	for _, s := range likeStrs {
		if s != userID {
			if like.Likes == "" {
				like.Likes = s
			} else {
				like.Likes += "," + s
			}
		} else {
			userIDExists = true
		}
	}

	// If the user does not exist
	if userIDExists == false {
		utils.SetStatus(w, 404)
		return
	}

	// Update result
	if like.Likes == "" {
		result = nil
	} else {
		result = like.Likes
	}

	fmt.Printf("final: %v\n", result)
	query = "UPDATE items set likes=(?) WHERE id=(?)"
	_, err = db.Exec(query, result, itemID)

	if err != nil {
		utils.SetStatus(w, 500)
		panic("Unable to update likes.")
		return
	}

	utils.SetStatus(w, 204)
	return

}

func iconGet(c web.C, w http.ResponseWriter, r *http.Request) {
	username := c.URLParams["username"]

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic("Unable to connect the DB.")
	}
	defer db.Close()

	// get user ID
	user, err := SelectUserByUsername(db, username)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 404)
			return
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}
	userID := user.ID
	iconExists := true
	base64txt, err := SelectIconByUserID(db, userID)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			// There is no icon.
			iconExists = false
			break
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}

	//var data map[string]interface{}
	w.Header().Set("Content-Type", "image/png")

	if iconExists {
		//log.Println("base64txt:", base64txt)
		decodedimg, err := base64.StdEncoding.DecodeString(base64txt)
		if err != nil {
			panic("Decode failed.")
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(decodedimg)))
		w.Write(decodedimg)
	} else {
		absPath, _ := filepath.Abs("./public/img/default_user_icon.png")
		img, err := ioutil.ReadFile(absPath)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(img)))
		w.Write(img)

	}

	return
}

func iconPost(c web.C, w http.ResponseWriter, r *http.Request) {
	username := c.URLParams["username"]

	err := r.ParseMultipartForm(10000000)
	// if err != nil do something

	file, _, err := r.FormFile("iconimage")
	if err != nil {
		utils.SetStatus(w, 400)
		return
	}
	defer file.Close()

	session, err := store.Get(r, "session")
	if err != nil {
		utils.SetStatus(w, 401)
		return
	}
	// If the request doesn't include session
	if session.Values["username"] == nil {
		utils.SetStatus(w, 401)
		return
	}

	//If the request is not yours
	if session.Values["username"] != username {
		utils.SetStatus(w, 403)
		return
	}

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic("Unable to connect the DB.")
	}
	defer db.Close()
	// get user ID
	user, err := SelectUserByUsername(db, username)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			utils.SetStatus(w, 500)
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	}
	userID := user.ID
	base64txt, err := SelectIconByUserID(db, userID)
	if err != nil {
		errCode, _ := strconv.Atoi(fmt.Sprintf("%s", err))
		switch errCode {
		case DBNOTFOUNDERR:
			// There is no icon.
			break
		case DBQUERYERR:
			panic("Unable to get the query results.")
		case DBSCANERR:
			panic("Unable to scan from the result.")
		default:
			panic(err)
		}
	} else {
		log.Println("base64txt:", base64txt)
		// Icon exists.
		utils.SetStatus(w, 409)
		return
	}

	img, err := ioutil.ReadAll(file)
	encodedimg := base64.StdEncoding.EncodeToString([]byte(img))
	log.Printf("base64:%s\n ", encodedimg)

	query := "INSERT INTO icon (user_id, icon) VALUES ((?), (?));"
	_, err = db.Exec(query, userID, encodedimg)
	if err != nil {
		utils.SetStatus(w, 500)
		panic("Failed to insert to items table.")
		return
	}
	utils.SetStatus(w, 201)
	return
}

func main() {
	store = sessions.NewCookieStore(securecookie.GenerateRandomKey(64))
	users = make(map[string]string)

	dbhost := os.Getenv("MYSQL_HOST")
	dbname := os.Getenv("MYSQL_DATABASE")
	dbuser := os.Getenv("MYSQL_USER")
	dbpass := os.Getenv("MYSQL_PASSWORD")
	dataSource = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&interpolateParams=true", dbuser, dbpass, dbhost, dbname)

	const location = "./public"
	absLocation, _ := filepath.Abs(location)
	_ = flag.Set("bind", ":5000")
	goji.Get("/initialize", initialize)
	goji.Post("/signin", signin)
	goji.Get("/signout", signout)
	goji.Get("/users/:username", usersGet)
	goji.Post("/users", usersPost)
	goji.Patch("/users/:username", usersPatch)
	goji.Delete("/users/:username", usersDelete)
	goji.Get("/users/:username/icon", iconGet)
	goji.Post("/users/:username/icon", iconPost)
	goji.Get("/items", itemsGet)
	goji.Get("/items/:item_id", itemsGetByID)
	goji.Post("/items", itemsPost)
	goji.Patch("/items/:item_id", itemsPatch)
	goji.Delete("/items/:item_id", itemsDelete)
	goji.Get("/items/:item_id/likes", likeGet)
	goji.Post("/items/:item_id/likes", likePost)
	goji.Delete("/items/:item_id/likes", likeDelete)
	goji.Get("/items/:item_id/comments", commentsGet)
	goji.Post("/items/:item_id/comments", commentsPost)
	goji.Delete("/items/:item_id/comments/:comment_id", commentsDelete)
	goji.Get("/*", http.FileServer(http.Dir(absLocation)))
	goji.Serve()

}
