package web

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/morpheusxaut/evepos/database"

	"github.com/dustin/go-humanize"
)

// Templates stores the parsed HTTP templates used by the web app
type Templates struct {
	template *template.Template
	database database.Connection
}

// SetupTemplates parses the HTTP templates from disk and stores them for later usage
func SetupTemplates(db database.Connection) *Templates {
	templates := &Templates{
		database: db,
	}

	templates.template = template.Must(template.New("").Funcs(templates.TemplateFunctions(nil)).ParseGlob("app/templates/*"))

	return templates
}

// ReloadTemplates re-reads the HTTP templates from disk and refreshes the output
func (templates *Templates) ReloadTemplates() {
	templates.template = template.Must(template.New("").Funcs(templates.TemplateFunctions(nil)).ParseGlob("app/templates/*"))
}

// ExecuteTemplates performs all replacement in the HTTP templates and sends the response to the client
func (templates *Templates) ExecuteTemplates(w http.ResponseWriter, r *http.Request, template string, response map[string]interface{}) error {
	return templates.template.Funcs(templates.TemplateFunctions(r)).ExecuteTemplate(w, template, response)
}

// TemplateFunctions prepares a map of functions to be used within templates
func (templates *Templates) TemplateFunctions(r *http.Request) template.FuncMap {
	return template.FuncMap{
		"IsResultNil":                func(r interface{}) bool { return templates.IsResultNil(r) },
		"FormatType":                 func(t int64) string { return templates.FormatType(t) },
		"FormatLocation":             func(m int64) string { return templates.FormatLocation(m) },
		"FormatState":                func(s int64) string { return templates.FormatState(s) },
		"FormatRemainingFuelTime":    func(u int64, q int64) string { return templates.FormatRemainingFuelTime(u, q) },
		"FormatInt64":                func(i int64) string { return templates.FormatInt64(i) },
		"CalculateRemainingFuelTime": func(u int64, q int64) int64 { return templates.CalculateRemainingFuelTime(u, q) },
	}
}

// IsResultNil checks whether the given result/interface is nil
func (templates *Templates) IsResultNil(r interface{}) bool {
	return (r == nil)
}

func (templates *Templates) FormatType(typeID int64) string {
	typeName, err := templates.database.QueryTypeName(typeID)
	if err != nil {
		return strconv.FormatInt(typeID, 10)
	}

	return typeName
}

func (templates *Templates) FormatLocation(moonID int64) string {
	location, err := templates.database.QueryLocationName(moonID)
	if err != nil {
		return strconv.FormatInt(moonID, 10)
	}

	return location
}

func (templates *Templates) FormatState(state int64) string {
	switch state {
	case 0:
		return "Unanchored"
	case 1:
		return "Anchored / Offline"
	case 2:
		return "Onlining"
	case 3:
		return "Reinforced"
	case 4:
		return "Online"
	default:
		return strconv.FormatInt(state, 10)
	}
}

func (templates *Templates) FormatRemainingFuelTime(usage int64, quantity int64) string {
	return humanize.Time(time.Now().Add(time.Hour * time.Duration(quantity/usage)))
}

func (templates *Templates) FormatInt64(i int64) string {
	return humanize.Comma(i)
}

func (templates *Templates) CalculateRemainingFuelTime(usage int64, quantity int64) int64 {
	remainingFuelTime := quantity / usage

	if remainingFuelTime > 0 {
		return remainingFuelTime
	} else {
		return 999999999
	}
}
