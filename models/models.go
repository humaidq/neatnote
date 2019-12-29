package models

import (
	"errors"
	"fmt"
	"git.sr.ht/~humaid/notes-overflow/modules/settings"
	_ "github.com/go-sql-driver/mysql" // MySQL driver support
	"github.com/hako/durafmt"
	"html/template"
	"log"
	"time"
	"xorm.io/core"
	"xorm.io/xorm"
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
	Iota        int64
	Created     string `xorm:"-"`
	CreatedUnix int64  `xorm:"created"`
}

type Course struct {
	Code        string `xorm:"pk varchar(64)"`
	Name        string `xorm:"notnull text"`
	Visible     bool   `xorm:"notnull"`
	Locked      bool   `xorm:"notnull"`
	PostsCount  int64  `xorm:"-"`
	Posts       []Post `xorm:"-"`
	CreatedUnix int64  `xorm:"created"`
	Created     string `xorm:"-"`
	UpdatedUnix int64  `xorm:"updated"`
}

type Post struct {
	PostID        int64     `xorm:"pk autoincr"`
	CourseCode    string    `xorm:"text notnull"`
	PosterID      string    `xorm:"notnull"`
	Poster        *User     `xorm:"-"`
	Locked        bool      `xorm:"notnull"` // Whether the comments are locked.
	Comments      []Comment `xorm:"-"`
	CommentsCount int64     `xorm:"-"`
	Title         string    `xorm:"text notnull"`
	Text          string    `xorm:"text notnull"`
	CreatedUnix   int64     `xorm:"created"`
	Created       string    `xorm:"-"`
	UpdatedUnix   int64     `xorm:"updated"`
}

type Comment struct {
	CommentID     int64         `xorm:"pk autoincr"`
	PostID        int64         `xorm:"notnull"`
	PosterID      string        `xorm:"notnull"`
	Poster        *User         `xorm:"-"`
	Text          string        `xorm:"notnull"`
	FormattedText template.HTML `xorm:"-"`
	CreatedUnix   int64         `xorm:"created"`
	Created       string        `xorm:"-"`
	UpdatedUnix   int64         `xorm:"updated"`
}

func UpdateUser(u *User) (err error) {
	_, err = engine.Update(u)
	return
}

func (c *Comment) LoadPoster() (err error) {
	if c == nil {
		return nil
	} else if c.Poster != nil {
		return nil
	}

	c.Poster, err = GetUser(c.PosterID)
	return
}

func (c *Comment) LoadCreated() (err error) {
	if c == nil {
		return nil
	}

	dur := time.Now().Sub(time.Unix(c.CreatedUnix, 0))
	c.Created = durafmt.Parse(dur).LimitFirstN(1).String()
	return
}

func GetAllUserPosts(user string) (p []Post, err error) {
	err = engine.Where("poster_id = ?", user).Find(&p)
	return
}

func (c *Course) LoadPosts() (err error) {
	err = engine.Where("course_code = ?", c.Code).Find(&c.Posts)
	if err != nil {
		return
	}
	for i := range c.Posts {
		c.Posts[i].Poster, _ = GetUser(c.Posts[i].PosterID)
		c.Posts[i].Created = calcDuration(c.Posts[i].CreatedUnix)
		c.Posts[i].CommentsCount, _ = engine.Where("post_id = ?", c.Posts[i].PostID).Count(new(Comment))
	}
	return
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
	c.Created = calcDuration(c.CreatedUnix)
	return c, nil
}

func calcDuration(unix int64) string {
	return durafmt.Parse(time.Now().Sub(time.Unix(unix, 0))).LimitFirstN(1).String()
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
	p.Created = calcDuration(p.CreatedUnix)
	p.Poster, _ = GetUser(p.PosterID)
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
	u.Created = calcDuration(u.CreatedUnix)
	return u, nil
}

// SetupEngine sets up an XORM engine according to the database configuration
// and syncs the schema.
func SetupEngine() *xorm.Engine {
	var err error
	dbConf := &settings.Config.DBConfig

	address := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Name)
	engine, err = xorm.NewEngine("mysql", address)

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
