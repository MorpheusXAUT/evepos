package session

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/morpheusxaut/evepos/database"
	"github.com/morpheusxaut/evepos/mail"
	"github.com/morpheusxaut/evepos/misc"
	"github.com/morpheusxaut/evepos/models"

	"github.com/boj/redistore"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/morpheusxaut/eveapi"
	"golang.org/x/crypto/bcrypt"
)

// Controller provides functionality to handle sessions and cached values as well as retrieval of data
type Controller struct {
	config   *misc.Configuration
	database database.Connection
	mail     *mail.Controller
	store    *redistore.RediStore

	poses               []*models.POS
	reminders           map[int64]*models.POSFuelReminder
	expiryTime          time.Time
	refreshTimer        *time.Timer
	refreshChan         chan bool
	emailReminderTicker *time.Ticker
	emailReminderChan   chan bool
}

// SetupSessionController prepares the controller's session store and sets a default session lifespan
func SetupSessionController(conf *misc.Configuration, db database.Connection, mailer *mail.Controller) (*Controller, error) {
	controller := &Controller{
		config:              conf,
		database:            db,
		mail:                mailer,
		poses:               make([]*models.POS, 0),
		reminders:           make(map[int64]*models.POSFuelReminder),
		expiryTime:          time.Time{},
		refreshTimer:        &time.Timer{},
		refreshChan:         make(chan bool),
		emailReminderTicker: time.NewTicker(60 * time.Minute),
		emailReminderChan:   make(chan bool),
	}

	store, err := redistore.NewRediStoreWithDB(10, "tcp", controller.config.RedisHost, controller.config.RedisPassword, controller.config.RedisDB, securecookie.GenerateRandomKey(64), securecookie.GenerateRandomKey(32))
	if err != nil {
		return nil, err
	}

	controller.store = store

	controller.store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   604800,
		HttpOnly: true,
	}

	gob.Register(&models.User{})

	return controller, nil
}

func (controller *Controller) StartRefreshTimer() {
	go func() {
		for {
			select {
			case <-controller.refreshTimer.C:
				misc.Logger.Debugln("Updating cache...")
				controller.RefreshCache()
				controller.refreshTimer = time.NewTimer(controller.expiryTime.Sub(time.Now()))
				misc.Logger.Debugf("Next cache update scheduled in %v", controller.expiryTime.Sub(time.Now()))
				controller.emailReminderChan <- true
			case <-controller.refreshChan:
				misc.Logger.Debugln("Updating cache, manually triggered...")
				controller.RefreshCache()
				controller.refreshTimer = time.NewTimer(controller.expiryTime.Sub(time.Now()))
				misc.Logger.Debugf("Next cache update scheduled in %v (manual trigger)", controller.expiryTime.Sub(time.Now()))
				controller.emailReminderChan <- true
			}
		}
	}()

	controller.refreshChan <- true
}

func (controller *Controller) RefreshCache() {
	var poses []*models.POS

	apiKeys, err := controller.database.LoadAllAPIKeys()
	if err != nil {
		misc.Logger.Errorf("Failed to load all API keys: [%v]", err)
		return
	}

	for _, apiKey := range apiKeys {
		api := eveapi.Simple(apiKey)

		starbaseList, err := api.CorpStarbaseList()
		if err != nil {
			misc.Logger.Errorf("Failed to retrieve starbase list: [%v]", err)
			return
		}

		for _, starbase := range starbaseList.Starbases {
			starbaseDetails, err := api.CorpStarbaseDetails(starbase.ID)
			if err != nil {
				misc.Logger.Errorf("Failed to retrieve starbase details for #%d: [%v]", starbase.ID, err)
				return
			}

			var posFuel *models.POSFuel

			for _, fuel := range starbaseDetails.Fuel {
				if fuel.TypeID == 4051 || fuel.TypeID == 4246 || fuel.TypeID == 4247 || fuel.TypeID == 4312 {
					fuelUsage, err := controller.database.QueryFuelUsage(starbase.TypeID, fuel.TypeID)
					if err != nil {
						misc.Logger.Errorf("Failed to query fuel usage: [%v]", err)
						return
					}

					fuelName, err := controller.database.QueryTypeName(fuel.TypeID)
					if err != nil {
						misc.Logger.Errorf("Failed to query type name: [%v]", err)
						return
					}

					posFuel = models.NewPOSFuel(fuel.TypeID, fuelName, fuelUsage, fuel.Quantity)
					break
				}
			}

			starbaseName, err := controller.database.QueryStarbaseName(starbase.ID)
			if err != nil {
				misc.Logger.Errorf("Failed to query starbase name: [%v]", err)
				return
			}

			capacity, err := controller.database.QueryCapacity(starbase.TypeID)
			if err != nil {
				misc.Logger.Errorf("Failed to query capacity: [%v]", err)
				return
			}

			poses = append(poses, models.NewPOS(starbase, starbaseDetails, posFuel, starbaseName, capacity))
		}

		controller.expiryTime = starbaseList.APIResult.CachedUntil.Time
	}

	controller.poses = poses
}

