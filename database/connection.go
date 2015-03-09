package database

import (
	"fmt"

	"github.com/morpheusxaut/evepos/database/mysql"
	"github.com/morpheusxaut/evepos/misc"
	"github.com/morpheusxaut/evepos/models"
)

// Connection provides an interface for communicating with a database backend in order to retrieve and persist the needed information
type Connection interface {
	// Connect tries to establish a connection to the database backend, returning an error if the attempt failed
	Connect() error

	// RawQuery performs a raw database query and returns a map of interfaces containing the retrieve data. An error is returned if the query failed
	RawQuery(query string, v ...interface{}) ([]map[string]interface{}, error)

	// LoadUserFromUsername retrieves the user with the given username from the database, returning an error if the query failed
	LoadUserFromUsername(username string) (*models.User, error)

	// LoadPasswordForUser retrieves the password associated with the given username from the database, returning an error if the query failed
	LoadPasswordForUser(username string) (string, error)

	// SaveUser saves a user to the database, returning the updated model or an error if the query failed
	SaveUser(user *models.User) (*models.User, error)
	// SaveLoginAttempt saves a login attempt to the database, returning an error if the query failed
	SaveLoginAttempt(loginAttempt *models.LoginAttempt) error
}

// SetupDatabase parses the database type set in the configuration and returns an appropriate database implementation or an error if the type is unknown
func SetupDatabase(conf *misc.Configuration) (Connection, error) {
	var database Connection

	switch Type(conf.DatabaseType) {
	case TypeMySQL:
		database = &mysql.DatabaseConnection{
			Config: conf,
		}
		break
	default:
		return nil, fmt.Errorf("Unknown type #%d", conf.DatabaseType)
	}

	return database, nil
}
