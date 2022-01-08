/*
	This controller handles account creations
*/

package controllers

// Packages used in our project
import (
	"github.com/revel/revel"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// Defining the controller used in this file
type AccountCreation struct {
	*revel.Controller
}

//Renders the account creation page
func (c AccountCreation) AccountCreation() revel.Result {
	return c.Render()
}

//Handles Account Creation
//Called whenever a "CreateAccount" form is submitted 
func (c AccountCreation) CreateAccount(NewUserName string, NewPassword string, NewEmail string, NewPasswordConfirmation string) revel.Result{
	//If passwords do not match, redirect to the Account Creation page
	if(NewPassword != NewPasswordConfirmation){
		c.Flash.Error("Passwords do not match.")
		return c.Redirect(AccountCreation.CreateAccount)
	}else if(DBCreateAccount(NewUserName, NewPassword, NewEmail, CurrentSess)){
		// If the creation of the account is successful, redirect to the login page.
		c.Flash.Success("Account Created!")
		return c.Redirect(AccountCreation.TermsOfService)
	}
	//If an error occured when creating the account, return to the account creation page. 
	c.Flash.Error("Error occured when creating the account, email or username already exists.")
	// defer db.Close()
	return c.Redirect(AccountCreation.CreateAccount)
}

//Creates the account from the submission data of AccountCreation.html
func DBCreateAccount(Username string, Password string, Email string, CurrentSess User) bool {
	var err error //Error to deploy

	//Checking for an already existing email
	//sqlStatement := db.Prepare(`SELECT COUNT(Email) FROM User WHERE Email = '?'`) 
	var AccountExist int
	err = db.QueryRow(`SELECT COUNT(Email) FROM User WHERE Email = ?`, Email).Scan(&AccountExist)

	//If an error occured during the check, Panic
	if err != nil {
		panic(err.Error())
	}

	//Returning false if an email already exists, 
	//Thus preventing a creation of an account 
	if AccountExist != 0 {
		return false
	}

	//Checking for an already existing username
	//UsersqlStatement := fmt.Sprintf(`SELECT COUNT(Username) FROM User WHERE Username = '%s'`, Username) 
	var UserExist int
	err = db.QueryRow(`SELECT COUNT(Username) FROM User WHERE Username = ?`, Username).Scan(&UserExist)

	//If an error occured during the check, Panic
	if err != nil {
		panic(err.Error())
	}

	//Returning false if the username already exists, 
	//Thus preventing a creation of an account 
	if UserExist != 0 {
		return false
	}

	//Hashing the password
	HashedPassword, Herr := Hash(Password)
	if Herr != nil{
		panic(Herr.Error())
	}

	//Loading the user into the user database table
	// saveUser := fmt.Sprintf(`INSERT INTO User(Username, Display_Name, Password, Email, Bio) VALUES ('%s', '%s', '%s', '%s', 'None')`, Username, Username, Password, Email)
	_, Loaderr := db.Exec(`INSERT INTO User(Username, Display_Name, Password, Email, Bio) VALUES (?, ?, ?, ?, 'None')`, Username, Username, HashedPassword, Email)

	//Error occured during the load, panic
	if Loaderr != nil {
		panic(Loaderr.Error())
	}

	//Returning true to detail an account creation success
	return true

}

//Renders the terms of service page
func (c AccountCreation) TermsOfService() revel.Result{
	return c.Render()	
}

//Hashing function for passwords
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}