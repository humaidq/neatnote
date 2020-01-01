package routes

import (
	"git.sr.ht/~humaid/neatnote/models"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

func AdminAddCourseHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("You are not authorised to do that!")
		ctx.Redirect("/login")
		return
	}

	user, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		panic(err)
	}

	if !user.IsAdmin {
		f.Error("You are not authorised to do that!")
		ctx.Redirect("/")
		return
	}

	ctx.Data["csrf_token"] = x.GetToken()

	ctx.HTML(200, "admin/add-course")
}

func AdminPostAddCourseHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("You are not authorised to do that!")
		ctx.Redirect("/login")
		return
	}

	user, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		panic(err)
	}

	if !user.IsAdmin {
		f.Error("You are not authorised to do that!")
		ctx.Redirect("/")
		return
	}

	courseCode := ctx.QueryTrim("coursecode")
	courseName := ctx.QueryTrim("coursename")

	// Check if course exists already
	if len(courseCode) < 1 || len(courseName) < 1 {
		f.Error("You must specify course code and name!")
		ctx.Redirect("/admin/add_course")
		return
	} else if _, err1 := models.GetCourse(courseCode); err1 == nil {
		f.Error("Course already exists!")
		ctx.Redirect("/admin/add_course")
		return
	}

	models.AddCourse(&models.Course{
		Code:    courseCode,
		Name:    courseName,
		Visible: true,
		Locked:  false,
	})

	ctx.Redirect("/?add=1")
}
