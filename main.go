package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/martini-contrib/strict"
	"html"
	"html/template"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	helpers := template.FuncMap{
		// Unescape unescapes and parses HTML from database objects.
		// Used in templates such as "/post/display.tmpl"
		"unescape": func(s string) template.HTML {
			return template.HTML(html.UnescapeString(s))
		},
		"time": func(t int64) string {
			return strings.Split(time.Unix(t, 0).String(), " ")[0]
		},
		"title": func(t interface{}) string {
			post, exists := t.(Post)
			if exists {
				return post.Title
			}
			return "Juuso Haavisto"
		},
	}

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte(os.Getenv("vertigo_hash")))
	m.Use(sessions.Sessions("user", store))
	m.Use(middleware())
	m.Use(strict.Strict)
	m.Use(gzip.All())
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
		Funcs:  []template.FuncMap{helpers}, // Specify helper function maps for templates to access.
	}))

	m.Get("/", Homepage)

	m.Group("/feeds", func(r martini.Router) {
		r.Get("", func(res render.Render) {
			res.Redirect("/feeds/rss", 302)
		})
		r.Get("/atom", ReadFeed)
		r.Get("/rss", ReadFeed)
	})

	m.Group("/post", func(r martini.Router) {

		// Please note that `/new` route has to be before the `/:title` route. Otherwise the program will try
		// to fetch for Post named "new".
		// For now I'll keep it this way to streamline route naming.
		r.Get("/new", ProtectedPage, func(res render.Render) {
			res.HTML(200, "post/new", nil)
		})
		r.Get("/:title", ReadPost)
		r.Get("/:title/edit", EditPost)
		r.Post("/:title/edit", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Post{}), binding.ErrorHandler, UpdatePost)
		r.Get("/:title/delete", DeletePost)
		r.Get("/:title/publish", PublishPost)
		r.Post("/new", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Post{}), binding.ErrorHandler, CreatePost)
		r.Post("/search", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Search{}), binding.ErrorHandler, SearchPost)

	})

	m.Group("/user", func(r martini.Router) {

		r.Get("", ProtectedPage, ReadUser)
		//r.Post("/delete", strict.ContentType("application/x-www-form-urlencoded"), ProtectedPage, binding.Form(Person{}), DeleteUser)

		//r.Get("/register", SessionRedirect, func(res render.Render) {
		//	res.HTML(200, "user/register", nil)
		//})
		//r.Post("/register", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Person{}), binding.ErrorHandler, CreateUser)

		r.Get("/login", SessionRedirect, func(res render.Render) {
			res.HTML(200, "user/login", nil)
		})
		r.Post("/login", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Person{}), LoginUser)
		r.Get("/logout", LogoutUser)

	})

	m.Group("/api", func(r martini.Router) {

		r.Get("", func(res render.Render) {
			res.HTML(200, "api/index", nil)
		})
		r.Get("/users", ReadUsers)
		r.Get("/user/:id", ReadUser)
		//r.Delete("/user", DeleteUser)
		//r.Post("/user", strict.ContentType("application/json"), binding.Json(Person{}), binding.ErrorHandler, CreateUser)
		//r.Post("/user/login", strict.ContentType("application/json"), binding.Json(Person{}), binding.ErrorHandler, LoginUser)
		//r.Get("/user/logout", LogoutUser)

		r.Get("/posts", ReadPosts)
		r.Get("/post/:title", ReadPost)
		r.Post("/post", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, CreatePost)
		r.Get("/post/:title/publish")
		r.Post("/post/:title/edit", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, UpdatePost)
		r.Get("/post/:title/delete", DeletePost)
		r.Post("/post", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, CreatePost)
		r.Post("/post/search/:query", strict.ContentType("application/json"), binding.Json(Search{}), binding.ErrorHandler, SearchPost)

	})

	m.Router.NotFound(strict.MethodNotAllowed, strict.NotFound)
	m.Run()

	log.Println("Vertigo started")
}
