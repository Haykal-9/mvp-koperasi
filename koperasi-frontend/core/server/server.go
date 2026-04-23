package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// Assets is assigned from outside (api/index.go or cmd/web/main.go).
// For Vercel: assigned an embed.FS. For local dev: assigned an os.DirFS.
var Assets fs.FS

// pageTemplates holds {layoutName: {pagePath: parsedTemplate}}
var pageTemplates = map[string]map[string]*template.Template{}

var templateFuncs = template.FuncMap{
	"rupiah": func(n float64) string {
		prefix := "Rp "
		if n < 0 {
			prefix = "Rp -"
			n = -n
		}
		s := fmt.Sprintf("%.0f", n)
		out := []byte{}
		cnt := 0
		for i := len(s) - 1; i >= 0; i-- {
			out = append([]byte{s[i]}, out...)
			cnt++
			if cnt%3 == 0 && i != 0 {
				out = append([]byte{'.'}, out...)
			}
		}
		return prefix + string(out)
	},
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	"statusClass": func(s string) string {
		return "badge-status-" + strings.ToLower(s)
	},
	"sumSimpanan": func(m model.Member) float64 {
		return m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela
	},
	"add": func(a, b float64) float64 { return a + b },
	"sub": func(a, b float64) float64 { return a - b },
	"mul": func(a, b float64) float64 { return a * b },
	"div": func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		return a / b
	},
}

// readTemplate reads a template file from the embedded FS.
func readTemplate(p string) (string, error) {
	b, err := fs.ReadFile(Assets, p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// loadTemplates parses each page template together with its layout from the embedded FS.
func loadTemplates() {
	pageTemplates["base"] = map[string]*template.Template{}
	pageTemplates["auth"] = map[string]*template.Template{}

	baseLayout := "templates/layouts/base.html"
	authLayout := "templates/layouts/auth.html"

	register := func(layout, layoutFile, pagePath string) {
		name := pageNameFromPath(pagePath)

		layoutContent, err := readTemplate(layoutFile)
		if err != nil {
			log.Printf("failed to read layout %s: %v", layoutFile, err)
			return
		}
		pageContent, err := readTemplate(pagePath)
		if err != nil {
			log.Printf("failed to read page %s: %v", pagePath, err)
			return
		}

		tmpl := template.New(path.Base(layoutFile)).Funcs(templateFuncs)
		tmpl = template.Must(tmpl.Parse(layoutContent))
		tmpl = template.Must(tmpl.New(path.Base(pagePath)).Parse(pageContent))

		pageTemplates[layout][name] = tmpl
	}

	// Auth pages
	authEntries, _ := fs.Glob(Assets, "templates/auth/*.html")
	for _, p := range authEntries {
		register("auth", authLayout, p)
	}

	// Module pages
	moduleDirs := []string{"dashboard", "member", "product", "order", "loan", "finance"}
	for _, dir := range moduleDirs {
		entries, _ := fs.Glob(Assets, "templates/"+dir+"/*.html")
		for _, p := range entries {
			register("base", baseLayout, p)
		}
	}
}

func pageNameFromPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "templates/")
	p = strings.TrimSuffix(p, ".html")
	return p
}

// renderPage looks up the template, injects session info + flash messages, then executes.
func renderPage(c *gin.Context, layout, page string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}

	sess := sessions.Default(c)
	if _, ok := data["UserNama"]; !ok {
		if v := sess.Get("user_nama"); v != nil {
			data["UserNama"] = v
		}
		if v := sess.Get("user_role"); v != nil {
			data["UserRole"] = v
		}
	}

	if _, ok := data["FlashError"]; !ok {
		if msg, has := handler.PopFlash(c, "error"); has {
			data["FlashError"] = msg
		}
	}
	if _, ok := data["FlashSuccess"]; !ok {
		if msg, has := handler.PopFlash(c, "success"); has {
			data["FlashSuccess"] = msg
		}
	}

	if _, ok := data["CartCount"]; !ok {
		data["CartCount"] = handler.CartCount(c)
	}

	tmpl, ok := pageTemplates[layout][page]
	if !ok {
		c.String(http.StatusInternalServerError, "template not found: %s/%s", layout, page)
		return
	}
	entry := "base"
	if layout == "auth" {
		entry = "auth_base"
	}
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, entry, data); err != nil {
		log.Printf("render error: %v", err)
	}
}

