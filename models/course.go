package models

import (
	"errors"
)

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
	UpdatedUnix int64  `xorm:"updated"`
}

// AddCourse adds a new Course to the database.
func AddCourse(c *Course) (err error) {
	_, err = engine.Insert(c)
	return err
}

// GetCourses returns a list of all courses in the database.
func GetCourses() (courses []Course) {
	engine.Find(&courses)
	return courses
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
	return c, nil
}

// LoadPosts will load all the posts of the course into a non-mapped field.
func (c *Course) LoadPosts() (err error) {
	err = engine.Where("course_code = ?", c.Code).Find(&c.Posts)
	if err != nil {
		return
	}
	for i := range c.Posts {
		c.Posts[i].Poster, _ = GetUser(c.Posts[i].PosterID)
		c.Posts[i].CommentsCount, _ = engine.Where("post_id = ?", c.Posts[i].PostID).Count(new(Comment))
	}
	return
}
