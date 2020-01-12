package models

import (
	"errors"
	"html/template"
)

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
	UpdatedUnix   int64         `xorm:"updated"`
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

// AddComment adds a new Comment to the database.
func AddComment(c *Comment) (err error) {
	_, err = engine.Insert(c)
	return err
}

// UpdateComment updates a comment in the database.
func UpdateComment(c *Comment) (err error) {
	_, err = engine.Id(c.CommentID).Update(c)
	return
}

// GetComment gets a comment based on the ID.
// It will return the pointer to the Comment, and whether there was an error.
func GetComment(id string) (*Comment, error) {
	c := new(Comment)
	has, err := engine.ID(id).Get(c)
	if err != nil {
		return c, err
	} else if !has {
		return c, errors.New("Comment does not exist")
	}
	return c, nil
}

// DeleteComment deletes a comment from the database.
func DeleteComment(id string) (err error) {
	_, err = engine.Id(id).Delete(&Comment{})
	return
}
