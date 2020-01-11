package routes

import (
	"git.sr.ht/~humaid/neatnote/models"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

// AdminAddCourseHandler response for adding a new course.
func AdminAddCourseHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["Title"] = "Add course"
	ctx.HTML(200, "admin/add-course")
}

// AdminPostAddCourseHandler post response for adding a new course.
func AdminPostAddCourseHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	courseCode := ctx.QueryTrim("coursecode")
	courseName := ctx.QueryTrim("coursename")

	// Check if course exists already
	if len(courseCode) < 1 || len(courseName) < 1 {
		f.Error("You must specify course code and name!")
		ctx.Redirect("/a/addcourse")
		return
	} else if _, err1 := models.GetCourse(courseCode); err1 == nil {
		f.Error("Course already exists!")
		ctx.Redirect("/a/addcourse")
		return
	}

	models.AddCourse(&models.Course{
		Code:    courseCode,
		Name:    courseName,
		Visible: true,
		Locked:  false,
	})

	ctx.Redirect("/")
}
