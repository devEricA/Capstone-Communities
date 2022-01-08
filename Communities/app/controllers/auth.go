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

// Used as a way to control renders
type Auth struct {
	*revel.Controller
}

//Renders the login page
func (c Auth) Login() revel.Result {
	return c.Render()
}

//Function that checks whether or not the inputted login credentials are valid
func (c Auth) LogValidate(LoginUserName string, LoginPassword string) revel.Result{
	//If the login is successful, direct to the Home page
	//Set a flag that the login is successful
	if(DBLogin(LoginUserName, LoginPassword, CurrentSess)){
		LoggedIn = true 
		ActiveUser = LoginUserName
		return c.Redirect(App.Home, ActiveUser)
	}
	//When invalid credentials are inputted, load up an erro message stating that the input is valid. 
	c.Flash.Error("Invalid Username or Password")
	return c.Redirect(Auth.Login)
}

//Attempts to login user. QueryRow throws error if no user + pass combo found
func DBLogin(Username string, Password string, CurrentSess User) bool {
	var err error //Error to deploy
	var HashedPassword string
	//SQL statment to query for the credententials. 
	//sqlStatement := fmt.Sprintf(`SELECT Username, Display_Name, Bio FROM User WHERE Username = ?  AND Password = '?'`, Username, Password)
	UserSearch := db.QueryRow(`SELECT Username, Password FROM User WHERE Username = ?`, Username)
	err = UserSearch.Scan(&CurrentSess.Username, &HashedPassword)
	//Checking for the existence of the user
	if err == sql.ErrNoRows{
		return false
	}else if err == nil{
		if CheckHash(Password, HashedPassword){
			return true;
		}else{
			return false;
		}
	}else{
		panic(err)
	}
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
