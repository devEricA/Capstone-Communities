package controllers

// Packages used in our project
import (
	"github.com/revel/revel"
	"image/color"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"fmt"
	"bufio"
	"os"
   "time"
   "log"
   "strconv"
)

// Used as a way to control renders
type App struct {
	*revel.Controller
}

type User struct {
	Username     string
	Display_Name string
	Bio          string
}

// Global Variables
var CurrentSess User                  //User info
var LoggedIn bool                     //Whether or not the user is logged in
var ActiveUser string				  //Current user that is using the application
var db *sql.DB
var logger *log.Logger
// var dbAsHtml *sql.DB

const (
	addr     = "Lubbock, TX"
	lat, lng = 33.563521, -101.879336
)

// performance logging helpers
func startPerfMeasure() time.Time {
   return time.Now()
}

func finishPerfMeasure(start time.Time, name string) {
   duration := time.Since(start)
   logger.Println(name + " execution time: " + strconv.Itoa(int(duration.Milliseconds())))
}

//By default, Index is the first page that loads in Revel
//We are using this to open up our database and make queries. 
func (c App) Index() revel.Result {
   //Setup logging 
   performanceLogfile, f_err := os.OpenFile("perf.log",
      os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
   if f_err != nil {
      log.Println(f_err)
   }
   defer performanceLogfile.Close()
   logger = log.New(performanceLogfile, "[PERFORMANCE] ", log.LstdFlags)

	//Error that will display if the database connection fails
	var err error

	//Opening the connection to the database
	//This assumes that the database username and password is root
   maria_pwd := os.Getenv("MYSQL_PWD")
   db, err = sql.Open("mysql", "root:"+maria_pwd+"@tcp(127.0.0.1:3306)/serverstorage")

	//If database fails to connect, display the error mentioning that the database failed to connect
	if err != nil {
		panic(err.Error())
		c.Flash.Error("Database failed to load")
	}

	//Load the communities nearby
	LoadAllCommunities()
	LoadAllPosts()

	// TODO: Find a way to take sql input and output in HTML
	// dbAsHtml, err = sql.Open("mysql", )

	//Ping the database in order to ensure that it is connected
	db.Ping()

	//After connecting to the database, redirect to the Login page
	return c.Redirect(App.Login)
}

//Renders the login page
func (c App) Login() revel.Result {
	return c.Render()
}


//Handles Account Creation
//Called whenever a "CreateAccount" form is submitted 
func (c App) CreateAccount(NewUserName string, NewPassword string, NewEmail string, NewPasswordConfirmation string) revel.Result{
	//If passwords do not match, redirect to the Account Creation page
	if(NewPassword != NewPasswordConfirmation){
		c.Flash.Error("Passwords do not match.")
		return c.Redirect(App.AccountCreation)
	}else if(DBCreateAccount(NewUserName, NewPassword, NewEmail, CurrentSess)){
		// If the creation of the account is successful, redirect to the login page.
		c.Flash.Success("Account Created! You may login now. ")
		return c.Redirect(App.Login)
	}
	//If an error occured when creating the account, return to the account creation page. 
	c.Flash.Error("Error occured when creating the account, email or username already exists.")
	// defer db.Close()
	return c.Redirect(App.AccountCreation)
}

//Function that checks whether or not the inputted login credentials are valid
func (c App) LogValidate(LoginUserName string, LoginPassword string) revel.Result{
	//If the login is successful, direct to the Home page
	//Set a flag that the login is successful
	if(DBLogin(LoginUserName, LoginPassword, CurrentSess)){
		LoggedIn = true 
		ActiveUser = LoginUserName
		return c.Redirect(App.Home, ActiveUser)
	}
	//When invalid credentials are inputted, load up an erro message stating that the input is valid. 
	c.Flash.Error("Invalid Username or Password")
	return c.Redirect(App.Login)
}

//Home Page
func (c App) Home(LoginUserName string) revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	LoginUserName = ActiveUser
	//Create the map for the user to explore
	createMap(lat, lng)

	//Load the communities nearby
	LoadAllCommunities()
	LoadAllPosts()

	//TODO: Render user communities, latest posts, and communities on the map
	return c.Render(LoginUserName)
}
 
