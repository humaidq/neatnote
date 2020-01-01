package models

import (
	"errors"
	"fmt"
	"git.sr.ht/~humaid/neatnote/modules/settings"
	_ "github.com/go-sql-driver/mysql" // MySQL driver support
	"github.com/hako/durafmt"
	_ "github.com/mattn/go-sqlite3" // SQLite driver support
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

// User represents a website user.
// It keeps track of the iota, settings (such as badges), and whether they
// have administrative privileges.
type User struct {
	Username    string `xorm:"pk"`
	FullName    string `xorm:"text null"`
	Badge       string `xorm:"text null"`
	IsAdmin     bool   `xorm:"bool"`
	Iota        int64
	Created     string `xorm:"-"`
	CreatedUnix int64  `xorm:"created"`
	Upvoted     []int64
}

// Course represents a sub-forum on a website, and is defined with a course
// code and a name.
// It keeps track of number of posts, whether that forum is locked, visible,
// and so on.
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

// Post represents a post in one of the Courses by one of the Users.
// It keeps track of the poster, course, comments (and comment count), whether
// it is an anonymous post, and so on.
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
	Anonymous     bool      `xorm:"notnull"`
	AnonName      string    `xorm:"text null"`
}

// Comment represents a comment on a Post. It keeps track of the poster and
// which post it is posted to.
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

// UpdateUser updates a user in the database.
func UpdateUser(u *User) (err error) {
	_, err = engine.Id(u.Username).Update(u)
	return
}

// UpdateUserBadge updates a user in the database including the Badge field,
// even if the field is empty.
func UpdateUserBadge(u *User) (err error) {
	_, err = engine.Id(u.Username).Cols("badge").Update(u)
	return
}

// LoadPoster loads the poster of a comment in the non-mapped field of the
// Comment struct.
func (c *Comment) LoadPoster() (err error) {
	if c == nil {
		return nil
	} else if c.Poster != nil {
		return nil
	}

	c.Poster, err = GetUser(c.PosterID)
	return
}

// LoadCreated loads the created time of a comment in a non-mapped field
// relative to the current time.
func (c *Comment) LoadCreated() (err error) {
	if c == nil {
		return nil
	}

	dur := time.Now().Sub(time.Unix(c.CreatedUnix, 0))
	c.Created = durafmt.Parse(dur).LimitFirstN(1).String()
	return
}

// GetAllUserPosts returns all of the post created by a specific user.
func GetAllUserPosts(user string) (p []Post, err error) {
	err = engine.Where("poster_id = ?", user).Find(&p)
	return
}

// LoadPosts will load all the posts of the course into a non-mapped field.
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

// LoadPostsCount loads the posts count of a course in a non-mapped field.
func (c *Course) LoadPostsCount() (err error) {
	c.PostsCount, err = engine.Where("course_code = ?", c.Code).Count(new(Post))
	return
}

// GetCourse gets a Course based on a course code.
// It will return a pointer to the Course struct, and whether there was an
// error or not.
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

// calcDuration calculates the duration between two timestamps.
// It will return a fancy formatted duration in text.
func calcDuration(unix int64) string {
	return durafmt.Parse(time.Now().Sub(time.Unix(unix, 0))).LimitFirstN(1).String()
}

// LoadComments loads the comments of the post into a non-mapped field.
func (p *Post) LoadComments() (err error) {
	return engine.Where("post_id = ?", p.PostID).Find(&p.Comments)
}

// GetPost gets a Post based on the ID.
// It will return the pointer to the Post, and whether there was an error.
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

// HasUser returns whether a user exists in the database.
func HasUser(user string) (has bool) {
	has, _ = engine.Get(&User{Username: user})
	return has
}

// AddPost adds a new Post to the database.
func AddPost(p *Post) (err error) {
	_, err = engine.Insert(p)
	return err
}

// AddCourse adds a new Course to the database.
func AddCourse(c *Course) (err error) {
	_, err = engine.Insert(c)
	return err
}

// AddUser adds a new User to the database.
func AddUser(u *User) (err error) {
	_, err = engine.Insert(u)
	return err
}

// AddComment adds a new Comment to the database.
func AddComment(c *Comment) (err error) {
	_, err = engine.Insert(c)
	return err
}

// GetCourses returns a list of all courses in the database.
func GetCourses() (courses []Course) {
	engine.Find(&courses)
	return courses
}

// GetUser gets a user based on their username.
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

	switch dbConf.Type {
	case settings.MySQL:
		engine, err = xorm.NewEngine("mysql", address)
	case settings.SQLite:
		engine, err = xorm.NewEngine("sqlite3", dbConf.Path)
	}

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
