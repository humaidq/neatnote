package routes

import (
	"bytes"
	"fmt"
	"git.sr.ht/~humaid/notes-overflow/models"
	"git.sr.ht/~humaid/notes-overflow/modules/mailer"
	"git.sr.ht/~humaid/notes-overflow/modules/settings"
	"github.com/badoux/checkmail"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	macaron "gopkg.in/macaron.v1"
	"html/template"
	"math/rand"
	"strconv"
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

func ProfileHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "profile")
}

func PostProfileHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		ctx.Redirect("/login")
		return // TODO some error handling
	}

	// TODO validate fullname
	err := models.UpdateUser(&models.User{
		Username: sess.Get("user").(string),
		FullName: ctx.Query("fullname"),
	})
	if err != nil {
		panic(err)
	}

	ctx.Redirect("/profile")
}

func LogoutHandler(ctx *macaron.Context, sess session.Store) {
	sess.Set("auth", LoggedOut)
	//sess.Flush()
	ctx.Redirect("/")
}

func AdminAddCourseHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		return // TODO some error handling
	}

	user, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		return
	}

	if !user.IsAdmin {
		ctx.Redirect("/")
		return
	}

	ctx.Data["csrf_token"] = x.GetToken()

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
		ctx.Redirect("/")
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

func PostLoginHandler(ctx *macaron.Context, sess session.Store) {
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

	if !settings.DevMode {
		go mailer.EmailCode(to, code)
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
	} else if sess.Get("auth") == LoggedIn || sess.Get("auth") != Verification {
		ctx.Redirect("/")
		return
	}
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["email"] = sess.Get("user")
	if ctx.Query("err") == "1" {
		ctx.Data["invalid"] = 1
	}
	ctx.HTML(200, "validate_login")
}

func CancelHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != Verification {
		ctx.Redirect("/login")
		return
	}

	sess.Set("auth", LoggedOut)
	ctx.Redirect("/login")
}

func PostVerifyHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") == LoggedOut {
		ctx.Redirect("/login")
		return
	} else if sess.Get("auth") == LoggedIn {
		ctx.Redirect("/")
		return
	}
	if ctx.QueryTrim("code") != sess.Get("code") && !settings.DevMode {
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

func CourseHandler(ctx *macaron.Context, sess session.Store) {
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
	post.LoadComments()
	ctx.Data["Post"] = post

	ctx.Data["PosterID"] = strings.Split(post.PosterID, "@")[0]

	ctx.Data["FormattedPost"] = template.HTML(markdownToHTML(post.Text))

	for i := range post.Comments {
		post.Comments[i].LoadCreated()
		post.Comments[i].LoadPoster()
		post.Comments[i].FormattedText =
			template.HTML(markdownToHTML(post.Comments[i].Text))
	}

	ctx.Data["csrf_token"] = x.GetToken()

	ctx.HTML(200, "post")
}

func markdownToHTML(s string) string {
	//	ext := blackfriday.Safelink & blackfriday.NoreferrerLinks
	//		& blackfriday.NoopenerLinks & blackfriday.HrefTargetBlank & blackfriday.SmartyPants
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(s), &buf); err != nil {
		panic(err)
	}
	return string(bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes()))
}

func PostCommentPostHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		ctx.Redirect("/login")
		return
	}

	postID, _ := strconv.ParseInt(ctx.Params("post"), 10, 64)

	// TODO check if post/course exists
	com := &models.Comment{
		PosterID: sess.Get("user").(string),
		PostID:   postID,
		Text:     ctx.Query("text"),
	}

	err := models.AddComment(com)

	if err != nil {
		panic(err)
	}

	ctx.Redirect(fmt.Sprintf("/course/%s/%s#c-%d", ctx.Params("course"),
		ctx.Params("post"), com.CommentID))
}

func CreatePostHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		ctx.Redirect("/login")
		return
	}
	// check if course exists TODO

	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "create-post")
}

func PostCreatePostHandler(ctx *macaron.Context, sess session.Store) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		ctx.Redirect("/login")
		return
	}
	// check if course exists TODO

	// TODO error handling
	post := &models.Post{
		CourseCode: ctx.Params("course"),
		PosterID:   sess.Get("user").(string),
		Locked:     false,
		Title:      ctx.Query("title"),
		Text:       ctx.Query("text"),
	}
	err := models.AddPost(post)
	if err != nil {
		panic(err)
	}

	ctx.Redirect(fmt.Sprintf("/course/%s/%d", ctx.Params("course"), post.PostID))
}
