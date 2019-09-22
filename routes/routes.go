package routes

import (
	"fmt"
	"git.sr.ht/~humaid/notes-overflow/models"
	"git.sr.ht/~humaid/notes-overflow/modules/mailer"
	"github.com/badoux/checkmail"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	macaron "gopkg.in/macaron.v1"
	"html/template"
	"math/rand"
	"strings"
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
	ctx.Data["Courses"] = models.GetCourses()
	ctx.HTML(200, "index")
}

func LogoutHandler(ctx *macaron.Context, sess session.Store) {
	sess.Set("auth", LoggedOut)
	//sess.Flush()
	ctx.Redirect("/")
}

func AdminAddCourseHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		return // TODO some error handling
	}

	user, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		return
	}

	if !user.IsAdmin {
		return
	}

	ctx.HTML(200, "admin/add-course")
}

func AdminPostAddCourseHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		return // TODO some error handling
	}

	user, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		return
	}

	if !user.IsAdmin {
		return
	}

	courseCode := ctx.Query("coursecode")
	courseName := ctx.Query("coursename")

	// Check if course exists already
	if _, err1 := models.GetCourse(courseCode); err1 == nil {
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

func CourseHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)

	course, err := models.GetCourse(ctx.Params("course"))
	if err != nil {
		ctx.Redirect("/")
		return // TODO proper error
	}

	course.LoadPosts()

	ctx.Data["Course"] = course

	ctx.HTML(200, "course")
}

func PostPageHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)

	course, err := models.GetCourse(ctx.Params("course"))
	if err != nil {
		ctx.Redirect("/")
		return // TODO proper error
	}

	ctx.Data["Course"] = course

	var post *models.Post
	post, err = models.GetPost(ctx.Params("post"))
	ctx.Data["Post"] = post

	ctx.Data["Poster"] = strings.Split(post.Poster, "@")[0]

	unsafe := blackfriday.Run([]byte(post.Text))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	ctx.Data["FormattedPost"] = template.HTML(html)

	ctx.HTML(200, "post")
}

func PostCommentPostHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {

}
func CreatePostHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		ctx.Redirect("/login")
		return
	}
	// check if course exists TODO

	ctx.HTML(200, "create-post")
}
func PostCreatePostHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		ctx.Redirect("/login")
		return
	}
	// check if course exists TODO

	// TODO error handling
	err := models.AddPost(&models.Post{
		CourseCode: ctx.Params("course"),
		Poster:     sess.Get("user").(string),
		Locked:     false,
		Title:      ctx.Query("title"),
		Text:       ctx.Query("text"),
	})
	if err != nil {
		panic(err)
	}

	ctx.Redirect(fmt.Sprintf("/course/%s", ctx.Params("course")))
}
