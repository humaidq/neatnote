package routes

import (
	"fmt"
	"git.sr.ht/~humaid/notes-overflow/models"
	"git.sr.ht/~humaid/notes-overflow/modules/mailer"
	"github.com/badoux/checkmail"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
	"math/rand"
)

const (
	LoggedOut = iota
	Verification
	LoggedIn
)

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
			// TODO problem here...
			fmt.Println("Cannot load auth'd user! ", err)
		}

	}
	ctx.Data["SiteTitle"] = "Notes Overflow"
}

func HomepageHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	ctx.HTML(200, "index")
}

func LogoutHandler(ctx *macaron.Context, sess session.Store) {
	sess.Set("auth", LoggedOut)
	//sess.Flush()
	ctx.Redirect("/")
}

func LoginHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == Verification {
		ctx.Redirect("/verify")
		return
	} else if sess.Get("auth") == LoggedIn {
		ctx.Redirect("/")
		return
	}
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "login")
}

func PostLoginHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == Verification {
		ctx.Redirect("/verify")
		return
	} else if sess.Get("auth") == LoggedIn {
		ctx.Redirect("/")
		return
	}
	// Generate code
	code := fmt.Sprint(rand.Intn(8999) + 1000)
	to := fmt.Sprintf("%s@hw.ac.uk", ctx.Query("email"))
	err := checkmail.ValidateFormat(to)
	if err != nil {
		ctx.PlainText(200, []byte("Invalid email")) // TODO replace all plaintext with proper response
		return
	}

	err = mailer.EmailCode(to, code)
	if err != nil {
		ctx.PlainText(200, []byte("Failed to email, go back and check email."))
		return
	}
	sess.Set("auth", Verification)
	sess.Set("code", code)
	sess.Set("user", to)
	ctx.Redirect("/verify")
}

func VerifyHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == LoggedOut {
		ctx.Redirect("/login")
		return
	} else if sess.Get("auth") == LoggedIn {
		ctx.Redirect("/")
		return
	}
	ctx.Data["email"] = sess.Get("user")
	ctx.HTML(200, "validate_login")
}

func PostVerifyHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == LoggedOut {
		ctx.Redirect("/login")
		return
	} else if sess.Get("auth") == LoggedIn {
		ctx.Redirect("/")
		return
	}
	if ctx.Query("code") != sess.Get("code") {
		ctx.Redirect("/verify?err=1") // TODO proper error
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
