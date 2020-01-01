package routes

import (
	"git.sr.ht/~humaid/neatnote/models"
	"git.sr.ht/~humaid/neatnote/modules/settings"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

const (
	// LoggedOut is when a user is logged out.
	LoggedOut = iota
	// Verification is when a user is in the verification process.
	Verification
	// LoggedIn is when the user is verified and logged in.
	LoggedIn
)

// ctxInit initialises the context using the session for every page.
// This handles verifying the login status and setting some global
// template variables.
func ctxInit(ctx *macaron.Context, sess session.Store) {
	if sess.Get("auth") == nil {
		sess.Set("auth", LoggedOut)
	}
	if sess.Get("auth") == LoggedIn {
		ctx.Data["LoggedIn"] = 1
		ctx.Data["Username"] = sess.Get("user")
		if user, err := models.GetUser(sess.Get("user").(string)); err == nil {
			ctx.Data["User"] = user
		} else {
			// Let's log out the user
			ctx.Data["LoggedIn"] = 0
			sess.Set("auth", LoggedOut)
		}
	}
	ctx.Data["UniEmailDomain"] = settings.Config.UniEmailDomain
	if settings.Config.DevMode {
		ctx.Data["DevMode"] = 1
	}
	ctx.Data["AvailableBadges"] = append(settings.Config.Badges, "None")
	ctx.Data["SiteTitle"] = settings.Config.SiteName
}
