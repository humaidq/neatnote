package models

import (
	"errors"
	"fmt"
	"git.sr.ht/~humaid/notes-overflow/modules/settings"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq" // PostgreSQL driver support
	"log"
	"xorm.io/core"
)

var (
	engine *xorm.Engine
	tables []interface{}
)

func init() {
	tables = append(tables,
		new(User),
		new(Course),
		new(Post),
		new(Comment),
	)
}

type User struct {
	Username string `xorm:"pk"`
	FullName string `xorm:"text null"`
	IsAdmin  bool   `xorm:"bool"`
}

type Course struct {
	Code    string `xorm:"pk text"`
	Name    string `xorm:"notnull text"`
	Visible bool   `xorm:"notnull"`
	Locked  bool   `xorm:"notnull"`
	Posts   []Post `xorm:"-"`
}

// PostType represents the type of the post submission.
type PostType int

const (
	// Markdown is when markdown is used in the post.
	Markdown = iota
	// PDF is when the post is a PDF upload.
	PDF
)

type Post struct {
	PostID     int64     `xorm:"pk autoincr"`
	CourseCode string    `xorm:"text notnull"`
	Type       PostType  `xorm:"notnull"`
	Locked     bool      `xorm:"notnull"` // Whether the comments are locked.
	Comments   []Comment `xorm:"-"`
	Text       string    `xorm:"null"`
}

type Comment struct {
	CommentID int64  `xorm:"pk autoincr"`
	PostID    int64  `xorm:"notnull"`
	Text      string `xorm:"notnull"`
}

func HasUser(user string) (has bool) {
	has, _ = engine.Get(&User{Username: user})
	return has
}

func AddUser(u *User) (err error) {
	_, err = engine.Insert(u)
	return err
}

func GetUser(user string) (*User, error) {
	u := new(User)
	has, err := engine.ID(user).Get(u)
	if err != nil {
		return u, err
	} else if !has {
		return u, errors.New("Repository does not exist")
	}
	return u, nil
}

// SetupEngine sets up an XORM engine according to the database configuration
// and syncs the schema.
func SetupEngine() *xorm.Engine {
	var err error
	dbConf := &settings.DBConfig

	address := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Name, dbConf.SSLMode)
	engine, err = xorm.NewEngine("postgres", address)

	if err != nil {
		log.Fatal("Unable to connect/load the database! ", err)
	}

	engine.SetMapper(core.GonicMapper{}) // So ID becomes 'id' instead of 'i_d'
	err = engine.Sync(tables...)         // Sync the schema of tables

	if err != nil {
		log.Fatal("Unable to sync schema! ", err)
	}

	return engine
}
