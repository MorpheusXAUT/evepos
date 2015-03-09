package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/morpheusxaut/evepos/misc"

	"github.com/gorilla/mux"
)

// IndexGetHandler displays the index page of the web app
func (controller *Controller) IndexGetHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 1
	response["pageTitle"] = "Index"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	response["loggedIn"] = loggedIn
	response["status"] = 0
	response["result"] = nil

	controller.SendResponse(w, r, "index", response)
}

// LoginGetHandler displays the login page of the web app
func (controller *Controller) LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 2
	response["pageTitle"] = "Login"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	response["loggedIn"] = loggedIn

	if loggedIn {
		http.Redirect(w, r, controller.Session.GetLoginRedirect(r), http.StatusSeeOther)
		return
	}

	response["status"] = 0
	response["result"] = nil

	controller.SendResponse(w, r, "login", response)
}

// LoginPostHandler handles submitted data from the login page and verifies the user's credentials
func (controller *Controller) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 2
	response["pageTitle"] = "Login"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	response["loggedIn"] = loggedIn

	if loggedIn {
		http.Redirect(w, r, controller.Session.GetLoginRedirect(r), http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		misc.Logger.Warnf("Failed to parse form: [%v]", err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Failed to parse form, please try again!")

		controller.SendResponse(w, r, "login", response)

		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if len(username) == 0 || len(password) == 0 {
		misc.Logger.Warnf("Received empty username or password")

		response["status"] = 1
		response["result"] = fmt.Errorf("Empty username or password, please try again!")

		controller.SendResponse(w, r, "login", response)

		return
	}

	err = controller.Session.Authenticate(w, r, username, password)
	if err != nil {
		misc.Logger.Warnf("Failed to authenticate user: [%v]", err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Invalid username or password, please try again!")

		controller.SendResponse(w, r, "login", response)

		return
	}

	http.Redirect(w, r, controller.Session.GetLoginRedirect(r), http.StatusSeeOther)
}

// LoginResetGetHandler allows the user to reset their password
func (controller *Controller) LoginResetGetHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 2
	response["pageTitle"] = "Reset Password"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	if loggedIn {
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	response["loggedIn"] = loggedIn
	response["status"] = 0
	response["result"] = nil

	controller.SendResponse(w, r, "loginreset", response)
}

// LoginResetPostHandler allows the user to reset their password
func (controller *Controller) LoginResetPostHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 2
	response["pageTitle"] = "Reset Password"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	if loggedIn {
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	response["loggedIn"] = loggedIn

	err := r.ParseForm()
	if err != nil {
		misc.Logger.Warnf("Failed to parse form: [%v]", err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Failed to parse form, please try again!")

		controller.SendResponse(w, r, "loginreset", response)

		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")

	if len(username) == 0 && len(email) == 0 {
		misc.Logger.Warnf("Received empty username or email")

		response["status"] = 1
		response["result"] = fmt.Errorf("Empty username or email, please try again!")

		controller.SendResponse(w, r, "loginreset", response)

		return
	}

	err = controller.Session.SendPasswordReset(w, r, username, email)
	if err != nil {
		misc.Logger.Warnf("Failed to send password reset: [%v]", err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Failed to send password reset, please try again!")

		controller.SendResponse(w, r, "loginreset", response)

		return
	}

	response["status"] = 2
	response["result"] = "Password reset mail sent! Please use the provided link to change your password!"

	controller.SendResponse(w, r, "loginreset", response)
}

// LoginResetVerifyGetHandler provides the user with a form to reset their password
func (controller *Controller) LoginResetVerifyGetHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 2
	response["pageTitle"] = "Reset Password"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	if loggedIn {
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	response["loggedIn"] = loggedIn

	err := r.ParseForm()
	if err != nil {
		misc.Logger.Warnf("Failed to parse form: [%v]", err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Failed to parse form, please try again!")

		controller.SendResponse(w, r, "loginreset", response)

		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	verification := r.FormValue("verification")

	if len(email) == 0 || len(username) == 0 || len(verification) == 0 {
		misc.Logger.Warnf("Received empty email, username or verification code")

		response["status"] = 1
		response["result"] = fmt.Errorf("Empty email, username or verification code, please try again!")

		controller.SendResponse(w, r, "loginreset", response)

		return
	}

	response["status"] = 0
	response["result"] = nil
	response["email"] = email
	response["username"] = username
	response["verification"] = verification

	controller.SendResponse(w, r, "loginresetverify", response)
}

// LoginResetVerifyPostHandler updates the user's password as per choice
func (controller *Controller) LoginResetVerifyPostHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 2
	response["pageTitle"] = "Reset Password"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	if loggedIn {
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	response["loggedIn"] = loggedIn

	err := r.ParseForm()
	if err != nil {
		misc.Logger.Warnf("Failed to parse form: [%v]", err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Failed to parse form, please try again!")

		controller.SendResponse(w, r, "loginreset", response)

		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	verification := r.FormValue("verification")
	password := r.FormValue("password")

	if len(email) == 0 || len(username) == 0 || len(verification) == 0 || len(password) == 0 {
		misc.Logger.Warnf("Received empty email, username, verification, old or password")

		response["status"] = 1
		response["result"] = fmt.Errorf("Empty email, username, verification, old or password, please try again!")

		controller.SendResponse(w, r, "loginreset", response)

		return
	}

	err = controller.Session.VerifyPasswordReset(w, r, email, username, verification, password)
	if err != nil {
		misc.Logger.Warnf("Failed to verify password reset: [%v]", err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Failed to reset password, please try again!")

		controller.SendResponse(w, r, "loginreset", response)

		return
	}

	response["status"] = 2
	response["result"] = "Successfully changed password!"

	controller.SendResponse(w, r, "login", response)
}

// LogoutGetHandler destroys the user's current session and thus logs him out
func (controller *Controller) LogoutGetHandler(w http.ResponseWriter, r *http.Request) {
	controller.Session.DestroySession(w, r)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PosesGetHandler displays an overview over all monitored POSes and their status
func (controller *Controller) PosesGetHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 3
	response["pageTitle"] = "POSes"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	if !loggedIn {
		err := controller.Session.SetLoginRedirect(w, r, "/poses")
		if err != nil {
			misc.Logger.Warnf("Failed to set login redirect: [%v]", err)
			controller.SendRawError(w, http.StatusInternalServerError, err)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	response["loggedIn"] = loggedIn
	response["status"] = 0
	response["result"] = nil

	controller.SendResponse(w, r, "poses", response)
}

// PosDetailsGetHandler displays more information about a selected POS
func (controller *Controller) PosDetailsGetHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 3
	response["pageTitle"] = "POS details"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	if !loggedIn {
		err := controller.Session.SetLoginRedirect(w, r, "/poses")
		if err != nil {
			misc.Logger.Warnf("Failed to set login redirect: [%v]", err)
			controller.SendRawError(w, http.StatusInternalServerError, err)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	response["loggedIn"] = loggedIn

	vars := mux.Vars(r)
	posID, err := strconv.ParseInt(vars["posid"], 10, 64)
	if err != nil {
		misc.Logger.Warnf("Failed to parse POS ID %q: [%v]", vars["posid"], err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Invalid POS ID, please try again!")

		controller.SendResponse(w, r, "posdetails", response)

		return
	}

	pos, err := controller.Session.LoadPOSDetails(posID)
	if err != nil {
		misc.Logger.Warnf("Failed to load POS details: [%v]", err)

		response["status"] = 1
		response["result"] = fmt.Errorf("Failed to load POS details, please try again!")

		controller.SendResponse(w, r, "posdetails", response)

		return
	}

	response["pos"] = pos
	response["status"] = 0
	response["result"] = nil

	controller.SendResponse(w, r, "posdetails", response)
}

// LegalGetHandler displays some legal information as well as copyright disclaimers and contact info
func (controller *Controller) LegalGetHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["pageType"] = 4
	response["pageTitle"] = "Legal"

	loggedIn := controller.Session.IsLoggedIn(w, r)

	response["loggedIn"] = loggedIn
	response["status"] = 0
	response["result"] = nil

	controller.SendResponse(w, r, "legal", response)
}
