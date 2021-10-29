package controllers

import (
	"github.com/revel/revel"
	"image/color"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"fmt"
)

type App struct {
	*revel.Controller
}

type User struct {
	User_ID      int64
	Display_Name string
	Bio          string
}

var CurrentSess User                  //User info
var LoggedIn bool                     //Login success
var NoAcc = "Wrong Email or Password" //Fail String
var LoginSuccess = "Welcome " 
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
	// c.Flash.Error("Can I set these off?")
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
/*
func (c App) LogValidate(LoginUserName string, LoginPassword string) revel.Result{
	if(DBLogin(LoginUserName, LoginPassword, CurrentSess)){
		LoggedIn = true; 
		return c.Redirect(App.Home)
	}
	else{
		c.Flash.Error("Invalid Username or Password")
		return c.Redirect(App.Login)
	}
}



func (c App) CreatePost() revel.Result{

}

func (c App) CreateCommunity() revel.Result{

}
*/

func (c App) Home(LoginUserName string, LoginPassword string) revel.Result{
	c.Validation.Required(LoginUserName).Message("Username is required.")
    c.Validation.MinSize(LoginUserName, 3).Message("Username is not long enough, it must have at least 3 characters.")
	c.Validation.Required(LoginPassword).Message("Password is required.")
	c.Validation.MinSize(LoginPassword, 5).Message("Password must be at least 5 characters.")

    if c.Validation.HasErrors() {
        c.Validation.Keep()
        c.FlashParams()
        return c.Redirect(App.Login)
    }

	createMap(lat, lng)
	//TODO: Render user communities, latest posts, and communities on the map
	return c.Render(LoginUserName, LoginPassword)
}
 
func (c App) AccountCreation() revel.Result {
	return c.Render()
}

func (c App) AccRecovery() revel.Result {
	return c.Render()
}

func (c App) Profile(LoginUserName string, LoginPassword string) revel.Result {
	return c.Render(LoginUserName, LoginPassword)
}

func (c App) CreateCommunity() revel.Result{
	return c.Render()
}

func (c App) Community() revel.Result{
	//TODO: Render Community Name, Posts, Description, and Events
	return c.Render()
}

func (c App) NewPost() revel.Result{
	return c.Render()
}

func (c App) NewEvent() revel.Result{
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

func DBCreateAccount(Username string, Password string, Email string, CurrentSess User) bool {
	var err error

	//UserSearch, err := db.Query("SELECT COUNT(User_Email) FROM User WHERE User_Email = $1 ", &Email)
	sqlStatement := fmt.Sprintf(`SELECT COUNT(Email) FROM User WHERE Email = '%s'`, Email)
	fmt.Printf("SQL statement is :: " + sqlStatement)
	//UserSearch, err := db.Prepare(sqlStatement)
	var AccountExist int
	if err != nil {
		panic(err.Error())
	}

	err = db.QueryRow(sqlStatement).Scan(&AccountExist)
	if err != nil {
		panic(err.Error())
	}

	//UserSearch.Close()

	if AccountExist != 0 {
		return false
	}

	_, err = db.Exec("INSERT INTO User (Username, Password, Email) VALUES (?, ?, ?)", &Username, &Password, &Email)

	if err != nil {
		panic(err.Error())
	}
	return true

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
//Attempts to login user. QueryRow throws error if no user + pass combo found
func DBLogin(Username int, Password string, CurrentSess User) bool {
	var err error

	UserSearch := db.QueryRow("SELECT User_ID, Display_Name, Bio FROM User WHERE User_ID = ? AND Password = ? ", &Username, &Password)
	switch err = UserSearch.Scan(&CurrentSess.User_ID, &CurrentSess.Display_Name, &CurrentSess.Bio); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return true
	default:
		panic(err)
	}
}

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