//Renders the account creation page
func (c App) AccountCreation() revel.Result {
	return c.Render()
}

//Renders account recovery
func (c App) AccRecovery() revel.Result {
	return c.Render()
}

//Renders Profilepage
func (c App) Profile(CurrentUser string) revel.Result {
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	return c.Render(ActiveUser)
}

//Function for updating the user name
//Called whenever the "UpdateUserName" form is submitted in Profile.html
func (c App) UpdateUserName(NewUserName string) revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	var err error //General Error
	var UserNameAlreadyExists int //Checks for whether or not the username exists already

	//SQL statement to check whether or not the username exists. 
	checkStatement := fmt.Sprintf(`SELECT COUNT(Username) FROM User WHERE Username = '%s'`, NewUserName)
	err = db.QueryRow(checkStatement).Scan(&UserNameAlreadyExists)
	if err != nil {
		//If something went wrong with the query, panic
		panic(err.Error())
	}

	//If the username already exists, flash an error, then redirect back to the Profile page 
	if UserNameAlreadyExists != 0 {
		c.Flash.Error("That username already exists")
		return c.Redirect(App.Profile)
	}
	//Query to update the username
	updateQuery := fmt.Sprintf(`UPDATE User SET Username = '%s', Display_Name = '%s' WHERE Username='%s'`, NewUserName, NewUserName, ActiveUser)
	_, Updateerr := db.Exec(updateQuery)

	//If something went wrong during the render, panic
	if Updateerr != nil {
		panic(Updateerr.Error())
	}

	//If user update is successful, update the active user to the new username, 
	//then display a success message and redirect to the profile. 
	ActiveUser = NewUserName
	c.Flash.Success("Username has been updated!")
	return c.Redirect(App.Profile)

}


//Function for updating the password
//Called whenever the "UpdatePassword" form is submitted in Profile.html
func (c App) UpdatePassword(NewPassword string, NewPasswordConfirm string) revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	//If passwords do not match, redirect back to the profile
	//and display an error message
	if(NewPassword != NewPasswordConfirm){
		c.Flash.Error("Passwords do not match")
		return c.Redirect(App.Profile)
	}

	//Query to update to password for the user
	updateQuery := fmt.Sprintf(`UPDATE User SET Password = '%s' WHERE Username = '%s'`, NewPassword, ActiveUser)
	_, Updateerr := db.Exec(updateQuery)

	//If something went wrong with the password update, panic
	if Updateerr != nil {
		panic(Updateerr.Error())
	}

	//Display a Success message detailing that the password has been updated
	//Then redirect to the profile page
	c.Flash.Success("Password has been updated!")
	return c.Redirect(App.Profile)

}

//Renders the community creation page
func (c App) CreateCommunity() revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	return c.Render()
}

//Function that constructs the community
//Called whenever a "ConstructCommunity" form is submitted in CreateCommunity.html
func (c App) ConstructCommunity(NewCommunityName string, CommunityDescription string) revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	var err error	// error that returns during the query
	var communityAlreadyExists int // Checking for the existence of community, don't want duplicates due to confusion
	var numberOfCommunities int	// keeps track of the number of communities

	CommunityDuplicateCheck :=fmt.Sprintf(`SELECT COUNT(NAME) FROM Communities WHERE NAME = '%s'`, NewCommunityName)
	err = db.QueryRow(CommunityDuplicateCheck).Scan(&communityAlreadyExists)
	//Error occured during the check, panic
	if err != nil {
		panic(err.Error())
	}
	if communityAlreadyExists != 0{
		c.Flash.Error("That Community already exists")
		return c.Redirect(App.CreateCommunity)
	}


	//Checks for number of communities
	CommunityQuery := `SELECT COUNT(Community_ID) FROM Communities`
	err = db.QueryRow(CommunityQuery).Scan(&numberOfCommunities)
	//Error occured during the check, panic
	if err != nil {
		panic(err.Error())
	}

	//Adding the inputted community
	addCommunityQuery := fmt.Sprintf(`INSERT INTO Communities(Community_ID, Name, Description, City) VALUES (%d, '%s', '%s', '%s')`, numberOfCommunities, NewCommunityName, CommunityDescription, addr)
	_, Loaderr := db.Exec(addCommunityQuery)

	//if an error occured with the community addition, panic
	if Loaderr != nil {
		panic(Loaderr.Error())
	}
	
	//Refresh the Communities tab to pick up the newly registered community
	LoadAllCommunities()

	//Display a success message detailing the community creation. 
	c.Flash.Success("New Community Created!")
	return c.Redirect(App.Home)

}

