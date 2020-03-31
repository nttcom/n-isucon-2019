package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

// DBUtilError is Custom err structure.
type DBUtilError struct {
	Msg  string
	Code int
}

const (
	// DBNOTFOUNDERR is set if the result is empty.
	DBNOTFOUNDERR int = iota
	// DBQUERYERR is set when query execution fails.
	DBQUERYERR
	// DBSCANERR is set when scanning fails.
	DBSCANERR
	// DBCONVERR is set when failed to unmarshal.
	DBCONVERR
)

func (err *DBUtilError) Error() string {
	return fmt.Sprintf("%d", err.Code)
}

// SelectUserByUsername searches the user from users table by username.
func SelectUserByUsername(db *sql.DB, searchname string) (user *User, err error) {
	user = new(User)

	query := "SELECT id, username, password_hash, salt, created_at, updated_at FROM users WHERE username=(?)"
	rows, err := db.Query(query, searchname)
	if err != nil {
		err := &DBUtilError{Msg: "DBQUERYERR", Code: DBQUERYERR}
		return user, err
	}
	defer rows.Close()

	if rows.Next() == false {
		err := &DBUtilError{Msg: "DBNOTFOUNDERR", Code: DBNOTFOUNDERR}
		return user, err
	}

	err = rows.Scan(&user.ID, &user.Username, &user.passwordHash, &user.salt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		err := &DBUtilError{Msg: "DBSCANERR", Code: DBSCANERR}
		return user, err
	}

	return user, err
}

// SelectUserByUserID searches the user from users table by user_id.
func SelectUserByUserID(db *sql.DB, userID int) (user *User, err error) {
	user = new(User)

	query := "SELECT id, username, password_hash, salt, created_at, updated_at FROM users WHERE id=(?)"
	rows, err := db.Query(query, userID)
	if err != nil {
		err := &DBUtilError{Msg: "DBQUERYERR", Code: DBQUERYERR}
		return user, err
	}

	defer rows.Close()

	if rows.Next() == false {
		err := &DBUtilError{Msg: "DBNOTFOUNDERR", Code: DBNOTFOUNDERR}
		return user, err
	}

	err = rows.Scan(&user.ID, &user.Username, &user.passwordHash, &user.salt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		err := &DBUtilError{Msg: "DBSCANERR", Code: DBSCANERR}
		return user, err
	}

	return user, err
}

// UsernameExists returns "true" if username is already in use.
func UsernameExists(username string, db *sql.DB) bool {

	query := "SELECT username FROM users WHERE username=(?)"
	rows, err := db.Query(query, username)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var searchResult string
	if rows.Next() {
		rows.Scan(&searchResult)
	}

	// if exist
	if searchResult == username {
		return true
	}
	return false
}

// ItemExists returns "true" if the item Exists.
func ItemExists(itemID string, db *sql.DB) bool {
	query := "SELECT id FROM items WHERE id=(?)"
	rows, err := db.Query(query, itemID)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	return rows.Next()
}

// SelectItemByID searches the item from items table by the item ID.
func SelectItemByID(db *sql.DB, itemID string) (item *DetailedItem, err error) {
	item = new(DetailedItem)
	query := "SELECT id, user_id, title, body, created_at, updated_at, likes from items WHERE id=(?)"
	rows, err := db.Query(query, itemID)
	if err != nil {
		err := &DBUtilError{Msg: "DBQUERYERR", Code: DBQUERYERR}
		return item, err
	}
	defer rows.Close()

	if rows.Next() == false {
		err := &DBUtilError{Msg: "DBNOTFOUNDERR", Code: DBNOTFOUNDERR}
		return item, err
	}

	err = rows.Scan(&item.ID, &item.userID, &item.Title, &item.Body, &item.CreatedAt, &item.UpdatedAt, &item.likes)
	if err != nil {
		err := &DBUtilError{Msg: "DBSCANERR", Code: DBSCANERR}
		return item, err
	}

	if item.likes.Valid {
		item.Likes = item.likes.String
	}

	var user *User
	user, err = SelectUserByUserID(db, item.userID)
	if err != nil {
		panic(err)
		return
	}
	item.Username = user.Username

	return item, err
}