func (controller *Controller) StartEmailReminderTicker() {
	go func() {
		for {
			select {
			case <-controller.emailReminderTicker.C:
				misc.Logger.Debugln("Checking POS fuel reminder...")
				controller.CheckEmailReminder()
				misc.Logger.Debugln("Next POS fuel reminder check scheduled in 60 minutes...")
				break
			case <-controller.emailReminderChan:
				misc.Logger.Debugln("Checking POS fuel reminder, manually triggered...")
				controller.CheckEmailReminder()
				misc.Logger.Debugln("Next POS fuel reminder check scheduled in ??? minutes (manual trigger)...")
				break
			}
		}
	}()
}

func (controller *Controller) CheckEmailReminder() {
	var lowPoses []*models.POS

	for _, pos := range controller.poses {
		if pos.Base.State == 4 {
			misc.Logger.Tracef("Reducing fuel (%d left, deducing %d) for POS #%d...", pos.Fuel.Quantity, pos.Fuel.Usage, pos.Base.ID)

			pos.Fuel.Quantity -= pos.Fuel.Usage
			remainingHours := pos.Fuel.Quantity / pos.Fuel.Usage

			_, ok := controller.reminders[pos.Base.ID]
			if ok && remainingHours > 36 {
				misc.Logger.Tracef("POS #%d has fuel > 36h (%dh left), removing from reminder list...", remainingHours, pos.Base.ID)

				delete(controller.reminders, pos.Base.ID)
			} else if ok && remainingHours <= 36 {
				misc.Logger.Tracef("POS #%d still low on fuel (%dh left), reminder sent out already!", pos.Base.ID, remainingHours)
			} else if !ok && remainingHours <= 36 {
				misc.Logger.Tracef("POS #%d low on fuel (%dh left), adding to reminder list...", pos.Base.ID, remainingHours)

				lowPoses = append(lowPoses, pos)
				controller.reminders[pos.Base.ID] = models.NewPOSFuelReminder(pos)
			} else {
				misc.Logger.Tracef("POS #%d still has enough fuel (%dh left), skipping...", pos.Base.ID, remainingHours)
			}
		}
	}

	if len(lowPoses) == 0 {
		misc.Logger.Debugln("No POSes low on fuel or all remembers already sent. YAY \\o/")
		return
	}

	users, err := controller.database.LoadAllUsers()
	if err != nil {
		misc.Logger.Errorf("Failed to load all users: [%v]", err)
		return
	}

	for _, user := range users {
		err = controller.mail.SendFuelReminder(user.Username, user.Email, lowPoses)
		if err != nil {
			misc.Logger.Errorf("Failed to send fuel reminder: [%v]", err)
		}
	}
}

func (controller *Controller) CalculateFuelShoppingList(poses []*models.POS) (*models.FuelShoppingList, error) {
	var fuelList []*models.Fuel

	for _, pos := range poses {
		if pos.Base.State == 4 {
			missingFuel := (pos.Capacity / 5) - pos.Fuel.Quantity
			if missingFuel <= 0 {
				continue
			}

			addFuel := true
			for _, fuel := range fuelList {
				if fuel.TypeID == pos.Fuel.TypeID {
					fuel.Quantity += missingFuel
					fuel.Volume = fuel.Quantity * 5
					addFuel = false
					break
				}
			}

			if addFuel {
				fuelList = append(fuelList, models.NewFuel(pos.Fuel.TypeID, pos.Fuel.TypeName, missingFuel))
			}
		}
	}

	fuelShoppingList := models.NewFuelShoppingList(fuelList)

	return fuelShoppingList, nil
}

// DestroySession destroys a user's session by setting a negative maximum age
func (controller *Controller) DestroySession(w http.ResponseWriter, r *http.Request) {
	loginSession, _ := controller.store.Get(r, "eveposLogin")
	dataSession, _ := controller.store.Get(r, "eveposData")

	loginSession.Options.MaxAge = -1
	dataSession.Options.MaxAge = -1

	err := sessions.Save(r, w)
	if err != nil {
		misc.Logger.Errorf("Failed to destroy session: [%v]", err)
	}
}

// IsLoggedIn checks whether the user is currently logged in and has an appropriate timestamp set
func (controller *Controller) IsLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	loginSession, _ := controller.store.Get(r, "eveposLogin")

	if loginSession.IsNew {
		return false
	}

	timeStamp, ok := loginSession.Values["timestamp"].(int64)
	if !ok {
		return false
	}

	if time.Now().Sub(time.Unix(timeStamp, 0)).Minutes() >= 168 {
		controller.DestroySession(w, r)
		return false
	}

	verifiedEmail, ok := loginSession.Values["verifiedEmail"].(bool)
	if !ok {
		return false
	}

	if !verifiedEmail {
		return false
	}

	return true
}

