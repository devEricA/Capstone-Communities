/*
	This controller starts up the application
*/

package controllers

// Packages used in our project
import (
	"github.com/revel/revel"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// Used as a way to control renders
type Startup struct {
	*revel.Controller
}

//By default, Index is the first page that loads in Revel
//We are using this to open up our database and make queries. 
func (c Startup) Index() revel.Result {
   	
	//Error that will display if the database connection fails
	var err error

	//Opening the connection to the database
	// maria_pwd := os.Getenv("MYSQL_PWD")
	maria_pwd := "root"
	db, err = sql.Open("mysql", "root:"+maria_pwd+"@tcp(127.0.0.1:3306)/serverstorage")

	//If database fails to connect, display the error mentioning that the database failed to connect
	if err != nil {
		panic(err.Error())
		c.Flash.Error("Database failed to load")
	}

	//Load the communities nearby
	LoadAllCommunities()
	LoadAllPosts()

	//Ping the database in order to ensure that it is connected
	db.Ping()

	//After connecting to the database, redirect to the Login page
	return c.Redirect(Auth.Login)
}