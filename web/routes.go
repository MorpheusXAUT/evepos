package web

import (
	"net/http"
)

// Route stores information about a web route being handled
type Route struct {
	// Name represents a name for the web route
	Name string
	// Methods contains all HTTP methods available to access this route
	Methods []string
	// Pattern defines the URL pattern used to match this route
	Pattern string
	// HandlerFunc represents the web handler function to call for this route
	HandlerFunc http.HandlerFunc
}

// SetupRoutes initialises all used web routes and returns them for the router
func SetupRoutes(controller *Controller) []Route {
	r := []Route{
		Route{
			Name:        "IndexGet",
			Methods:     []string{"GET"},
			Pattern:     "/",
			HandlerFunc: controller.IndexGetHandler,
		},
		Route{
			Name:        "LoginGet",
			Methods:     []string{"GET"},
			Pattern:     "/login",
			HandlerFunc: controller.LoginGetHandler,
		},
		Route{
			Name:        "LoginPost",
			Methods:     []string{"POST"},
			Pattern:     "/login",
			HandlerFunc: controller.LoginPostHandler,
		},
		Route{
			Name:        "LoginResetGet",
			Methods:     []string{"GET"},
			Pattern:     "/login/reset",
			HandlerFunc: controller.LoginResetGetHandler,
		},
		Route{
			Name:        "LoginResetPost",
			Methods:     []string{"POST"},
			Pattern:     "/login/reset",
			HandlerFunc: controller.LoginResetPostHandler,
		},
		Route{
			Name:        "LoginResetVerifyGet",
			Methods:     []string{"GET"},
			Pattern:     "/login/reset/verify",
			HandlerFunc: controller.LoginResetVerifyGetHandler,
		},
		Route{
			Name:        "LoginResetVerifyPost",
			Methods:     []string{"POST"},
			Pattern:     "/login/reset/verify",
			HandlerFunc: controller.LoginResetVerifyPostHandler,
		},
		Route{
			Name:        "LogoutGet",
			Methods:     []string{"GET"},
			Pattern:     "/logout",
			HandlerFunc: controller.LogoutGetHandler,
		},
		Route{
			Name:        "PosesGet",
			Methods:     []string{"GET"},
			Pattern:     "/poses",
			HandlerFunc: controller.PosesGetHandler,
		},
		Route{
			Name:        "PosDetailsGet",
			Methods:     []string{"GET"},
			Pattern:     "/pos/{posid:[0-9]+}",
			HandlerFunc: controller.PosDetailsGetHandler,
		},
		Route{
			Name:        "LegalGet",
			Methods:     []string{"GET"},
			Pattern:     "/legal",
			HandlerFunc: controller.LegalGetHandler,
		},
	}

	return r
}
