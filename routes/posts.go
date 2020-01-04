package routes

import (
	"bytes"
	"fmt"
	"git.sr.ht/~humaid/neatnote/models"
	"git.sr.ht/~humaid/neatnote/modules/common"
	"git.sr.ht/~humaid/neatnote/modules/namegen"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	macaron "gopkg.in/macaron.v1"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

var (
	// simpleTextExp matches a simple text string.
	simpleTextExp = regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)
	// htmlTagExp roughly matches any HTML tag.
	htmlTagExp = regexp.MustCompile(`\<?(\/)?[a-zA-Z0-9 "=\n:\/\.\@\#\&\;\+\-\?\,\_]+\>`)
)

// CourseHandler response for a course page.
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

// PostPageHandler response for a post page.
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
	if err != nil {
		panic(err)
	}
	post.LoadComments()
	ctx.Data["Post"] = post

	ctx.Data["PosterID"] = strings.Split(post.PosterID, "@")[0]

	ctx.Data["FormattedPost"] = template.HTML(markdownToHTML(post.Text))
	if post.Locked || course.Locked {
		ctx.Data["Locked"] = 1
	}

	if post.Anonymous {
		ctx.Data["Poster"] = post.AnonName
	} else {
		if post.Poster.FullName == "" {
			ctx.Data["Poster"] = strings.Split(post.PosterID, "@")[0]
		} else {
			ctx.Data["Poster"] = post.Poster.FullName
		}
	}

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

	u, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		panic(err)
	}
	if common.ContainsInt64(u.Upvoted, post.PostID) {
		ctx.Data["Upvoted"] = 1
	}

	ctx.Data["csrf_token"] = x.GetToken()

	ctx.HTML(200, "post")
}

// markdownToHTML converts a string (in Markdown) and outputs (X)HTML.
// The input may also contain HTML, and the output is sanitized.
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

// PostCommentPostHandler post response for posting a comment to a post.
func PostCommentPostHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before you comment!")
		ctx.Redirect("/login")
		return
	}

	postID, _ := strconv.ParseInt(ctx.Params("post"), 10, 64)

	if c, err := models.GetCourse(ctx.Params("course")); err != nil {
		f.Error("Welp! The course doesn't exist.")
		ctx.Redirect("/")
		return
	} else if c.Locked {
		f.Error("This course is locked and you cannot comment on a post.")
		ctx.Redirect(fmt.Sprintf("/course/%s/%s", ctx.Params("course"),
			ctx.Params("post")))
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

	if getMarkdownLength(ctx.QueryTrim("text")) < 2 {
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

// CreatePostHandler response for creating a new post.
func CreatePostHandler(ctx *macaron.Context, x csrf.CSRF, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before you create a post!")
		ctx.Redirect("/login")
		return
	}
	if c, err := models.GetCourse(ctx.Params("course")); err != nil {
		f.Error("Welp! The course no longer exist.")
		ctx.Redirect("/")
		return
	} else if c.Locked {
		f.Error("This course is locked and you cannot create a post.")
		ctx.Redirect(fmt.Sprintf("/course/%s", ctx.Params("course")))
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

// PostCreatePostHandler post response for creating a new post.
func PostCreatePostHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before you create a post!")
		ctx.Redirect("/login")
		return
	}

	if c, err := models.GetCourse(ctx.Params("course")); err != nil {
		f.Error("Welp! The course doesn't exist.")
		ctx.Redirect("/")
		return
	} else if c.Locked {
		f.Error("This course is locked and you cannot create a post.")
		ctx.Redirect(fmt.Sprintf("/course/%s", ctx.Params("course")))
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
		Anonymous:  false,
	}

	if ctx.Query("anon") == "on" {
		post.Anonymous = true
		post.AnonName = namegen.GetName()
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

// UpvotePostHandler post response for posting a comment to a post.
func UpvotePostHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctxInit(ctx, sess)
	if sess.Get("auth") != LoggedIn {
		f.Error("Please login before you comment!")
		ctx.Redirect("/login")
		return
	}

	postID, _ := strconv.ParseInt(ctx.Params("post"), 10, 64)

	if c, err := models.GetCourse(ctx.Params("course")); err != nil {
		f.Error("Welp! The course doesn't exist.")
		ctx.Redirect("/")
		return
	} else if c.Locked {
		f.Error("This course is locked and you cannot comment on a post.")
		ctx.Redirect(fmt.Sprintf("/course/%s/%s", ctx.Params("course"),
			ctx.Params("post")))
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

	u, err := models.GetUser(sess.Get("user").(string))
	if err != nil {
		panic(err)
	}

	if common.ContainsInt64(u.Upvoted, postID) {
		err = models.UnvotePost(sess.Get("user").(string), postID)
		if err != nil {
			f.Error(fmt.Sprintf("%s.", err))
			ctx.Redirect(fmt.Sprintf("/course/%s/%s", ctx.Params("course"),
				ctx.Params("post")))
			return
		}
		f.Info("You have unvoted the post.")
	} else {
		err = models.UpvotePost(sess.Get("user").(string), postID)
		if err != nil {
			f.Error(fmt.Sprintf("%s.", err))
			ctx.Redirect(fmt.Sprintf("/course/%s/%s", ctx.Params("course"),
				ctx.Params("post")))
			return
		}
		f.Info("Post upvoted.")
	}

	ctx.Redirect(fmt.Sprintf("/course/%s/%s", ctx.Params("course"),
		ctx.Params("post")))
}
