package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"koperasi-frontend/core/handler"
	ecommerce "koperasi-frontend/core/handler/ecommerce"
	"koperasi-frontend/core/middleware"
	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
	"koperasi-frontend/core/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Assets is assigned from outside (api/index.go or cmd/web/main.go).
// For Vercel: assigned an embed.FS. For local dev: assigned an os.DirFS.
var Assets fs.FS

type AppOptions struct {
	SessionSecret string
	SessionSecure bool
}

// sidebarMemberID me-resolve MemberID dari nama anggota untuk shortcut sidebar.
// Di-set di SetupApp saat DB terhubung; nil bila berjalan tanpa DB.
var sidebarMemberID func(ctx context.Context, nama string) int

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
	"iadd":     func(a, b int) int { return a + b },
	"contains": func(s, substr string) bool { return strings.Contains(s, substr) },
	"float64":  func(i int) float64 { return float64(i) },
	"itoa":     strconv.Itoa,
}

const csrfSessionKey = "csrf_token"

func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Next()
			return
		}

		sess := sessions.Default(c)
		token, _ := sess.Get(csrfSessionKey).(string)
		if token == "" {
			var err error
			token, err = newCSRFToken()
			if err != nil {
				c.String(http.StatusInternalServerError, "Gagal menyiapkan token keamanan.")
				c.Abort()
				return
			}
			sess.Set(csrfSessionKey, token)
			if err := sess.Save(); err != nil {
				c.String(http.StatusInternalServerError, "Gagal menyimpan token keamanan.")
				c.Abort()
				return
			}
		}
		c.Set(csrfSessionKey, token)

		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		submitted := c.GetHeader("X-CSRF-Token")
		if submitted == "" {
			submitted = c.PostForm("_csrf")
		}
		if !validCSRFToken(token, submitted) {
			c.String(http.StatusForbidden, "CSRF token tidak valid.")
			c.Abort()
			return
		}
		c.Next()
	}
}

func newCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func validCSRFToken(expected, submitted string) bool {
	if expected == "" || submitted == "" || len(expected) != len(submitted) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(submitted)) == 1
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
	if _, ok := data["CSRFToken"]; !ok {
		data["CSRFToken"] = c.GetString(csrfSessionKey)
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

	// Sidebar (base layout) needs the MemberID for ANGGOTA shortcut links.
	// sidebarMemberID di-set saat startup (DB); nil bila DB tidak terhubung.
	if _, ok := data["MemberID"]; !ok {
		if nama, ok := data["UserNama"].(string); ok && nama != "" && sidebarMemberID != nil {
			if id := sidebarMemberID(c.Request.Context(), nama); id > 0 {
				data["MemberID"] = id
			}
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

func SetupApp(db *gorm.DB, opts AppOptions) *gin.Engine {
	if len(opts.SessionSecret) < 32 {
		panic("server: session secret must be at least 32 characters")
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	loadTemplates()

	store := cookie.NewStore([]byte(opts.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   opts.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("koperasi_session", store))
	r.Use(csrfMiddleware())

	// Serve static files from embedded FS
	staticFS, err := fs.Sub(Assets, "static")
	if err != nil {
		log.Fatalf("failed to sub static FS: %v", err)
	}
	r.StaticFS("/static", http.FS(staticFS))

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard")
	})

	// ===== KOPERASI PROTECTED =====
	// Fase 2: M1 Keanggotaan dari DB (repository -> service -> handler).
	// Jika db nil (mis. DATABASE_URL belum diset), service nil & handler memberi pesan.
	var memSvc *service.MemberService
	var prodSvc *service.ProductService
	var orderSvc *service.OrderService
	var loanSvc *service.LoanService
	var finSvc *service.FinanceService
	var ecStoreSvc *service.ECStoreService
	var ecAcctSvc *service.ECAccountService
	var ecShopSvc *service.ECShopService
	var ecSellerSvc *service.ECSellerService
	var ecAdminSvc *service.ECAdminService
	var dashSvc *service.DashboardService
	var authSvc *service.AuthService
	if db != nil {
		authSvc = service.NewAuthService(repository.NewUserRepository(db))
		memberRepo := repository.NewMemberRepository(db)
		prodRepo := repository.NewProductRepository(db)
		orderRepo := repository.NewOrderRepository(db)
		loanRepo := repository.NewLoanRepository(db)
		memSvc = service.NewMemberService(memberRepo)
		// Fase 4a: M3 POS dari DB (produk, stok, order, jurnal penjualan).
		prodSvc = service.NewProductService(prodRepo)
		orderSvc = service.NewOrderService(orderRepo)
		// Fase 4b: M4 Pinjaman dari DB (loans, installments, jurnal pinjaman).
		loanSvc = service.NewLoanService(loanRepo)
		// Fase 4c: M2 Keuangan dari DB (jurnal, ringkasan, jurnal manual).
		finSvc = service.NewFinanceService(repository.NewJournalRepository(db))
		// Fase 4d-1: M5 E-Commerce storefront/katalog dari DB.
		ecStoreSvc = service.NewECStoreService(repository.NewECommerceRepository(db))
		// Fase 4d-2: akun EC (auth/SSO, integrasi koperasi, dashboard buyer) dari DB.
		ecAcctSvc = service.NewECAccountService(repository.NewECommerceRepository(db))
		// Fase 4d-3: alur beli EC (checkout, order, wishlist) dari DB.
		ecShopSvc = service.NewECShopService(repository.NewECommerceRepository(db))
		// Fase 4d-5: seller (dashboard, produk, order, earnings) dari DB.
		ecSellerSvc = service.NewECSellerService(repository.NewECommerceRepository(db))
		// Fase 4d-6: admin (moderasi, voucher, analytics, audit) dari DB.
		ecAdminSvc = service.NewECAdminService(repository.NewECommerceRepository(db))
		// Fase 4e: dashboard agregat dari DB + sidebar MemberID (lepas mock dari runtime).
		dashSvc = service.NewDashboardService(memberRepo, loanRepo, prodRepo, orderRepo)
		sidebarMemberID = func(ctx context.Context, nama string) int {
			if m, _ := memberRepo.FindByNama(ctx, nama); m != nil {
				return m.ID
			}
			return 0
		}
	}

	// ===== KOPERASI AUTH =====
	authH := handler.NewAuthHandler(renderPage, authSvc)
	r.GET("/login", authH.ShowLogin)
	r.POST("/login", authH.DoLogin)
	r.GET("/register", authH.ShowRegister)
	r.POST("/register", authH.DoRegister)
	r.GET("/logout", authH.DoLogout)

	dashH := handler.NewDashboardHandler(renderPage, dashSvc)
	memH := handler.NewMemberHandler(renderPage, memSvc)
	prodH := handler.NewProductHandler(renderPage, prodSvc)
	orderH := handler.NewOrderHandler(renderPage, orderSvc, prodSvc)
	loanH := handler.NewLoanHandler(renderPage, loanSvc)
	finH := handler.NewFinanceHandler(renderPage, finSvc)

	protected := r.Group("/")
	protected.Use(handler.AuthRequired())
	{
		// Dashboard — handler internally routes to staff/anggota view by role
		protected.GET("/dashboard", dashH.Index)

		// ===== Keanggotaan =====
		// Staff-only views: full member list, registration by pengurus, approve/reject.
		protected.GET("/members", handler.RequireRole("OWNER", "KASIR"), memH.List)
		// Form pendaftaran — anggota baru self-register, OWNER bisa daftarkan dari sisi pengurus.
		// KASIR tidak boleh: tugas KASIR adalah operasional, bukan administratif.
		protected.GET("/members/register", handler.RequireRole("OWNER", "ANGGOTA"), memH.ShowRegister)
		protected.POST("/members/register", handler.RequireRole("OWNER", "ANGGOTA"), memH.DoRegister)
		protected.POST("/members/:id/approve", handler.RequireRole("OWNER"), memH.Approve)
		protected.POST("/members/:id/reject", handler.RequireRole("OWNER"), memH.Reject)
		// Detail/simpanan/resign — ownership-guarded inside the handler (ANGGOTA only own).
		protected.GET("/members/:id", memH.Detail)
		protected.GET("/members/:id/simpanan", memH.ShowSimpanan)
		protected.POST("/members/:id/simpanan", memH.DoSimpanan)
		protected.GET("/members/:id/resign", memH.ShowResign)
		protected.POST("/members/:id/resign", memH.DoResign)

		// ===== Produk =====
		// Browsing & detail open to all logged-in users. Staff actions gated.
		protected.GET("/products", prodH.Catalog)
		protected.GET("/products/:id", prodH.Detail)
		protected.GET("/products/create", handler.RequireRole("OWNER", "KASIR"), prodH.ShowCreate)
		protected.POST("/products/create", handler.RequireRole("OWNER", "KASIR"), prodH.DoCreate)
		protected.GET("/products/review", handler.RequireRole("OWNER"), prodH.Review)
		protected.POST("/products/:id/approve", handler.RequireRole("OWNER"), prodH.Approve)
		protected.POST("/products/:id/reject", handler.RequireRole("OWNER"), prodH.Reject)
		protected.GET("/products/:id/stock", handler.RequireRole("OWNER", "KASIR"), prodH.ShowStock)
		protected.POST("/products/:id/stock", handler.RequireRole("OWNER", "KASIR"), prodH.DoStock)

		// ===== Cart & Order =====
		// Cart/checkout — every authenticated user can shop.
		protected.GET("/cart", orderH.ShowCart)
		protected.POST("/cart/add", orderH.AddToCart)
		protected.POST("/cart/update", orderH.UpdateCart)
		protected.POST("/cart/remove", orderH.RemoveFromCart)
		protected.GET("/checkout", orderH.ShowCheckout)
		protected.POST("/checkout", orderH.DoCheckout)
		// Order list — ANGGOTA filtered to own orders inside the handler.
		protected.GET("/orders", orderH.List)
		// Order detail — ownership-guarded inside the handler.
		protected.GET("/orders/:id", orderH.Detail)
		// Konfirmasi terima oleh pembeli sendiri — boleh ANGGOTA (handler cek ownership).
		protected.POST("/orders/:id/complete", orderH.Complete)
		// Komplain hanya boleh oleh pemilik order (handler cek ownership).
		protected.GET("/orders/:id/complain", orderH.ShowComplain)
		protected.POST("/orders/:id/complain", orderH.DoComplain)
		// Tindakan staff: kirim & daftar/resolve komplain.
		protected.POST("/orders/:id/ship", handler.RequireRole("OWNER", "KASIR"), orderH.Ship)
		protected.GET("/orders/complaints", handler.RequireRole("OWNER", "KASIR"), orderH.Complaints)
		protected.POST("/orders/:id/resolve", handler.RequireRole("OWNER"), orderH.Resolve)

		// ===== Pinjaman =====
		// List — ANGGOTA filtered to own loans inside the handler.
		protected.GET("/loans", loanH.List)
		protected.GET("/loans/apply", loanH.ShowApply)
		protected.POST("/loans/apply", loanH.DoApply)
		// Detail — ownership-guarded inside the handler.
		protected.GET("/loans/:id", loanH.Detail)
		// Pay angsuran — pemilik pinjaman bisa bayar sendiri (handler cek ownership).
		protected.POST("/loans/:id/pay", loanH.Pay)
		// Tindakan pengurus.
		protected.POST("/loans/:id/approve", handler.RequireRole("OWNER"), loanH.Approve)
		protected.POST("/loans/:id/reject", handler.RequireRole("OWNER"), loanH.Reject)
		protected.POST("/loans/:id/disburse", handler.RequireRole("OWNER"), loanH.Disburse)

		// ===== Keuangan Koperasi (OWNER only — jurnal & ringkasan kas) =====
		protected.GET("/finance/journals", handler.RequireRole("OWNER"), finH.Journals)
		protected.GET("/finance/journals/create", handler.RequireRole("OWNER"), finH.ShowCreateJournal)
		protected.POST("/finance/journals/create", handler.RequireRole("OWNER"), finH.DoCreateJournal)
		protected.GET("/finance/summary", handler.RequireRole("OWNER"), finH.Summary)
	}

	// ===== E-COMMERCE PUBLIC (no auth) =====
	ecAuthH := ecommerce.NewAuthHandler(renderPage, ecAcctSvc)
	ecStoreH := ecommerce.NewStoreHandler(renderPage, ecStoreSvc)

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
	ecBuyerH := ecommerce.NewBuyerHandler(renderPage, ecAcctSvc)
	ecWishlistH := ecommerce.NewWishlistHandler(renderPage, ecShopSvc)

	ecProtected := r.Group("/ecommerce")
	ecProtected.Use(middleware.RequireECommerceAuth(ecAcctSvc))
	{
		// Buyer routes
		ecProtected.GET("/buyer", ecBuyerH.Dashboard)

		// Wishlist
		ecProtected.GET("/wishlist", ecWishlistH.List)
		ecProtected.POST("/wishlist/toggle", ecWishlistH.Toggle)

		// Orders & Checkout
		ecOrderH := ecommerce.NewOrderHandler(renderPage, ecShopSvc)
		ecProtected.GET("/checkout", ecOrderH.ShowCheckout)
		ecProtected.POST("/checkout", ecOrderH.DoCheckout)
		ecProtected.GET("/orders", ecOrderH.OrderList)
		ecProtected.GET("/orders/:id/track", ecOrderH.ShowTracking)

		// Profile & Addresses
		ecProfileH := ecommerce.NewProfileHandler(renderPage, ecAcctSvc)
		ecProtected.GET("/profile", ecProfileH.Profile)
		ecProtected.GET("/profile/addresses", ecProfileH.Addresses)
		ecProtected.GET("/profile/addresses/new", ecProfileH.ShowCreateAddress)
		ecProtected.POST("/profile/addresses", ecProfileH.CreateAddress)
		ecProtected.POST("/profile/addresses/:id/default", ecProfileH.SetDefaultAddress)
		ecProtected.POST("/profile/addresses/:id/delete", ecProfileH.DeleteAddress)
		ecProtected.GET("/profile/settings", ecProfileH.Settings)
		ecProtected.POST("/profile/settings", ecProfileH.UpdateSettings)

		// Reviews
		ecReviewH := ecommerce.NewReviewHandler(renderPage, ecShopSvc)
		ecProtected.GET("/order/:id/review", ecReviewH.ShowReview)
		ecProtected.POST("/order/:id/review", ecReviewH.DoReview)

		// Points & Loyalty
		ecPointsH := ecommerce.NewPointsHandler(renderPage, ecAcctSvc)
		ecProtected.GET("/points", ecPointsH.Balance)
		ecProtected.GET("/points/convert", ecPointsH.ConvertForm)
		ecProtected.POST("/points/convert", ecPointsH.DoConvert)
		ecProtected.GET("/points/link-member", ecPointsH.LinkMemberPage)
		ecProtected.POST("/points/link-member", ecPointsH.DoLinkMember)
		ecProtected.GET("/api/points/conversion-status", ecPointsH.CheckConversionStatus)

		// Koperasi integration API
		ecIntegrationH := ecommerce.NewIntegrationHandler(ecAcctSvc)
		ecProtected.GET("/api/koperasi/member/:id", ecIntegrationH.GetKoperasiMember)
		ecProtected.POST("/api/koperasi/link", ecIntegrationH.LinkToKoperasi)
		ecProtected.POST("/api/koperasi/simpanan/add", ecIntegrationH.AddToSimpanan)

		// Member linking (requires auth)
		ecProtected.POST("/link-member", ecAuthH.LinkKoperasiMember)
		ecProtected.POST("/unlink-member", ecAuthH.UnlinkKoperasiMember)

		// Seller routes (require seller status)
		ecSellerH := ecommerce.NewSellerHandler(renderPage, ecSellerSvc)
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
		ecAdminH := ecommerce.NewAdminHandler(renderPage, ecAdminSvc)
		ecAdmin := ecProtected.Group("/admin")
		// Area admin terbuka untuk ADMIN dan PENGURUS (moderator operasional).
		ecAdmin.Use(middleware.RequireECommerceAdminArea())
		{
			// adminOnly menjaga rute sensitif agar tetap eksklusif ADMIN (PENGURUS ditolak).
			adminOnly := middleware.RequireECommerceAdmin()

			ecAdmin.GET("", ecAdminH.Dashboard)
			ecAdmin.GET("/analytics", ecAdminH.Analytics)
			ecAdmin.GET("/sellers", ecAdminH.SellerApprovals)
			ecAdmin.POST("/sellers/:id/approve", ecAdminH.ApproveSeller)
			ecAdmin.POST("/sellers/:id/reject", ecAdminH.RejectSeller)
			ecAdmin.GET("/products", ecAdminH.ProductApprovals)
			ecAdmin.POST("/products/:id/approve", ecAdminH.ApproveProduct)
			ecAdmin.POST("/products/:id/reject", ecAdminH.RejectProduct)
			ecAdmin.GET("/vouchers", adminOnly, ecAdminH.VoucherManagement)
			ecAdmin.POST("/vouchers/create", adminOnly, ecAdminH.CreateVoucher)
			ecAdmin.POST("/vouchers/:id/toggle", adminOnly, ecAdminH.ToggleVoucher)
			ecAdmin.GET("/orders", ecAdminH.OrderManagement)
			ecAdmin.GET("/members", adminOnly, ecAdminH.MemberManagement)
			ecAdmin.GET("/points", ecAdminH.PointsMonitoring)
			ecAdmin.GET("/audit", adminOnly, ecAdminH.AuditLog)
		}
	}

	return r
}