// SelectCommentsByID searches the comment from comments table by the item ID.
func SelectCommentsByID(db *sql.DB, itemID string) (comments *Comments, err error) {
	comments = new(Comments)

	query := "SELECT comment_001, comment_002, comment_003, comment_004, comment_005, comment_006, comment_007, comment_008, comment_009, comment_010, comment_011, comment_012, comment_013, comment_014, comment_015, comment_016, comment_017, comment_018, comment_019, comment_020, comment_021, comment_022, comment_023, comment_024, comment_025, comment_026, comment_027, comment_028, comment_029, comment_030, comment_031, comment_032, comment_033, comment_034, comment_035, comment_036, comment_037, comment_038, comment_039, comment_040, comment_041, comment_042, comment_043, comment_044, comment_045, comment_046, comment_047, comment_048, comment_049, comment_050, comment_051, comment_052, comment_053, comment_054, comment_055, comment_056, comment_057, comment_058, comment_059, comment_060, comment_061, comment_062, comment_063, comment_064, comment_065, comment_066, comment_067, comment_068, comment_069, comment_070, comment_071, comment_072, comment_073, comment_074, comment_075, comment_076, comment_077, comment_078, comment_079, comment_080, comment_081, comment_082, comment_083, comment_084, comment_085, comment_086, comment_087, comment_088, comment_089, comment_090, comment_091, comment_092, comment_093, comment_094, comment_095, comment_096, comment_097, comment_098, comment_099, comment_100 FROM comments WHERE id=(?)"
	rows, err := db.Query(query, itemID)

	if err != nil {
		err := &DBUtilError{Msg: "DBQUERYERR", Code: DBQUERYERR}
		return comments, err
	}
	defer rows.Close()

	if rows.Next() == false {
		err := &DBUtilError{Msg: "DBNOTFOUNDERR", Code: DBNOTFOUNDERR}
		return comments, err
	}

	err = rows.Scan(&comments.jsonRAW[0], &comments.jsonRAW[1], &comments.jsonRAW[2], &comments.jsonRAW[3], &comments.jsonRAW[4], &comments.jsonRAW[5], &comments.jsonRAW[6], &comments.jsonRAW[7], &comments.jsonRAW[8], &comments.jsonRAW[9], &comments.jsonRAW[10], &comments.jsonRAW[11], &comments.jsonRAW[12], &comments.jsonRAW[13], &comments.jsonRAW[14], &comments.jsonRAW[15], &comments.jsonRAW[16], &comments.jsonRAW[17], &comments.jsonRAW[18], &comments.jsonRAW[19], &comments.jsonRAW[20], &comments.jsonRAW[21], &comments.jsonRAW[22], &comments.jsonRAW[23], &comments.jsonRAW[24], &comments.jsonRAW[25], &comments.jsonRAW[26], &comments.jsonRAW[27], &comments.jsonRAW[28], &comments.jsonRAW[29], &comments.jsonRAW[30], &comments.jsonRAW[31], &comments.jsonRAW[32], &comments.jsonRAW[33], &comments.jsonRAW[34], &comments.jsonRAW[35], &comments.jsonRAW[36], &comments.jsonRAW[37], &comments.jsonRAW[38], &comments.jsonRAW[39], &comments.jsonRAW[40], &comments.jsonRAW[41], &comments.jsonRAW[42], &comments.jsonRAW[43], &comments.jsonRAW[44], &comments.jsonRAW[45], &comments.jsonRAW[46], &comments.jsonRAW[47], &comments.jsonRAW[48], &comments.jsonRAW[49], &comments.jsonRAW[50], &comments.jsonRAW[51], &comments.jsonRAW[52], &comments.jsonRAW[53], &comments.jsonRAW[54], &comments.jsonRAW[55], &comments.jsonRAW[56], &comments.jsonRAW[57], &comments.jsonRAW[58], &comments.jsonRAW[59], &comments.jsonRAW[60], &comments.jsonRAW[61], &comments.jsonRAW[62], &comments.jsonRAW[63], &comments.jsonRAW[64], &comments.jsonRAW[65], &comments.jsonRAW[66], &comments.jsonRAW[67], &comments.jsonRAW[68], &comments.jsonRAW[69], &comments.jsonRAW[70], &comments.jsonRAW[71], &comments.jsonRAW[72], &comments.jsonRAW[73], &comments.jsonRAW[74], &comments.jsonRAW[75], &comments.jsonRAW[76], &comments.jsonRAW[77], &comments.jsonRAW[78], &comments.jsonRAW[79], &comments.jsonRAW[80], &comments.jsonRAW[81], &comments.jsonRAW[82], &comments.jsonRAW[83], &comments.jsonRAW[84], &comments.jsonRAW[85], &comments.jsonRAW[86], &comments.jsonRAW[87], &comments.jsonRAW[88], &comments.jsonRAW[89], &comments.jsonRAW[90], &comments.jsonRAW[91], &comments.jsonRAW[92], &comments.jsonRAW[93], &comments.jsonRAW[94], &comments.jsonRAW[95], &comments.jsonRAW[96], &comments.jsonRAW[97], &comments.jsonRAW[98], &comments.jsonRAW[99])
	if err != nil {
		err := &DBUtilError{Msg: "DBSCANERR", Code: DBSCANERR}
		return comments, err
	}

	comments.ItemID, _ = strconv.Atoi(itemID)

	for _, x := range comments.jsonRAW {
		if x.Valid == false {
			continue
		}
		comment := new(Comment)
		err := json.Unmarshal([]byte(x.String), comment)
		if err != nil {
			err := &DBUtilError{Msg: "DBCONVERR", Code: DBCONVERR}
			return comments, err
		}
		user, err := SelectUserByUsername(db, comment.Username)
		if err != nil {
			panic("Unexpected err.")
		}

		comment.userID = user.ID
		comments.Comments = append(comments.Comments, *comment)
	}

	return comments, err
}

// SelectIconByUserID searches the comment from comments table by the user ID.
func SelectIconByUserID(db *sql.DB, userID int) (base64txt string, err error) {
	base64txt = ""

	query := "SELECT icon FROM icon WHERE user_id=(?)"
	rows, err := db.Query(query, userID)

	if err != nil {
		err := &DBUtilError{Msg: "DBQUERYERR", Code: DBQUERYERR}
		return base64txt, err
	}
	defer rows.Close()

	if rows.Next() == false {
		err := &DBUtilError{Msg: "DBNOTFOUNDERR", Code: DBNOTFOUNDERR}
		return base64txt, err
	}

	err = rows.Scan(&base64txt)
	if err != nil {
		err := &DBUtilError{Msg: "DBSCANERR", Code: DBSCANERR}
		return base64txt, err
	}

	return base64txt, err
}
