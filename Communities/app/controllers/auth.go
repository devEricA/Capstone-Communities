/*
	This controller handles authenication
*/

package controllers

// Packages used in our project
import (
	"github.com/revel/revel"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// Defining the controller used in this file
type Auth struct {
	*revel.Controller
}

//Renders the login page
func (c Auth) Login() revel.Result {
	return c.Render()
}

// Grabs user data
// Taken from the Hotels example of Revel, slightly modified to work with this instance
func (c Auth) getUser(username string) (user *models.User) {
	user = &models.User{}
	_,  err := c.Session.GetInto("fulluser", user, false)
	if user.Username == username {
		return user
	}
	
	//Grabbing the user
	UserSearch := db.QueryRow(`SELECT Username, Password FROM User WHERE Username = ?`, Username)
	err = c.Txn.SelectOne(user, UserSearch)
	if err != nil {
		if err != sql.ErrNoRows {
			//c.Txn.Select(user, c.Db.SqlStatementBuilder.Select("*").From("User").Limit(1))
			count := db.QueryRow(`SELECT COUNT(*) FROM User `) 
			//count, _ := c.Txn.SelectInt(c.Db.SqlStatementBuilder.Select("count(*)").From("User"))
			c.Log.Error("Failed to find user", "user", username, "error",err, "count", count)
		}
		return nil
	}
	c.Session["fulluser"] = user
	return
}

//Attempts to login user. QueryRow throws error if no user + pass combo found
func (c Auth) DBLogin(Username string, Password string, CurrentSess User) bool {
	user := c.getUser(Username)
	if err == sql.ErrNoRows{
		return false
	}else if err == nil{
		if CheckHash(Password, HashedPassword){
			c.Session["user"] = username
			return true;
		}else{
			return false;
		}
	}else{
		panic(err)
	}
}

//Function that checks whether or not the inputted login credentials are valid
func (c Auth) LogValidate(LoginUserName string, LoginPassword string) revel.Result{
	//If the login is successful, direct to the Home page
	//Set a flag that the login is successful
	if(DBLogin(LoginUserName, LoginPassword, CurrentSess)){
		LoggedIn = true 
		return c.Redirect(App.Home)
	}
	//When invalid credentials are inputted, load up an erro message stating that the input is valid. 
	c.Flash.Error("Invalid Username or Password")
	return c.Redirect(Auth.Login)
}

//Checking for hash matches
func CheckHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

//Renders account recovery
func (c Auth) AccRecovery() revel.Result {
	return c.Render()
}
