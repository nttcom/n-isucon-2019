package main

import "database/sql"

// User is structure
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	passwordHash string
	salt         string
}

// Item is structure
type Item struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	Username  string `json:"username"`
	likes     sql.NullString
	likeCount int
}

// Items is structure
type Items struct {
	Items []Item `json:"items"`
	Count int    `json:"count"`
}

// DetailedItem is structure
type DetailedItem struct {
	ID        int `json:"id"`
	userID    int
	Username  string `json:"username"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Likes     string `json:"likes"`
	likes     sql.NullString
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Likes is structure
type Likes struct {
	Likes     string `json:"likes"`
	LikeCount int    `json:"like_count"`
	id        int
	likes     sql.NullString
}

// Comment is structure
type Comment struct {
	Username  string `json:"username"`
	ID        int    `json:"item_id"`
	userID    int
	CommentID int    `json:"comment_id"`
	Comment   string `json:"comment"`
}

// Comments is structure
type Comments struct {
	Comments []Comment `json:"comments"`
	jsonRAW  [100]sql.NullString
	ItemID   int `json:"item_id"`
}