func (c App) ConstructPost(PostTitle string, PostContent string) revel.Result{
	var err error
	var TitleExists int
	var DescriptionExists int
	var numberOfPosts int

	TitleQuery := fmt.Sprintf(`SELECT COUNT(Title) FROM Posts WHERE Title = '%s'`, PostTitle)
	err = db.QueryRow(TitleQuery).Scan(&TitleExists)
	if err != nil{
		panic(err.Error())
	}

	if TitleExists !=0{
		c.Flash.Error("A post with that title already exists")
		return c.Redirect(App.NewPost)
	}

	DescriptionQuery := fmt.Sprintf(`SELECT COUNT(Text) FROM Posts WHERE Text = '%s'`)
	err = db.QueryRow(DescriptionQuery).Scan(&DescriptionExists)
	if err != nil{
		panic(err.Error())
	}

	if DescriptionExists != 0{
		c.Flash.Error("A post with the description already exists")
		return c.Redirect(App.NewPost)
	}

	PostCountQuery := fmt.Sprintf(`SELECT COUNT(Post_ID) FROM Posts`)
	err = db.QueryRow(PostCountQuery).Scan(&numberOfPosts)
	if err != nil{
		panic(err.Error())
	}

	LoadPostQuery := fmt.Sprintf(`INSERT INTO Posts(Post_ID, Title, Text, Community, Username_FID) VALUES (%d, '%s', '%s', %d, '%s')`, numberOfPosts, PostTitle, PostContent, 0, ActiveUser)
	_, Loaderr := db.Exec(LoadPostQuery)
	if Loaderr != nil{
		panic(Loaderr.Error())
	}

	c.Flash.Success("Post Created!")
	LoadAllPosts()
	return c.Redirect(App.Home)

}

//Renders the Community page
func (c App) Community() revel.Result{
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	//TODO: Render Community Name, Posts, Description, and Events
	return c.Render()
}

//Renders the New Post page
func (c App) NewPost() revel.Result{
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	return c.Render()
}

//Renders the New Event page
func (c App) NewEvent() revel.Result{
	if(!LoggedIn){
		return c.Redirect(App.Login);
	}
	return c.Render()
}

//Creates the map for the home page
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

//Creates the account from the submission data of AccountCreation.html
func DBCreateAccount(Username string, Password string, Email string, CurrentSess User) bool {
	var err error //Error to deploy

	//Checking for an already existing email
	sqlStatement := fmt.Sprintf(`SELECT COUNT(Email) FROM User WHERE Email = '%s'`, Email) 
	var AccountExist int
	err = db.QueryRow(sqlStatement).Scan(&AccountExist)

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
	UsersqlStatement := fmt.Sprintf(`SELECT COUNT(Username) FROM User WHERE Username = '%s'`, Username) 
	var UserExist int
	err = db.QueryRow(UsersqlStatement).Scan(&UserExist)

	//If an error occured during the check, Panic
	if err != nil {
		panic(err.Error())
	}

	//Returning false if the username already exists, 
	//Thus preventing a creation of an account 
	if UserExist != 0 {
		return false
	}

	//Loading the user into the user database table
	saveUser := fmt.Sprintf(`INSERT INTO User(Username, Display_Name, Password, Email, Bio) VALUES ('%s', '%s', '%s', '%s', 'None')`, Username, Username, Password, Email)
	_, Loaderr := db.Exec(saveUser)

	//Error occured during the load, panic
	if Loaderr != nil {
		panic(Loaderr.Error())
	}

	//Returning true to detail an account creation success
	return true

}

