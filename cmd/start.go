package cmd

import (
	"fmt"
	"git.sr.ht/~humaid/notes-overflow/models"
	"git.sr.ht/~humaid/notes-overflow/modules/settings"
	"git.sr.ht/~humaid/notes-overflow/routes"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/captcha"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/mysql"
	"github.com/urfave/cli"
	macaron "gopkg.in/macaron.v1"
	"log"
	"net/http"
)

// CmdStart represents a command-line command
// which starts the bot.
var CmdStart = cli.Command{
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
	m.Use(macaron.Renderer())
	m.Use(cache.Cacher())
	sqlConfig := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		settings.DBConfig.User, settings.DBConfig.Password, settings.DBConfig.Host, settings.DBConfig.Name)
	fmt.Println(sqlConfig)
	m.Use(session.Sessioner(session.Options{
		Provider:       "mysql",
		ProviderConfig: sqlConfig,
	}))
	m.Use(csrf.Csrfer())
	m.Use(captcha.Captchaer())

	// Web routes
	m.Get("/", routes.HomepageHandler)
	m.Get("/profile", routes.ProfileHandler)
	m.Post("/profile", csrf.Validate, routes.PostProfileHandler)
	m.Get("/qna", routes.QnAHandler)
	m.Get("/guidelines", routes.GuidelinesHandler)

	// Login and verification
	m.Get("/login", routes.LoginHandler)
	m.Post("/login", csrf.Validate, routes.PostLoginHandler)
	m.Get("/logout", routes.LogoutHandler)
	m.Get("/verify", routes.VerifyHandler)
	m.Post("/verify", csrf.Validate, routes.PostVerifyHandler)
	m.Get("/cancel", routes.CancelHandler)

	m.Group("/admin", func() {
		m.Get("/add_course", routes.AdminAddCourseHandler)
		m.Post("/add_course", csrf.Validate, routes.AdminPostAddCourseHandler)
	})

	m.Group("/course/:course", func() {
		m.Get("/", routes.CourseHandler)
		m.Get("/post", routes.CreatePostHandler)
		m.Post("/post", csrf.Validate, routes.PostCreatePostHandler)
		m.Group("/:post", func() {
			m.Get("/", routes.PostPageHandler)
			m.Post("/", csrf.Validate, routes.PostCommentPostHandler)
		})
	})

	log.Printf("Starting web server on port %s\n", settings.SitePort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", settings.SitePort), m))
	return nil
}
