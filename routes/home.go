package routes

import (
	"git.sr.ht/~humaid/neatnote/models"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

// HomepageHandler response for the home page.
func HomepageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	courses := models.GetCourses()
	for i := range courses {
		courses[i].LoadPostsCount()
	}
	ctx.Data["Courses"] = courses
	ctx.HTML(200, "index")
}

// QnAHandler response for the Questions and Answers page.
func QnAHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	ctx.HTML(200, "qna")
}

// GuidelinesHandler response for the Guidelines page.
func GuidelinesHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	ctx.HTML(200, "guidelines")
}
