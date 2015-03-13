package mysql

import (
	"fmt"

	"github.com/morpheusxaut/evepos/misc"
	"github.com/morpheusxaut/evepos/models"

	// Blank import of the MySQL driver to use with sqlx
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/morpheusxaut/eveapi"
)

// DatabaseConnection provides an implementation of the Connection interface using a MySQL database
type DatabaseConnection struct {
	// Config stores the current configuration values being used
	Config *misc.Configuration

	conn *sqlx.DB
}

// Connect tries to establish a connection to the MySQL backend, returning an error if the attempt failed
func (c *DatabaseConnection) Connect() error {
	conn, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true", c.Config.DatabaseUser, c.Config.DatabasePassword, c.Config.DatabaseHost, c.Config.DatabaseSchema))
	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

// RawQuery performs a raw MySQL query and returns a map of interfaces containing the retrieve data. An error is returned if the query failed
func (c *DatabaseConnection) RawQuery(query string, v ...interface{}) ([]map[string]interface{}, error) {
	rows, err := c.conn.Query(query, v...)
	if err != nil {
		return nil, err
	}

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	var results []map[string]interface{}

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)

		resultRow := make(map[string]interface{})

		for i, col := range columns {
			resultRow[col] = values[i]
		}

		results = append(results, resultRow)
	}

	return results, nil
}

func (c *DatabaseConnection) LoadAllAPIKeys() ([]eveapi.Key, error) {
	var apiKeys []eveapi.Key

	err := c.conn.Select(&apiKeys, "SELECT id, vcode FROM apikeys")
	if err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (c *DatabaseConnection) LoadAllUsers() ([]*models.User, error) {
	var users []*models.User

	err := c.conn.Select(&users, "SELECT id, username, password, email, verifiedemail, active FROM users")
	if err != nil {
		return nil, err
	}

	return users, nil
}

// LoadUserFromUsername retrieves the user (and its associated groups and user roles) with the given username from the database, returning an error if the query failed
func (c *DatabaseConnection) LoadUserFromUsername(username string) (*models.User, error) {
	user := &models.User{}

	err := c.conn.Get(user, "SELECT id, username, password, email, verifiedemail, active FROM users WHERE username LIKE ?", username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// LoadPasswordForUser retrieves the password associated with the given username from the MySQL database, returning an error if the query failed
func (c *DatabaseConnection) LoadPasswordForUser(username string) (string, error) {
	row := c.conn.QueryRowx("SELECT password FROM users WHERE username LIKE ?", username)

	var password string

	err := row.Scan(&password)
	if err != nil {
		return "", err
	}

	return password, nil
}

func (c *DatabaseConnection) QueryLocationName(moonID int64) (string, error) {
	var locationName string

	err := c.conn.Get(&locationName, "SELECT itemName FROM mapDenormalize WHERE itemID = ?", moonID)
	if err != nil {
		return "", err
	}

	return locationName, nil
}

func (c *DatabaseConnection) QueryTypeName(typeID int64) (string, error) {
	var typeName string

	err := c.conn.Get(&typeName, "SELECT typeName FROM invTypes WHERE typeID = ?", typeID)
	if err != nil {
		return "", err
	}

	return typeName, nil
}

func (c *DatabaseConnection) QueryFuelUsage(posTypeID int64, fuelTypeID int64) (int64, error) {
	var usage int64

	err := c.conn.Get(&usage, "SELECT quantity FROM invControlTowerResources WHERE controlTowerTypeID = ? AND resourceTypeID = ?", posTypeID, fuelTypeID)
	if err != nil {
		return -1, err
	}

	return usage, nil
}

// SaveUser saves a user to the MySQL database, returning the updated model or an error if the query failed
func (c *DatabaseConnection) SaveUser(user *models.User) (*models.User, error) {
	if user.ID > 0 {
		_, err := c.conn.Exec("UPDATE users SET username=?, password=?, email=?, verifiedemail=?, active=? WHERE id=?", user.Username, user.Password, user.Email, user.VerifiedEmail, user.Active, user.ID)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err := c.conn.Exec("INSERT INTO users(username, password, email, verifiedemail, active) VALUES(?, ?, ?, ?, ?)", user.Username, user.Password, user.Email, user.VerifiedEmail, user.Active)
		if err != nil {
			return nil, err
		}

		lastInsertedID, err := resp.LastInsertId()
		if err != nil {
			return nil, err
		}

		user.ID = lastInsertedID
	}

	return user, nil
}

// SaveLoginAttempt saves a login attempt to the MySQL database, returning an error if the query failed
func (c *DatabaseConnection) SaveLoginAttempt(loginAttempt *models.LoginAttempt) error {
	_, err := c.conn.Exec("INSERT INTO loginattempts(username, remoteaddr, useragent, successful) VALUES(?, ?, ?, ?)", loginAttempt.Username, loginAttempt.RemoteAddr, loginAttempt.UserAgent, loginAttempt.Successful)
	if err != nil {
		return err
	}

	return nil
}
