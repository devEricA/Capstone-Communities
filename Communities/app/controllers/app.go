package controllers

import (
	"github.com/revel/revel"
	"image/color"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"github.com/codingsince1985/geo-golang"
	"github.com/codingsince1985/geo-golang/openstreetmap"
	"fmt"
)

type App struct {
	*revel.Controller
}

type User struct {
	Username     string
	Display_Name string
	Bio          string
}

var CurrentSess User                  //User info
var LoggedIn bool                     //Login success
var ActiveUser string
var db *sql.DB
var dberr error



const (
	addr     = "Lubbock, TX"
	lat, lng = 33.563521, -101.879336
)

func (c App) Index() revel.Result {
	//Consider this as our "Starter Function Area"
	//Instantiate the database here!
	var err error

	db, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/serverstorage")

	if err != nil {
		panic(err.Error())
		c.Flash.Error("Database failed to load")
	}

	db.Ping()
	return c.Redirect(App.Login)
}

func (c App) Login() revel.Result {
	return c.Render()
}


func (c App) CreateAccount(NewUserName string, NewPassword string, NewEmail string, NewPasswordConfirmation string) revel.Result{
	if(NewPassword != NewPasswordConfirmation){
		c.Flash.Error("Passwords do not match.")
		return c.Redirect(App.AccountCreation)
	}else if(DBCreateAccount(NewUserName, NewPassword, NewEmail, CurrentSess)){
		c.Flash.Success("Account Created! You may login now. ")
		return c.Redirect(App.Login)
	}
	c.Flash.Error("Error occured when creating the account.")
	defer db.Close()
	return c.Redirect(App.AccountCreation)
}

func (c App) LogValidate(LoginUserName string, LoginPassword string) revel.Result{
	if(DBLogin(LoginUserName, LoginPassword, CurrentSess)){
		LoggedIn = true 
		ActiveUser = LoginUserName
		// boolDebug := fmt.Sprintf("Logged in? %t", LoggedIn)
		// fmt.Printf(boolDebug)
		// fmt.Printf("Active User is " + ActiveUser)
		return c.Redirect(App.Home, ActiveUser)
	}
	c.Flash.Error("Invalid Username or Password")
	return c.Redirect(App.Login)
}

/*

func (c App) CreatePost() revel.Result{

}

func (c App) CreateCommunity() revel.Result{

}
*/

func (c App) Home(CurrentUser string) revel.Result{
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	// boolDebug := fmt.Sprintf("Logged in? %t", LoggedIn)
	// fmt.Printf(boolDebug)
	// fmt.Printf("Active User is " + ActiveUser)
	createMap(lat, lng)
	//TODO: Render user communities, latest posts, and communities on the map
	return c.Render(ActiveUser)
}
 
func (c App) AccountCreation() revel.Result {
	return c.Render()
}

func (c App) AccRecovery() revel.Result {
	return c.Render()
}

func (c App) Profile(CurrentUser string) revel.Result {
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	
	return c.Render(ActiveUser)
}

func (c App) UpdateUserName(NewUserName string) revel.Result{
	var err error 
	var UserNameAlreadyExists int
	checkStatement := fmt.Sprintf(`SELECT COUNT(Username) FROM User WHERE Username = '%s'`, NewUserName)
	err = db.QueryRow(checkStatement).Scan(&UserNameAlreadyExists)
	if err != nil {
		panic(err.Error())
	}

	if UserNameAlreadyExists != 0 {
		c.Flash.Error("That username already exists")
		return c.Redirect(App.Profile)
	}
	updateQuery := fmt.Sprintf(`UPDATE User SET Username = '%s', Display_Name = '%s' WHERE Username='%s'`, NewUserName, NewUserName, ActiveUser)
	_, Updateerr := db.Exec(updateQuery)

	if Updateerr != nil {
		panic(Updateerr.Error())
	}
	ActiveUser = NewUserName
	c.Flash.Success("Username has been updated!")
	return c.Redirect(App.Profile)

}

func (c App) UpdatePassword(NewPassword string, NewPasswordConfirm string) revel.Result{
	if(NewPassword != NewPasswordConfirm){
		c.Flash.Error("Passwords do not match")
		return c.Redirect(App.Profile)
	}
	updateQuery := fmt.Sprintf(`UPDATE User SET Password = '%s' WHERE Username = '%s'`, NewPassword, ActiveUser)
	_, Updateerr := db.Exec(updateQuery)

	if Updateerr != nil {
		panic(Updateerr.Error())
	}
	c.Flash.Success("Password has been updated!")
	return c.Redirect(App.Profile)

}

func (c App) CreateCommunity() revel.Result{
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	return c.Render()
}

