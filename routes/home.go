package routes

import (
	"git.sr.ht/~humaid/neatnote/models"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

func HomepageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	courses := models.GetCourses()
	for i := range courses {
		courses[i].LoadPostsCount()
	}
	ctx.Data["Courses"] = courses
	ctx.HTML(200, "index")
}

func QnAHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	ctx.HTML(200, "qna")
}

func GuidelinesHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	ctx.HTML(200, "guidelines")
}
