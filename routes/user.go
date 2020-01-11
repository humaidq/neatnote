package routes

import (
	"git.sr.ht/~humaid/neatnote/models"
	"git.sr.ht/~humaid/neatnote/modules/settings"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

// ProfileHandler response for the profile page.
func ProfileHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["Title"] = "Profile"
	ctx.HTML(200, "profile")
}

func containsStringArray(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}
	return false
}

// PostProfileHandler post response for the profile page.
func PostProfileHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	fname := ctx.QueryTrim("fullname")

	if !simpleTextExp.Match([]byte(fname)) || len(fname) > 32 || len(fname) < 1 {
		f.Error("Your display name must only contain alphabet, numbers, and spaces. And cannot be over 32 characters.")
		ctx.Redirect("/profile")
		return
	}

	var badge string
	if ctx.Query("badge") != "None" {
		if containsStringArray(settings.Config.Badges, ctx.Query("badge")) {
			badge = ctx.Query("badge")
		} else {
			f.Error("Invalid badge selection, make sure you select a valid option.")
			ctx.Redirect("/profile")
			return
		}
	}

	u := &models.User{
		Username: sess.Get("user").(string),
		FullName: fname,
		Badge:    badge,
	}

	err := models.UpdateUserCols(u, "badge")
	if err != nil {
		panic(err)
	}
	err = models.UpdateUser(u)
	if err != nil {
		panic(err)
	}

	ctx.Redirect("/profile")
}

// PostDataHandler post response for requesting data (GDPR compliance).
func PostDataHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	u, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		panic(err)
	}
	var p []models.Post
	p, err = models.GetAllUserPosts(sess.Get("user").(string))
	if err != nil {
		panic(err)
	}

	ctx.JSON(200, map[string]interface{}{
		"user":  u,
		"posts": p,
	})
}