func (c App) ConstructCommunity(NewCommunityName string, CommunityDescription string) revel.Result{
	var err error
	var numberOfCommunities int
	CommunityQuery := `SELECT COUNT(Community_ID) FROM Communities`
	err = db.QueryRow(CommunityQuery).Scan(&numberOfCommunities)
	if err != nil {
		panic(err.Error())
	}
	addCommunityQuery := fmt.Sprintf(`INSERT INTO Communities(Community_ID, Name, Description, City) VALUES (%d, '%s', '%s', '%s')`, numberOfCommunities, NewCommunityName, CommunityDescription, addr)
	_, Loaderr := db.Exec(addCommunityQuery)

	if Loaderr != nil {
		panic(Loaderr.Error())
	}
	c.Flash.Success("New Community Created!")
	return c.Redirect(App.Home)

}

func (c App) Community() revel.Result{
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	//TODO: Render Community Name, Posts, Description, and Events
	return c.Render()
}

func (c App) NewPost() revel.Result{
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	return c.Render()
}

func (c App) NewEvent() revel.Result{
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	return c.Render()
}


func createMap(lat float64, lng float64) {
	ctx := sm.NewContext() //Creates a new map 'context' 
	ctx.SetSize(1080, 1080) //Sets the size of the map in pixels(?) 

	//This should change depending on the city, for now, default to 11.
	ctx.SetZoom(11) 




	ctx.AddObject( 
		sm.NewCircle( //Draw the circle on the map
			s2.LatLngFromDegrees(lat, lng), //Position by latitude / longitude
			color.RGBA{17, 104, 151, 0xff}, //Outline color
			color.RGBA{150, 212, 231, 1.0}, //Fill color
			20000.0, //Radius
			1, //Weight	
		),
	)

	ctx.AddObject(
		sm.NewMarker( //Draw a marker on the map
			s2.LatLngFromDegrees(lat, lng), //Position by latitude / longitude 
			color.RGBA{0xff, 0, 0, 0xff}, // Color of marker 
			16.0, //Size of the marker
		),
	)	


	img, err := ctx.Render() //Converts the map data to an image on OSM
	if err != nil {
		panic(err)
	}

	if err := gg.SavePNG("public/img/my-map.png", img); err != nil { //Downloads the image from Open Street Map
		panic(err)
	}
}

func reverseGeo(geocoder geo.Geocoder, lat float64, lng float64) string {

	//Example: 
	// address.City 		    // Seattle
	// address.State 	        // WA
	// address.CountryCode      // US 
	// There are more available fields, but we only use these.

	address, _ := geocoder.ReverseGeocode(lat, lng)

	//Concatenate the string, store it, then return the string
	var CityKey = address.CountryCode + address.State + address.City
	
	if(len(CityKey) > 50){
		CityKey = CityKey[0:50]
	}

	return CityKey
}

func DBCreateAccount(Username string, Password string, Email string, CurrentSess User) bool {
	var err error
	sqlStatement := fmt.Sprintf(`SELECT COUNT(Email) FROM User WHERE Email = '%s'`, Email) 
	var AccountExist int
	if err != nil {
		panic(err.Error())
	}

	//Query is deployed here
	err = db.QueryRow(sqlStatement).Scan(&AccountExist)
	if err != nil {
		panic(err.Error())
	}

	if AccountExist != 0 {
		return false
	}

	saveUser := fmt.Sprintf(`INSERT INTO User(Username, Display_Name, Password, Email, Bio) VALUES ('%s', '%s', '%s', '%s', 'None')`, Username, Username, Password, Email)
	// fmt.Printf("Executed Statement is :: " + saveUser)
	_, Loaderr := db.Exec(saveUser)

	if Loaderr != nil {
		panic(Loaderr.Error())
	}
	return true

}

//Attempts to login user. QueryRow throws error if no user + pass combo found
func DBLogin(Username string, Password string, CurrentSess User) bool {
	var err error

	sqlStatement := fmt.Sprintf(`SELECT Username, Display_Name, Bio FROM User WHERE Username = '%s'  AND Password = '%s'`, Username, Password)
	UserSearch := db.QueryRow(sqlStatement)
	switch err = UserSearch.Scan(&CurrentSess.Username, &CurrentSess.Display_Name, &CurrentSess.Bio); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return true
	default:
		panic(err)
	}
}
// func InitDB() {
// 	var err error

// 	db, err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/serverstorage")

// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	db.Ping()
// }

/*


//Searches for an email associated with an account and then returns whether one was found or not
func DBAccRecovery(Email string) string {
	var err error
	var AccountExist int
	var NoAcc = "Account Not Found"
	var YesAcc = "Recovery email sent!"

	UserSearch, err := db.Query("SELECT COUNT(User_Email) FROM User WHERE User_Email = ? ", &Email)

	if err != nil {
		panic(err.Error())
	}
	UserSearch.Scan(&AccountExist)
	UserSearch.Close()

	if AccountExist == 0 {
		return NoAcc
	} else {
		return YesAcc
	}
}

//Searches for an account already reigstered with an email and then will attempt to create account or return an account already exists

*/