// SetLoginRedirect saves the given path as a redirect after successful login
func (controller *Controller) SetLoginRedirect(w http.ResponseWriter, r *http.Request, redirect string) error {
	loginSession, _ := controller.store.Get(r, "eveposLogin")

	loginSession.Values["loginRedirect"] = redirect

	return loginSession.Save(r, w)
}

// GetLoginRedirect retrieves the previously set path for redirection after login
func (controller *Controller) GetLoginRedirect(r *http.Request) string {
	loginSession, _ := controller.store.Get(r, "eveposLogin")

	if loginSession.IsNew {
		return "/"
	}

	redirect, ok := loginSession.Values["loginRedirect"].(string)
	if !ok {
		return "/"
	}

	return redirect
}

// Authenticate validates the given username and password against the database and creates a new session with timestamp if successful
func (controller *Controller) Authenticate(w http.ResponseWriter, r *http.Request, username string, password string) error {
	storedPassword, err := controller.database.LoadPasswordForUser(username)

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))

	loginAttempt := models.NewLoginAttempt(username, r.RemoteAddr, r.UserAgent(), (err == nil))

	logErr := controller.database.SaveLoginAttempt(loginAttempt)
	if logErr != nil {
		misc.Logger.Errorf("Failed to log login attempt: [%v]", logErr)
	}

	if err != nil {
		return err
	}

	user, err := controller.database.LoadUserFromUsername(username)
	if err != nil {
		return err
	}

	user, err = controller.SetUser(w, r, user)
	if err != nil {
		return err
	}

	loginSession, _ := controller.store.Get(r, "eveposLogin")

	loginSession.Values["username"] = user.Username
	loginSession.Values["userID"] = user.ID
	loginSession.Values["timestamp"] = time.Now().Unix()
	loginSession.Values["verifiedEmail"] = user.VerifiedEmail

	loginSession.Options.MaxAge = 604800

	return sessions.Save(r, w)
}

// SendPasswordReset sends an email with a verification link to reset a user's password to the given address
func (controller *Controller) SendPasswordReset(w http.ResponseWriter, r *http.Request, username string, email string) error {
	user, err := controller.database.LoadUserFromUsername(username)
	if err != nil {
		return err
	}

	if !strings.EqualFold(email, user.Email) {
		return fmt.Errorf("Email addresses do not match")
	}

	verification := misc.GenerateRandomString(32)

	err = controller.mail.SendPasswordReset(username, email, verification)
	if err != nil {
		return err
	}

	loginSession, _ := controller.store.Get(r, "eveposLogin")

	loginSession.Values["passwordReset"] = verification

	return sessions.Save(r, w)
}

func (controller *Controller) VerifyPasswordReset(w http.ResponseWriter, r *http.Request, email string, username string, verification string, password string) error {
	user, err := controller.database.LoadUserFromUsername(username)
	if err != nil {
		return err
	}

	if !strings.EqualFold(email, user.Email) {
		return fmt.Errorf("Email addresses do not match")
	}

	loginSession, _ := controller.store.Get(r, "eveposLogin")

	passwordReset, ok := loginSession.Values["passwordReset"].(string)
	if !ok {
		return fmt.Errorf("Failed to retrieve password reset code from login session")
	}

	if !strings.EqualFold(passwordReset, verification) {
		return fmt.Errorf("Failed to verify password reset code")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	user, err = controller.SetUser(w, r, user)
	if err != nil {
		return err
	}

	return nil
}

func (controller *Controller) LoadPOSes() ([]*models.POS, error) {
	if time.Now().After(controller.expiryTime) {
		misc.Logger.Debugln("Cache expired, manually triggering update")
		controller.refreshChan <- true
	}

	return controller.poses, nil
}

// GetUser returns the user-object stored in the data session
func (controller *Controller) GetUser(r *http.Request) (*models.User, error) {
	dataSession, _ := controller.store.Get(r, "eveposData")

	user, ok := dataSession.Values["user"].(*models.User)
	if !ok {
		return nil, fmt.Errorf("Failed to retrieve user from data session")
	}

	return user, nil
}

// SetUser saves the given user object to the database and updates the data session reference
func (controller *Controller) SetUser(w http.ResponseWriter, r *http.Request, user *models.User) (*models.User, error) {
	user, err := controller.database.SaveUser(user)
	if err != nil {
		return nil, err
	}

	dataSession, _ := controller.store.Get(r, "eveposData")

	dataSession.Values["user"] = user

	return user, sessions.Save(r, w)
}
