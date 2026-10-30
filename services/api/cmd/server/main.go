package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"globalvideoplatform/services/api/internal/admin"
	"globalvideoplatform/services/api/internal/ads"
	"globalvideoplatform/services/api/internal/channels"
	"globalvideoplatform/services/api/internal/comments"
	"globalvideoplatform/services/api/internal/config"
	"globalvideoplatform/services/api/internal/copyright"
	"globalvideoplatform/services/api/internal/events"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/live"
	"globalvideoplatform/services/api/internal/moderation"
	"globalvideoplatform/services/api/internal/partners"
	"globalvideoplatform/services/api/internal/payments"
	"globalvideoplatform/services/api/internal/playlists"
	"globalvideoplatform/services/api/internal/recommendations"
	"globalvideoplatform/services/api/internal/security"
	"globalvideoplatform/services/api/internal/settlements"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/studio"
	"globalvideoplatform/services/api/internal/subscriptions"
	"globalvideoplatform/services/api/internal/uploads"
	"globalvideoplatform/services/api/internal/users"
	"globalvideoplatform/services/api/internal/videos"
)

func buildRouter(memStore *store.MemoryStore) http.Handler {
	authMiddleware := identity.NewAuthMiddleware(memStore)

	// Security infrastructure (TASK-010~028, TASK-084~085)
	rateLimiter := security.NewRateLimiter(1200, time.Minute) // 1200 req/min per IP (generous for tests/CI)
	idempotencyManager := security.NewIdempotencyManager(24 * time.Hour)
	adminGuard := security.NewAdminGuard()

	r := chi.NewRouter()

	// Global Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Security hardening headers (HSTS, CSP nonce, X-Frame-Options, etc.)
	r.Use(security.SecurityHeadersMiddleware)

	// Sliding-window rate limiting per client IP
	r.Use(rateLimiter.Middleware)

	// CORS Configuration for Nuxt 4 web consoles & mobile Dio
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "https://*.brand.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-Id", "X-Idempotency-Key", "X-Webhook-Signature"},
		ExposedHeaders:   []string{"Link", "X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(authMiddleware.Authenticate)
	// Validate current credentials before considering any cached response.
	r.Use(idempotencyManager.Middleware)

	// Liveness Probe (TASK-086)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":    "UP",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	// Readiness Probe (TASK-086)
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if memStore == nil {
			identity.WriteError(w, r, http.StatusServiceUnavailable, "NOT_READY", "Storage subsystem not ready")
			return
		}
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":    "READY",
			"ready":     true,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"subsystems": map[string]string{
				"storage":   "operational",
				"security":  "operational",
				"ad_engine": "operational",
			},
		})
	})

	// Prometheus Metrics Endpoint (TASK-087)
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		metrics := fmt.Sprintf(
			"# HELP gvp_uptime_seconds Total server uptime in seconds\n"+
				"# TYPE gvp_uptime_seconds gauge\ngvp_uptime_seconds %d\n"+
				"# HELP gvp_users_total Total registered users\n"+
				"# TYPE gvp_users_total gauge\ngvp_users_total %d\n"+
				"# HELP gvp_videos_total Total catalog videos\n"+
				"# TYPE gvp_videos_total gauge\ngvp_videos_total %d\n"+
				"# HELP gvp_channels_total Total channels\n"+
				"# TYPE gvp_channels_total gauge\ngvp_channels_total %d\n"+
				"# HELP gvp_ad_events_total Total collected ad telemetry events\n"+
				"# TYPE gvp_ad_events_total counter\ngvp_ad_events_total %d\n",
			time.Now().Unix(),
			len(memStore.Users),
			len(memStore.Videos),
			len(memStore.Channels),
			len(memStore.AdEvents),
		)
		w.Write([]byte(metrics))
	})

	// Handlers
	userH := users.NewHandler(memStore)
	channelH := channels.NewHandler(memStore)
	videoH := videos.NewHandler(memStore)
	uploadH := uploads.NewHandler(memStore)
	recomH := recommendations.NewHandler(memStore)
	commH := comments.NewHandler(memStore)
	liveH := live.NewHandler(memStore)
	adsH := ads.NewHandler(memStore)
	settleH := settlements.NewHandler(memStore)
	copyH := copyright.NewHandler(memStore)
	modH := moderation.NewHandler(memStore)
	adminH := admin.NewHandler(memStore)
	studioH := studio.NewHandler(memStore)
	partnerH := partners.NewHandler(memStore)
	eventH := events.NewHandler(memStore)
	playlistH := playlists.NewHandler(memStore)
	subH := subscriptions.NewHandler(memStore)
	payH := payments.NewHandler(memStore)

	// API v1 Routing Tree
	r.Route("/v1", func(v1 chi.Router) {
		// Users / Me
		v1.Mount("/", userH.Routes(authMiddleware))

		// Channels & Subscriptions
		v1.Mount("/channels", channelH.Routes(authMiddleware))
		v1.Mount("/subscriptions", subH.Routes(authMiddleware))

		// Playlists
		v1.Mount("/playlists", playlistH.Routes(authMiddleware))

		// Payments & Double-Entry Ledger Webhooks
		v1.Mount("/payments", payH.Routes(authMiddleware))

		// Videos & Playback
		v1.Mount("/videos", videoH.Routes(authMiddleware))
		v1.Get("/videos/{id}/comments", commH.GetComments)
		v1.With(authMiddleware.RequireAuth).Post("/videos/{id}/comments", commH.AddComment)
		v1.With(authMiddleware.RequireAuth).Delete("/comments/{id}", commH.DeleteComment)
		v1.With(authMiddleware.RequireAuth).Post("/comments/{id}/like", commH.LikeComment)
		v1.With(authMiddleware.RequireAuth).Post("/comments/{id}/heart", commH.HeartComment)
		v1.With(authMiddleware.RequireAuth).Post("/comments/{id}/pin", commH.PinComment)

		// Uploads
		v1.Mount("/uploads", uploadH.Routes(authMiddleware))

		// Feeds & Search
		v1.Mount("/feed", recomH.Routes(authMiddleware))
		v1.Get("/search", recomH.Search)
		v1.Get("/search/suggestions", recomH.SearchSuggestions)

		// Live Streams
		v1.Mount("/live-streams", liveH.Routes(authMiddleware))

		// Ad Server, Decision Engine & Diagnostics (ARTICLE 6 & 7)
		v1.Mount("/ads", adsH.Routes(authMiddleware))
		v1.Mount("/advertiser", adsH.Routes(authMiddleware))
		v1.Get("/advertiser/reports", adsH.GetAdvertiserReports)

		// Studio
		v1.Mount("/studio", studioH.Routes(authMiddleware))
		v1.With(authMiddleware.RequireAuth).Get("/studio/earnings", settleH.GetEarnings)

		// Partners
		v1.Mount("/partners", partnerH.Routes(authMiddleware))
		v1.With(authMiddleware.RequireAuth).Get("/partners/campaigns", adsH.ListCampaigns)
		v1.With(authMiddleware.RequireAuth).Post("/partners/campaigns", adsH.CreateCampaign)
		v1.With(authMiddleware.RequireAuth).Get("/partners/settlements", settleH.GetSettlements)
		v1.With(authMiddleware.RequireAuth).Get("/partners/rights/claims", copyH.ListClaims)
		v1.With(authMiddleware.RequireAuth).Post("/partners/rights/claims", copyH.CreateClaim)

		// Admin
		v1.Group(func(admin chi.Router) {
			admin.Use(adminGuard.IPWhitelistMiddleware)
			admin.Mount("/admin", adminH.Routes(authMiddleware))
			admin.Get("/admin/ads/delivery-diagnostics", adsH.GetDeliveryDiagnostics)
			admin.With(authMiddleware.RequireRole("super_admin", "moderator")).Get("/admin/moderation", modH.ListQueue)
			admin.With(authMiddleware.RequireRole("super_admin", "moderator")).Post("/admin/moderation/{id}/resolve", modH.ResolveCase)
		})

		// Events & Analytics Ingestion
		v1.Mount("/events", eventH.Routes())
	})

	return r
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	var pgStore *store.PostgresStore
	if cfg.DatabaseURL != "" {
		var err error
		pgStore, err = store.NewPostgresStore(cfg.DatabaseURL)
		if err != nil {
			slog.Warn("PostgreSQL not immediately reachable, falling back to resilient in-memory store", "error", err)
		} else {
			slog.Info("PostgreSQL 18 connection pool initialized", "url", store.SanitizeDatabaseURL(cfg.DatabaseURL))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := pgStore.InitSchema(ctx); err != nil {
				slog.Warn("PostgreSQL schema initialization deferred", "error", err)
			} else {
				slog.Info("PostgreSQL 18 tables verified and ready")
			}
			cancel()
		}
	}

	memStore := store.NewMemoryStoreWithDB(pgStore)
	r := buildRouter(memStore)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Starting Global Video Platform Go API Server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server listen failure", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced shutdown", "error", err)
	}
	slog.Info("Server stopped cleanly")
	fmt.Println("Server stopped")
}
