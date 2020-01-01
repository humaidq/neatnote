package routes

import (
	"fmt"
	"git.sr.ht/~humaid/neatnote/models"
	"git.sr.ht/~humaid/neatnote/modules/mailer"
	"git.sr.ht/~humaid/neatnote/modules/settings"
	"github.com/badoux/checkmail"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
	"math/rand"
	"strings"
)

// LogoutHandler response for logging out.
func LogoutHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	if sess.Get("auth") != LoggedIn {
		f.Info("You are already logged out!")
		ctx.Redirect("/")
		return
	}
	sess.Set("auth", LoggedOut)
	//sess.Flush()
	ctx.Redirect("/")
}

// LoginHandler response for logging in page.
func LoginHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == Verification {
		ctx.Redirect("/verify")
		return
	} else if sess.Get("auth") == LoggedIn {
		f.Info("You are already logged in!")
		ctx.Redirect("/")
		return
	}
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "login")
}

// PostLoginHandler post response for login page.
func PostLoginHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == Verification {
		f.Warning("You need to verify before you continue.")
		ctx.Redirect("/verify")
		return
	} else if sess.Get("auth") == LoggedIn {
		f.Info("You are already logged in!")
		ctx.Redirect("/")
		return
	}
	// Generate code
	code := fmt.Sprint(rand.Intn(8999) + 1000)
	to := fmt.Sprintf("%s%s", strings.ToLower(ctx.QueryTrim("email")), settings.Config.UniEmailDomain)
	err := checkmail.ValidateFormat(to)
	if err != nil {
		f.Error("You provided an invalid email.")
		ctx.Redirect("/login")
		return
	}

	if !settings.Config.DevMode {
		go mailer.EmailCode(to, code)
	}
	sess.Set("auth", Verification)
	sess.Set("code", code)
	sess.Set("user", to)
	ctx.Redirect("/verify")
}

// VerifyHandler response for verification page.
func VerifyHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == LoggedOut {
		ctx.Redirect("/login")
		return
	} else if sess.Get("auth") == LoggedIn || sess.Get("auth") != Verification {
		f.Info("You are already logged in!")
		ctx.Redirect("/")
		return
	}
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["email"] = sess.Get("user")
	ctx.HTML(200, "validate_login")
}

// CancelHandler post response for canceling verification.
func CancelHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != Verification {
		ctx.Redirect("/login")
		return
	}

	sess.Set("auth", LoggedOut)
	ctx.Redirect("/login")
}

// PostVerifyHandler post reponse for verification.
func PostVerifyHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == LoggedOut {
		ctx.Redirect("/login")
		return
	} else if sess.Get("auth") == LoggedIn {
		ctx.Redirect("/")
		return
	}
	if ctx.QueryTrim("code") != sess.Get("code") && !settings.Config.DevMode {
		f.Error("The code you entered is invalid, make sure you use the latest code sent to you.")
		ctx.Redirect("/verify")
		return
	}
	sess.Set("auth", LoggedIn)
	if !models.HasUser(sess.Get("user").(string)) {
		models.AddUser(&models.User{
			Username: sess.Get("user").(string),
			IsAdmin:  false,
		})
	}
	ctx.Redirect("/")
}