//Attempts to login user. QueryRow throws error if no user + pass combo found
func DBLogin(Username string, Password string, CurrentSess User) bool {
	var err error //Error to deploy

	//SQL statment to query for the credententials. 
	sqlStatement := fmt.Sprintf(`SELECT Username, Display_Name, Bio FROM User WHERE Username = '%s'  AND Password = '%s'`, Username, Password)
	UserSearch := db.QueryRow(sqlStatement)
	// UserSearch := db.QueryROw
	//Switching the error and chacking to see whether or not the user exists. 
	switch err = UserSearch.Scan(&CurrentSess.Username, &CurrentSess.Display_Name, &CurrentSess.Bio); err {
	case sql.ErrNoRows://User does not exist case, return false
		return false
	case nil://Err nil, there is indeed a user, return true. 
		return true
	default:
		panic(err)
	}
}


func LoadAllCommunities(){
	path := "app/views/Communities.html"
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err !=nil{
		panic(err)
	}
	// Variable used to separate Community positions
	var topTrack int = 3

	allCommunitiesQuery := `SELECT Name, Description FROM Communities`
	allCommunities, Qerr := db.Query(allCommunitiesQuery)
	if Qerr != nil{
		panic(Qerr)
	}
	// var Name string;
	// for allCommunities.Next(){
	// 	readerr := allCommunities.Scan(&Name)
	// 	if readerr != nil{
	// 		panic(readerr.Error())
	// 	}
	// 	fmt.Print("Community " + Name + " exists.\n")
	// }

	defer file.Close()
	sqlToHtml := bufio.NewWriter(file)
	
	//top = 3%
	// styleForWindows := `
	// <style>
	// .communityWindow{
	// 	width: 94%;
	// 	height: 10%;
	// 	border: 10px solid black;
	// 	position: absolute;
	// 	right: 1%;
	//   }
	// // </style>`
	// _, StyleRenderErr := sqlToHtml.WriteString(styleForWindows)
	// if StyleRenderErr != nil{
	// 	panic(StyleRenderErr.Error())
	// }
	for allCommunities.Next(){
		var Name string
		var Description string
		
		readerr := allCommunities.Scan(&Name, &Description)
		if readerr != nil{
			panic(readerr.Error())
		}
		percentage := fmt.Sprintf("style= \"top:%d%%;\"", topTrack)
		
		htmlToRender := fmt.Sprintf(`
			<div class = "communityWindow" %s>
				<b>%s</b><br/>
				%s
				<form action="/Community" method="GET" right="1%%">
					<input type="submit" value="Visit Community" />
				</form>
		    </div>
			`,percentage, Name, Description)
		_, htmlRenderErr := sqlToHtml.WriteString(htmlToRender)
		if htmlRenderErr != nil{
			panic(htmlRenderErr.Error())
		}
		topTrack += 11
	}
	if err := sqlToHtml.Flush(); err != nil {
        panic(err.Error())
    }

}

func LoadAllPosts(){
	path := "app/views/LatestPosts.html"
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err !=nil{
		panic(err)
	}

	var toptrack int = 4

	allPostsQuery := `SELECT Title, Text FROM Posts`
	allPosts, Qerr := db.Query(allPostsQuery)
	if Qerr != nil{
		panic(err)
	}

	defer file.Close()
	sqlToHtml := bufio.NewWriter(file)

	for allPosts.Next(){
		var Title string
		var Text string
		readerr := allPosts.Scan(&Title, &Text)
		if readerr != nil{
			panic(readerr.Error())
		}
		percentage := fmt.Sprintf("style= \"top:%d%%;\"", toptrack)

		htmlToRender := fmt.Sprintf(`
		<div class = "postWindow" %s>
			<b> %s </b><br/>
			%s
		</div>
		`, percentage, Title, Text)
		_, htmlRenderErr := sqlToHtml.WriteString(htmlToRender)

		if htmlRenderErr != nil{
			panic(htmlRenderErr.Error())
		}
		toptrack += 11
	}

	if err := sqlToHtml.Flush(); err != nil {
		panic(err.Error())
	}

}

/*
func LoadAllPosts() revel.Result{
	//Load the HTML file here
	path := "app/views/Posts.html"
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err !=nil{
		panic(err)
	}




}
*/
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