func SetupApp() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	loadTemplates()

	store := cookie.NewStore([]byte("koperasi-secret-key-change-me"))
	r.Use(sessions.Sessions("koperasi_session", store))

	// Serve static files from embedded FS
	staticFS, err := fs.Sub(Assets, "static")
	if err != nil {
		log.Fatalf("failed to sub static FS: %v", err)
	}
	r.StaticFS("/static", http.FS(staticFS))

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard")
	})

	// ===== AUTH =====
	authH := handler.NewAuthHandler(renderPage)
	r.GET("/login", authH.ShowLogin)
	r.POST("/login", authH.DoLogin)
	r.GET("/register", authH.ShowRegister)
	r.POST("/register", authH.DoRegister)
	r.GET("/logout", authH.DoLogout)

	// ===== PROTECTED =====
	dashH := handler.NewDashboardHandler(renderPage)
	memH := handler.NewMemberHandler(renderPage)
	prodH := handler.NewProductHandler(renderPage)
	orderH := handler.NewOrderHandler(renderPage)
	loanH := handler.NewLoanHandler(renderPage)
	finH := handler.NewFinanceHandler(renderPage)

	protected := r.Group("/")
	protected.Use(handler.AuthRequired())
	{
		protected.GET("/dashboard", dashH.Index)

		protected.GET("/members", memH.List)
		protected.GET("/members/register", memH.ShowRegister)
		protected.POST("/members/register", memH.DoRegister)
		protected.GET("/members/:id", memH.Detail)
		protected.POST("/members/:id/approve", memH.Approve)
		protected.POST("/members/:id/reject", memH.Reject)
		protected.GET("/members/:id/simpanan", memH.ShowSimpanan)
		protected.POST("/members/:id/simpanan", memH.DoSimpanan)
		protected.GET("/members/:id/resign", memH.ShowResign)
		protected.POST("/members/:id/resign", memH.DoResign)

		protected.GET("/products", prodH.Catalog)
		protected.GET("/products/create", prodH.ShowCreate)
		protected.POST("/products/create", prodH.DoCreate)
		protected.GET("/products/review", prodH.Review)
		protected.GET("/products/:id", prodH.Detail)
		protected.POST("/products/:id/approve", prodH.Approve)
		protected.POST("/products/:id/reject", prodH.Reject)
		protected.GET("/products/:id/stock", prodH.ShowStock)
		protected.POST("/products/:id/stock", prodH.DoStock)

		protected.GET("/cart", orderH.ShowCart)
		protected.POST("/cart/add", orderH.AddToCart)
		protected.POST("/cart/update", orderH.UpdateCart)
		protected.POST("/cart/remove", orderH.RemoveFromCart)
		protected.GET("/checkout", orderH.ShowCheckout)
		protected.POST("/checkout", orderH.DoCheckout)
		protected.GET("/orders", orderH.List)
		protected.GET("/orders/complaints", orderH.Complaints)
		protected.GET("/orders/:id", orderH.Detail)
		protected.POST("/orders/:id/ship", orderH.Ship)
		protected.POST("/orders/:id/complete", orderH.Complete)
		protected.GET("/orders/:id/complain", orderH.ShowComplain)
		protected.POST("/orders/:id/complain", orderH.DoComplain)
		protected.POST("/orders/:id/resolve", orderH.Resolve)

		protected.GET("/loans", loanH.List)
		protected.GET("/loans/apply", loanH.ShowApply)
		protected.POST("/loans/apply", loanH.DoApply)
		protected.GET("/loans/:id", loanH.Detail)
		protected.POST("/loans/:id/approve", loanH.Approve)
		protected.POST("/loans/:id/reject", loanH.Reject)
		protected.POST("/loans/:id/disburse", loanH.Disburse)
		protected.POST("/loans/:id/pay", loanH.Pay)

		protected.GET("/finance/journals", finH.Journals)
		protected.GET("/finance/journals/create", finH.ShowCreateJournal)
		protected.POST("/finance/journals/create", finH.DoCreateJournal)
		protected.GET("/finance/summary", finH.Summary)
	}

	return r
}
