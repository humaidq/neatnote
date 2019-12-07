package models

import (
	"errors"
	"fmt"
	"git.sr.ht/~humaid/notes-overflow/modules/settings"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq" // PostgreSQL driver support
	"html/template"
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
	Username    string `xorm:"pk"`
	FullName    string `xorm:"text null"`
	IsAdmin     bool   `xorm:"bool"`
	CreatedUnix int64  `xorm:"created"`
}

type Course struct {
	Code        string `xorm:"pk text"`
	Name        string `xorm:"notnull text"`
	Visible     bool   `xorm:"notnull"`
	Locked      bool   `xorm:"notnull"`
	PostsCount  int64  `xorm:"-"`
	Posts       []Post `xorm:"-"`
	CreatedUnix int64  `xorm:"created"`
	UpdatedUnix int64  `xorm:"updated"`
}

type Post struct {
	PostID      int64     `xorm:"pk autoincr"`
	CourseCode  string    `xorm:"text notnull"`
	Poster      string    `xorm:"notnull"`
	Locked      bool      `xorm:"notnull"` // Whether the comments are locked.
	Comments    []Comment `xorm:"-"`
	Title       string    `xorm:"text notnull"`
	Text        string    `xorm:"text notnull"`
	CreatedUnix int64     `xorm:"created"`
	UpdatedUnix int64     `xorm:"updated"`
}

type Comment struct {
	CommentID     int64         `xorm:"pk autoincr"`
	PostID        int64         `xorm:"notnull"`
	Poster        string        `xorm:"notnull"`
	Text          string        `xorm:"notnull"`
	FormattedText template.HTML `xorm:"-"`
	CreatedUnix   int64         `xorm:"created"`
	UpdatedUnix   int64         `xorm:"updated"`
}

func (c *Course) LoadPosts() (err error) {
	return engine.Where("course_code = ?", c.Code).Find(&c.Posts)
}

func (c *Course) LoadPostsCount() (err error) {
	c.PostsCount, err = engine.Where("course_code = ?", c.Code).Count(new(Post))
	return
}

func GetCourse(code string) (*Course, error) {
	c := new(Course)
	has, err := engine.ID(code).Get(c)
	if err != nil {
		return c, err
	} else if !has {
		return c, errors.New("Course does not exist")
	}
	return c, nil
}

func (p *Post) LoadComments() (err error) {
	return engine.Where("post_id = ?", p.PostID).Find(&p.Comments)
}

func GetPost(id string) (*Post, error) {
	p := new(Post)
	has, err := engine.ID(id).Get(p)
	if err != nil {
		return p, err
	} else if !has {
		return p, errors.New("Post does not exist")
	}
	return p, nil
}

func HasUser(user string) (has bool) {
	has, _ = engine.Get(&User{Username: user})
	return has
}

func AddPost(p *Post) (err error) {
	_, err = engine.Insert(p)
	return err
}
func AddCourse(c *Course) (err error) {
	_, err = engine.Insert(c)
	return err
}
func AddUser(u *User) (err error) {
	_, err = engine.Insert(u)
	return err
}

func AddComment(c *Comment) (err error) {
	_, err = engine.Insert(c)
	return err
}

func GetCourses() (courses []Course) {
	engine.Find(&courses)
	return courses
}

func GetUser(user string) (*User, error) {
	u := new(User)
	has, err := engine.ID(user).Get(u)
	if err != nil {
		return u, err
	} else if !has {
		return u, errors.New("User does not exist")
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
