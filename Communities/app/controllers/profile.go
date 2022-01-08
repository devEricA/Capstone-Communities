/*
	This controller handles profile updates
*/

package controllers

// Packages used in our project
import (
	"github.com/revel/revel"
	_ "github.com/go-sql-driver/mysql"
)

// Used as a way to control renders
type Profile struct {
	*revel.Controller
}

//Renders Profilepage
func (c Profile) Profile(CurrentUser string) revel.Result {
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	return c.Render(ActiveUser)
}

//Function for updating the user name
//Called whenever the "UpdateUserName" form is submitted in Profile.html
func (c Profile) UpdateUserName(NewUserName string) revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	var err error //General Error
	var UserNameAlreadyExists int //Checks for whether or not the username exists already

	//SQL statement to check whether or not the username exists. 
	//checkStatement := fmt.Sprintf(`SELECT COUNT(Username) FROM User WHERE Username = '%s'`, NewUserName)
	err = db.QueryRow(`SELECT COUNT(Username) FROM User WHERE Username = ?`, NewUserName).Scan(&UserNameAlreadyExists)
	if err != nil {
		//If something went wrong with the query, panic
		panic(err.Error())
	}

	//If the username already exists, flash an error, then redirect back to the Profile page 
	if UserNameAlreadyExists != 0 {
		c.Flash.Error("That username already exists")
		return c.Redirect(Profile.Profile)
	}
	//Query to update the username
	// updateQuery := fmt.Sprintf(`UPDATE User SET Username = '%s', Display_Name = '%s' WHERE Username='%s'`, NewUserName, NewUserName, ActiveUser)
	_, Updateerr := db.Exec(`UPDATE User SET Username = ?, Display_Name = ? WHERE Username=?`, NewUserName, NewUserName, ActiveUser)

	//If something went wrong during the render, panic
	if Updateerr != nil {
		panic(Updateerr.Error())
	}

	//If user update is successful, update the active user to the new username, 
	//then display a success message and redirect to the profile. 
	ActiveUser = NewUserName
	c.Flash.Success("Username has been updated!")
	return c.Redirect(Profile.Profile)

}


//Function for updating the password
//Called whenever the "UpdatePassword" form is submitted in Profile.html
func (c Profile) UpdatePassword(NewPassword string, NewPasswordConfirm string) revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	//If passwords do not match, redirect back to the profile
	//and display an error message
	if(NewPassword != NewPasswordConfirm){
		c.Flash.Error("Passwords do not match")
		return c.Redirect(Profile.Profile)
	}
	NewHashedPassword, Herr := Hash(NewPassword)
	if Herr != nil{
		panic(Herr.Error())
	}
	//Query to update to password for the user
	//updateQuery := fmt.Sprintf(`UPDATE User SET Password = '%s' WHERE Username = '%s'`, NewPassword, ActiveUser)
	_, Updateerr := db.Exec(`UPDATE User SET Password = ? WHERE Username = ?`, NewHashedPassword, ActiveUser)

	//If something went wrong with the password update, panic
	if Updateerr != nil {
		panic(Updateerr.Error())
	}

	//Display a Success message detailing that the password has been updated
	//Then redirect to the profile page
	c.Flash.Success("Password has been updated!")
	return c.Redirect(Profile.Profile)

}