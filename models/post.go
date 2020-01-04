package models

import (
	"errors"
	"git.sr.ht/~humaid/neatnote/modules/common"
)

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
	Iota          int64
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

// UpdatePost updates a post in the database.
func UpdatePost(p *Post) (err error) {
	_, err = engine.Id(p.PostID).Update(p)
	return
}

// AddPost adds a new Post to the database.
func AddPost(p *Post) (err error) {
	_, err = engine.Insert(p)
	return err
}

// GetAllUserPosts returns all of the post created by a specific user.
func GetAllUserPosts(user string) (p []Post, err error) {
	err = engine.Where("poster_id = ?", user).Find(&p)
	return
}

func UnvotePost(user string, post int64) (err error) {
	sess := engine.NewSession()
	if err = sess.Begin(); err != nil {
		return err
	}
	u := new(User)
	var has bool
	if has, err = sess.ID(user).Get(u); err != nil {
		return err
	} else if !has {
		return errors.New("User does not exist.")
	}

	if !common.ContainsInt64(u.Upvoted, post) {
		return errors.New("Cannot unvote a post which is not upvoted.")
	}

	p := new(Post)
	has, err = sess.ID(post).Get(p)
	if has, err = sess.ID(user).Get(u); err != nil {
		return err
	} else if !has {
		return errors.New("Post does not exist.")
	}

	_, err = sess.Id(u.Username).Cols("upvoted").Update(&User{
		Username: u.Username,
		Upvoted:  common.RemoveInt64(u.Upvoted, post),
	})
	if err != nil {
		return err
	}

	_, err = sess.ID(post).Cols("iota").Update(&Post{
		PostID: post,
		Iota:   p.Iota - 1,
	})
	if err != nil {
		return err
	}

	return sess.Commit()
}

func UpvotePost(user string, post int64) (err error) {
	sess := engine.NewSession()
	if err = sess.Begin(); err != nil {
		return err
	}
	u := new(User)
	var has bool
	if has, err = sess.ID(user).Get(u); err != nil {
		return err
	} else if !has {
		return errors.New("User does not exist.")
	}

	if common.ContainsInt64(u.Upvoted, post) {
		return errors.New("Already upvoted")
	}

	p := new(Post)
	has, err = sess.ID(post).Get(p)
	if has, err = sess.ID(user).Get(u); err != nil {
		return err
	} else if !has {
		return errors.New("Post does not exist.")
	}

	_, err = sess.Id(u.Username).Cols("upvoted").Update(&User{
		Username: u.Username,
		Upvoted:  append(u.Upvoted, post),
	})
	if err != nil {
		return err
	}

	_, err = sess.ID(post).Cols("iota").Update(&Post{
		PostID: post,
		Iota:   p.Iota + 1,
	})
	if err != nil {
		return err
	}

	return sess.Commit()
}
