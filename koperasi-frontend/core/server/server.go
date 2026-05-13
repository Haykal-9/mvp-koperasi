package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"
	"strconv"

	"koperasi-frontend/core/handler"
	ecommerce "koperasi-frontend/core/handler/ecommerce"
	"koperasi-frontend/core/middleware"
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
	"iadd": func(a, b int) int { return a + b },
	"contains": func(s, substr string) bool { return strings.Contains(s, substr) },
	"float64": func(i int) float64 { return float64(i) },
	"itoa": strconv.Itoa,
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
	pageTemplates["ec_base"] = map[string]*template.Template{}
	pageTemplates["ec_auth"] = map[string]*template.Template{}

	baseLayout := "templates/layouts/base.html"
	authLayout := "templates/layouts/auth.html"
	ecBaseLayout := "templates/ecommerce/layouts/ec_base.html"
	ecAuthLayout := "templates/ecommerce/layouts/ec_auth.html"

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

	// Auth pages (koperasi)
	authEntries, _ := fs.Glob(Assets, "templates/auth/*.html")
	for _, p := range authEntries {
		register("auth", authLayout, p)
	}

	// Module pages (koperasi)
	moduleDirs := []string{"dashboard", "member", "product", "order", "loan", "finance"}
	for _, dir := range moduleDirs {
		entries, _ := fs.Glob(Assets, "templates/"+dir+"/*.html")
		for _, p := range entries {
			register("base", baseLayout, p)
		}
	}

	// E-Commerce Auth pages
	ecAuthEntries, _ := fs.Glob(Assets, "templates/ecommerce/auth/*.html")
	for _, p := range ecAuthEntries {
		register("ec_auth", ecAuthLayout, p)
	}

	// E-Commerce Module pages
	ecModuleDirs := []string{"buyer", "seller", "admin", "store", "product", "order", "wishlist", "profile", "points", "components"}
	for _, dir := range ecModuleDirs {
		entries, _ := fs.Glob(Assets, "templates/ecommerce/"+dir+"/*.html")
		for _, p := range entries {
			register("ec_base", ecBaseLayout, p)
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

	// E-Commerce session data injection
	if strings.HasPrefix(layout, "ec_") {
		if _, ok := data["ECUsername"]; !ok {
			data["ECUsername"] = ecommerce.GetECUsername(c)
		}
		if _, ok := data["ECRole"]; !ok {
			data["ECRole"] = ecommerce.GetECRole(c)
		}
		if _, ok := data["IsSeller"]; !ok {
			data["IsSeller"] = ecommerce.IsECSeller(c)
		}
		// E-Commerce flash messages
		if _, ok := data["FlashError"]; !ok {
			if msg, has := handler.PopFlash(c, "ec_error"); has {
				data["FlashError"] = msg
			}
		}
		if _, ok := data["FlashSuccess"]; !ok {
			if msg, has := handler.PopFlash(c, "ec_success"); has {
				data["FlashSuccess"] = msg
			}
		}
	}

	tmpl, ok := pageTemplates[layout][page]
	if !ok {
		c.String(http.StatusInternalServerError, "template not found: %s/%s", layout, page)
		return
	}

	// Determine the entry template name
	entry := "base"
	switch layout {
	case "auth":
		entry = "auth_base"
	case "ec_base":
		entry = "ec_base"
	case "ec_auth":
		entry = "ec_auth_base"
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

	// ===== KOPERASI AUTH =====
	authH := handler.NewAuthHandler(renderPage)
	r.GET("/login", authH.ShowLogin)
	r.POST("/login", authH.DoLogin)
	r.GET("/register", authH.ShowRegister)
	r.POST("/register", authH.DoRegister)
	r.GET("/logout", authH.DoLogout)

	// ===== KOPERASI PROTECTED =====
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

	// ===== E-COMMERCE PUBLIC (no auth) =====
	ecAuthH := ecommerce.NewAuthHandler(renderPage)
	ecStoreH := ecommerce.NewStoreHandler(renderPage)

	r.GET("/ecommerce/login", ecAuthH.Login)
	r.POST("/ecommerce/login", ecAuthH.DoLogin)
	r.GET("/ecommerce/signup", ecAuthH.Signup)
	r.POST("/ecommerce/signup", ecAuthH.DoSignup)
	r.GET("/ecommerce/logout", ecAuthH.Logout)
	r.GET("/ecommerce/api/members/search", ecAuthH.SearchKoperasiMember)

	// Public store pages (browsable without login)
	r.GET("/ecommerce/store", ecStoreH.Home)
	r.GET("/ecommerce/category/:slug", ecStoreH.CategoryPage)
	r.GET("/ecommerce/products", ecStoreH.ProductCatalog)
	r.GET("/ecommerce/products/:id", ecStoreH.ProductDetail)
	r.GET("/ecommerce/seller/:id/profile", ecStoreH.SellerProfilePage)

	// ===== E-COMMERCE PROTECTED (auth required) =====
	ecBuyerH := ecommerce.NewBuyerHandler(renderPage)
	ecWishlistH := ecommerce.NewWishlistHandler(renderPage)

	ecProtected := r.Group("/ecommerce")
	ecProtected.Use(middleware.RequireECommerceAuth())
	{
		// Buyer routes
		ecProtected.GET("/buyer", ecBuyerH.Dashboard)

		// Wishlist
		ecProtected.GET("/wishlist", ecWishlistH.List)
		ecProtected.POST("/wishlist/toggle", ecWishlistH.Toggle)

		// Orders & Checkout
		ecOrderH := ecommerce.NewOrderHandler(renderPage)
		ecProtected.GET("/checkout", ecOrderH.ShowCheckout)
		ecProtected.POST("/checkout", ecOrderH.DoCheckout)
		ecProtected.GET("/orders", ecOrderH.OrderList)
		ecProtected.GET("/orders/:id/track", ecOrderH.ShowTracking)

		// Profile & Addresses
		ecProfileH := ecommerce.NewProfileHandler(renderPage)
		ecProtected.GET("/profile", ecProfileH.Profile)
		ecProtected.GET("/profile/addresses", ecProfileH.Addresses)
		ecProtected.GET("/profile/addresses/new", ecProfileH.ShowCreateAddress)
		ecProtected.POST("/profile/addresses", ecProfileH.CreateAddress)
		ecProtected.POST("/profile/addresses/:id/default", ecProfileH.SetDefaultAddress)
		ecProtected.POST("/profile/addresses/:id/delete", ecProfileH.DeleteAddress)
		ecProtected.GET("/profile/settings", ecProfileH.Settings)
		ecProtected.POST("/profile/settings", ecProfileH.UpdateSettings)

		// Reviews
		ecReviewH := ecommerce.NewReviewHandler(renderPage)
		ecProtected.GET("/order/:id/review", ecReviewH.ShowReview)
		ecProtected.POST("/order/:id/review", ecReviewH.DoReview)

		// Points & Loyalty
		ecPointsH := ecommerce.NewPointsHandler(renderPage)
		ecProtected.GET("/points", ecPointsH.Balance)
		ecProtected.GET("/points/convert", ecPointsH.ConvertForm)
		ecProtected.POST("/points/convert", ecPointsH.DoConvert)
		ecProtected.GET("/points/link-member", ecPointsH.LinkMemberPage)
		ecProtected.POST("/points/link-member", ecPointsH.DoLinkMember)
		ecProtected.GET("/api/points/conversion-status", ecPointsH.CheckConversionStatus)

		// Koperasi integration mock API
		ecIntegrationH := ecommerce.NewIntegrationHandler()
		ecProtected.GET("/api/koperasi/member/:id", ecIntegrationH.GetKoperasiMember)
		ecProtected.POST("/api/koperasi/link", ecIntegrationH.LinkToKoperasi)
		ecProtected.POST("/api/koperasi/simpanan/add", ecIntegrationH.AddToSimpanan)

		// Member linking (requires auth)
		ecProtected.POST("/link-member", ecAuthH.LinkKoperasiMember)
		ecProtected.POST("/unlink-member", ecAuthH.UnlinkKoperasiMember)

		// Seller routes (require seller status)
		ecSellerH := ecommerce.NewSellerHandler(renderPage)
		ecSeller := ecProtected.Group("/seller")
		ecSeller.Use(middleware.RequireECommerceSeller())
		{
			ecSeller.GET("", ecSellerH.Dashboard)
			ecSeller.GET("/products", ecSellerH.ProductList)
			ecSeller.GET("/products/create", ecSellerH.CreateProduct)
			ecSeller.POST("/products/create", ecSellerH.DoCreateProduct)
			ecSeller.GET("/orders", ecSellerH.OrderList)
			ecSeller.POST("/orders/:id/shipped", ecSellerH.MarkShipped)
			ecSeller.GET("/earnings", ecSellerH.Earnings)
		}

		// Admin routes (require ADMIN role)
		ecAdminH := ecommerce.NewAdminHandler(renderPage)
		ecAdmin := ecProtected.Group("/admin")
		ecAdmin.Use(middleware.RequireECommerceAdmin())
		{
			ecAdmin.GET("", ecAdminH.Dashboard)
			ecAdmin.GET("/analytics", ecAdminH.Analytics)
			ecAdmin.GET("/sellers", ecAdminH.SellerApprovals)
			ecAdmin.POST("/sellers/:id/approve", ecAdminH.ApproveSeller)
			ecAdmin.POST("/sellers/:id/reject", ecAdminH.RejectSeller)
			ecAdmin.GET("/products", ecAdminH.ProductApprovals)
			ecAdmin.POST("/products/:id/approve", ecAdminH.ApproveProduct)
			ecAdmin.POST("/products/:id/reject", ecAdminH.RejectProduct)
			ecAdmin.GET("/vouchers", ecAdminH.VoucherManagement)
			ecAdmin.POST("/vouchers/create", ecAdminH.CreateVoucher)
			ecAdmin.POST("/vouchers/:id/toggle", ecAdminH.ToggleVoucher)
			ecAdmin.GET("/orders", ecAdminH.OrderManagement)
			ecAdmin.GET("/members", ecAdminH.MemberManagement)
			ecAdmin.GET("/points", ecAdminH.PointsMonitoring)
			ecAdmin.GET("/audit", ecAdminH.AuditLog)
		}
	}

	return r
}


