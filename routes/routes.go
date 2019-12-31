package routes

import (
	"bytes"
	"fmt"
	"git.sr.ht/~humaid/neatnote/models"
	"git.sr.ht/~humaid/neatnote/modules/mailer"
	"git.sr.ht/~humaid/neatnote/modules/settings"
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
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	LoggedOut = iota
	Verification
	LoggedIn
)

var (
	simpleTextExp = regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)
	// htmlTagExp roughly matches any HTML tag.
	htmlTagExp = regexp.MustCompile(`\<?(\/)?[a-zA-Z0-9 "=\n:\/\.\@\#\&\;\+\-\?\,\_]+\>`)
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
			fmt.Println("Cannot load auth'd user! ", err)
			// Let's log out the user
			ctx.Data["LoggedIn"] = 0
			sess.Set("auth", LoggedOut)
		}
	}
	ctx.Data["UniEmailDomain"] = settings.Config.UniEmailDomain
	if settings.Config.DevMode {
		ctx.Data["DevMode"] = 1
	}
	ctx.Data["SiteTitle"] = "Neat Note"
}

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

func ProfileHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before editing your profile.")
		ctx.Redirect("/login", http.StatusUnauthorized)
		return
	}
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "profile")
}

func PostProfileHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before editing your profile.")
		ctx.Redirect("/login", http.StatusUnauthorized)
		return
	}
	fname := ctx.QueryTrim("fullname")

	if !simpleTextExp.Match([]byte(fname)) || len(fname) > 32 || len(fname) < 1 {
		f.Error("Your display name must only contain alphabet, numbers, and spaces. And cannot be over 32 characters.")
		ctx.Redirect("/profile")
		return
	}
	fmt.Println(fname)
	err := models.UpdateUser(&models.User{
		Username: sess.Get("user").(string),
		FullName: fname,
	})
	if err != nil {
		panic(err)
	}

	ctx.Redirect("/profile")
}

func PostDataHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before editing your profile.")
		ctx.Redirect("/login", http.StatusUnauthorized)
		return
	}

	u, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		panic(err)
	}
	var p []models.Post
	p, err = models.GetAllUserPosts(sess.Get("user").(string))

	ctx.JSON(200, map[string]interface{}{
		"user":  u,
		"posts": p,
	})
}

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

func CourseHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)

	course, err := models.GetCourse(ctx.Params("course"))
	if err != nil {
		f.Error("Welp! The course doesn't exist.")
		ctx.Redirect("/")
		return
	}

	course.LoadPosts()
	ctx.Data["Course"] = course
	ctx.HTML(200, "course")
}

func PostPageHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)

	course, err := models.GetCourse(ctx.Params("course"))
	if err != nil {
		f.Error("Welp! The course no longer exists.")
		ctx.Redirect("/")
		return
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
	if sess.Get("c.text") != nil {
		ctx.Data["ctext"] = sess.Get("c.text").(string)
		sess.Delete("c.text")
	}

	ctx.Data["csrf_token"] = x.GetToken()

	ctx.HTML(200, "post")
}

func markdownToHTML(s string) string {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(s), &buf); err != nil {
		panic(err)
	}
	return string(bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes()))
}

func PostCommentPostHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before you comment!")
		ctx.Redirect("/login")
		return
	}

	postID, _ := strconv.ParseInt(ctx.Params("post"), 10, 64)

	if _, err := models.GetCourse(ctx.Params("course")); err != nil {
		f.Error("Welp! The course doesn't exist.")
		ctx.Redirect("/")
		return
	}

	post, err := models.GetPost(ctx.Params("post"))
	if err != nil {
		f.Error("Welp! The post no longer exists!")
		ctx.Redirect(fmt.Sprintf("/course/%s", ctx.Params("course")))
		return
	}

	if post.Locked {
		f.Error("Comment cannot be posted as it has been locked.")
		ctx.Redirect(fmt.Sprintf("/course/%s/%s", ctx.Params("course"),
			ctx.Params("post")))
		return
	}

	if getMarkdownLength(ctx.QueryTrim("text")) < 8 {
		f.Error("The post is empty or too short!")
		// Pass over the text and title to errored page
		sess.Set("c.text", ctx.QueryTrim("text"))
		ctx.Redirect(fmt.Sprintf("/course/%s/%s", ctx.Params("course"), ctx.Params("post")))
		return
	}

	com := &models.Comment{
		PosterID: sess.Get("user").(string),
		PostID:   postID,
		Text:     ctx.QueryTrim("text"),
	}

	err = models.AddComment(com)

	if err != nil {
		panic(err)
	}

	ctx.Redirect(fmt.Sprintf("/course/%s/%s#c-%d", ctx.Params("course"),
		ctx.Params("post"), com.CommentID))
}

func CreatePostHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before you create a post!")
		ctx.Redirect("/login")
		return
	}
	if _, err := models.GetCourse(ctx.Params("course")); err != nil {
		f.Error("Welp! The course no longer exist.")
		ctx.Redirect("/")
		return
	}
	if sess.Get("p.title") != nil && len(sess.Get("p.title").(string)) > 0 {
		ctx.Data["ptitle"] = sess.Get("p.title").(string)
		ctx.Data["ptext"] = sess.Get("p.text").(string)
		sess.Delete("p.title")
		sess.Delete("p.text")
	}

	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "create-post")
}

func PostCreatePostHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before you create a post!")
		ctx.Redirect("/login")
		return
	}

	if _, err := models.GetCourse(ctx.Params("course")); err != nil {
		f.Error("Welp! The course doesn't exist.")
		ctx.Redirect("/")
		return
	}
	title := ctx.QueryTrim("title")
	text := ctx.Query("text")

	if !simpleTextExp.Match([]byte(title)) || len(title) > 32 || len(title) < 1 {
		f.Error("Invalid title")
		// Pass over the text and title to errored page
		sess.Set("p.title", title)
		sess.Set("p.text", text)
		ctx.Redirect(fmt.Sprintf("/course/%s/post", ctx.Params("course")))
		return
	} else if getMarkdownLength(text) < 8 {
		f.Error("The post is empty or too short!")
		// Pass over the text and title to errored page
		sess.Set("p.title", title)
		sess.Set("p.text", text)
		ctx.Redirect(fmt.Sprintf("/course/%s/post", ctx.Params("course")))
		return
	}

	post := &models.Post{
		CourseCode: ctx.Params("course"),
		PosterID:   sess.Get("user").(string),
		Locked:     false,
		Title:      ctx.QueryTrim("title"),
		Text:       ctx.Query("text"),
	}
	err := models.AddPost(post)
	if err != nil {
		panic(err)
	}

	ctx.Redirect(fmt.Sprintf("/course/%s/%d", ctx.Params("course"), post.PostID))
}

// getMarkdownLenngth renders the string provided, removes HTML tags and
// returns the length of the trimmed final string. This is useful to determine
// the post actual length.
func getMarkdownLength(s string) int {
	markdownHTMLWithoutTags := htmlTagExp.ReplaceAll([]byte(markdownToHTML(s)), []byte(""))
	return len(strings.TrimSpace(string(markdownHTMLWithoutTags)))
}
