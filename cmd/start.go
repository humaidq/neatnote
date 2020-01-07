package cmd

import (
	"fmt"
	"git.sr.ht/~humaid/neatnote/models"
	"git.sr.ht/~humaid/neatnote/modules/settings"
	"git.sr.ht/~humaid/neatnote/routes"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/captcha"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/mysql" // MySQL driver for persistent sessions
	"github.com/urfave/cli/v2"
	macaron "gopkg.in/macaron.v1"
	"html/template"
	"log"
	"net/http"
	"time"
)

// CmdStart represents a command-line command
// which starts the bot.
var CmdStart = &cli.Command{
	Name:    "run",
	Aliases: []string{"start", "web"},
	Usage:   "Start the web server",
	Action:  start,
}

func start(clx *cli.Context) (err error) {
	settings.LoadConfig()
	engine := models.SetupEngine()
	defer engine.Close()

	// Run macaron
	m := macaron.Classic()
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Funcs: []template.FuncMap{map[string]interface{}{
			"CalcTime": func(sTime time.Time) string {
				return fmt.Sprint(time.Since(sTime).Nanoseconds() / int64(time.Millisecond))
			},
		}},
		IndentJSON: true,
	}))

	m.Use(cache.Cacher())
	sessOpt := session.Options{
		CookieLifeTime: 15778800, // 6 months
		Gclifetime:     15778800,
		CookieName:     "hithereimacookie",
	}
	if settings.Config.DBConfig.Type == settings.MySQL {
		sqlConfig := fmt.Sprintf("%s:%s@tcp(%s)/%s",
			settings.Config.DBConfig.User, settings.Config.DBConfig.Password,
			settings.Config.DBConfig.Host, settings.Config.DBConfig.Name)
		sessOpt.Provider = "mysql"
		sessOpt.ProviderConfig = sqlConfig
		sessOpt.CookieLifeTime = 0
	}
	m.Use(session.Sessioner(sessOpt))
	m.Use(csrf.Csrfer())
	m.Use(captcha.Captchaer())

	// Web routes
	m.Get("/", routes.HomepageHandler)
	m.Get("/profile", routes.ProfileHandler)
	m.Post("/profile", csrf.Validate, routes.PostProfileHandler)
	m.Post("/profile/data.json", csrf.Validate, routes.PostDataHandler)
	m.Get("/qna", routes.QnAHandler)
	m.Get("/guidelines", routes.GuidelinesHandler)

	// Login and verification
	m.Get("/login", routes.LoginHandler)
	m.Post("/login", csrf.Validate, routes.PostLoginHandler)
	m.Get("/logout", routes.LogoutHandler)
	m.Get("/verify", routes.VerifyHandler)
	m.Post("/verify", csrf.Validate, routes.PostVerifyHandler)
	m.Post("/cancel", csrf.Validate, routes.CancelHandler)

	m.Group("/admin", func() {
		m.Get("/add_course", routes.AdminAddCourseHandler)
		m.Post("/add_course", csrf.Validate, routes.AdminPostAddCourseHandler)
	})

	m.Group("/course/:course", func() {
		m.Get("/", routes.CourseHandler)
		m.Get("/:sort(new|top)", routes.CourseHandler)
		m.Get("/post", routes.CreatePostHandler)
		m.Post("/post", csrf.Validate, routes.PostCreatePostHandler)
		m.Group("/:post", func() {
			m.Get("/", routes.PostPageHandler)
			m.Post("/", csrf.Validate, routes.PostCommentPostHandler)
			m.Get("/upvote", routes.UpvotePostHandler)
			m.Get("/edit", routes.EditPostHandler)
			m.Post("/edit", routes.PostEditPostHandler)
			m.Get("/del/:id", routes.DeleteCommentHandler)
			m.Get("/del", routes.DeletePostHandler)
			m.Get("/reveal", routes.RevealPosterHandler)
		})
	})

	log.Printf("Starting web server on port %s\n", settings.Config.SitePort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", settings.Config.SitePort), m))
	return nil
}
