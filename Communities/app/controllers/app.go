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
	"strconv"
   //"time"
   //"log"
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
var ActiveCommunity string			  //Current community that the user is looking at
var ActiveComDescription string 	  //Description of the community that the user is looking at
var db *sql.DB						  //Database Pointer

const (
	addr     = "Lubbock, TX"
	lat, lng = "33.563521", "-101.879336"
)

//Home Page
func (c App) Home(LoginUserName string) revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	LoginUserName = ActiveUser

	LatValue, Laerr := strconv.ParseFloat(lat,64)
	if Laerr != nil{
		panic(Laerr.Error())
	}
	LongValue, Loerr := strconv.ParseFloat(lng,64)
	if Loerr != nil{
		panic(Loerr.Error())
	}
	
	//Create the map for the user to explore
	createMap(LatValue, LongValue)
	//Load the communities nearby
	LoadAllCommunities()
	LoadAllPosts()

	//TODO: Render user communities, latest posts, and communities on the map
	return c.Render(LoginUserName)
}

//Renders the community creation page
func (c App) CreateCommunity() revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	return c.Render()
}

//Function that constructs the community
//Called whenever a "ConstructCommunity" form is submitted in CreateCommunity.html
func (c App) ConstructCommunity(NewCommunityName string, CommunityDescription string) revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	var err error	// error that returns during the query
	var communityAlreadyExists int // Checking for the existence of community, don't want duplicates due to confusion
	var numberOfCommunities int	// keeps track of the number of communities

	// CommunityDuplicateCheck :=fmt.Sprintf(`SELECT COUNT(NAME) FROM Communities WHERE NAME = '%s'`, NewCommunityName)
	err = db.QueryRow(`SELECT COUNT(NAME) FROM Communities WHERE NAME = ?`, NewCommunityName).Scan(&communityAlreadyExists)
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
	//addCommunityQuery := fmt.Sprintf(`INSERT INTO Communities(Community_ID, Name, Description, City) VALUES (%d, '%s', '%s', '%s')`, numberOfCommunities, NewCommunityName, CommunityDescription, addr)
	_, Loaderr := db.Exec(`INSERT INTO Communities(Community_ID, Name, Description, City, Longitude, Latitude) VALUES (?, ?, ?, ?, ?, ?)`, numberOfCommunities, NewCommunityName, CommunityDescription, addr, lng, lat)

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

//Function for handling Post Construction
func (c App) ConstructPost(PostTitle string, PostContent string, CurrentCommunity string) revel.Result{
	// Error, Title, Description, and Number of Post Variables
	var err error
	var TitleExists int
	var DescriptionExists int
	var numberOfPosts int
	
	CurrentCommunity = ActiveCommunity

	//Querying to check if there is an existing title: Part of guarding from reposts
	err = db.QueryRow(`SELECT COUNT(Title) FROM Posts WHERE Title = ?`, PostTitle).Scan(&TitleExists)
	if err != nil{
		panic(err.Error())
	}

	//If there is already a post with the title, display an error message, and redirect to form. 
	if TitleExists !=0{
		c.Flash.Error("A post with that title already exists")
		return c.Redirect(App.NewPost)
	}

	//Querying to check if there is a post with the same description: Part of guarding from reposts 
	err = db.QueryRow(`SELECT COUNT(Text) FROM Posts WHERE Text = ?`, PostContent).Scan(&DescriptionExists)
	if err != nil{
		panic(err.Error())
	}

	//If there is already a post with the description, display an error message, and redirect to form
	if DescriptionExists != 0{
		c.Flash.Error("A post with the description already exists")
		return c.Redirect(App.NewPost)
	}

	//Counting how many posts are in the database, and storing that count 
	PostCountQuery := fmt.Sprintf(`SELECT COUNT(Post_ID) FROM Posts`)
	err = db.QueryRow(PostCountQuery).Scan(&numberOfPosts)
	if err != nil{
		panic(err.Error())
	}

	//Grabbing the active community
	var activeCommunityID int
	Qerr :=db.QueryRow(`SELECT Community_ID FROM Communities WHERE Name = ?`, ActiveCommunity).Scan(&activeCommunityID)
	if Qerr != nil{
		panic(Qerr.Error())
	}

	//Inserting the post into the post database.
	_, Loaderr := db.Exec(`INSERT INTO Posts(Post_ID, Title, Text, Community, Username_FID) VALUES (?, ?, ?, ?, ?)`, numberOfPosts, PostTitle, PostContent, activeCommunityID, ActiveUser)
	if Loaderr != nil{
		panic(Loaderr.Error())
	}

	LoadAllPosts()
	
	//Creation of a new post, and redirecting to the homepage
	c.Flash.Success("Post Created!")
	return c.LoadAssociatedData(ActiveCommunity, ActiveComDescription)

}

//Renders the Community page
func (c App) Community(CurrentCommunity string, CurrentCommunityDescription string) revel.Result{
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	CurrentCommunity = ActiveCommunity
	CurrentCommunityDescription = ActiveComDescription
	return c.Render(CurrentCommunity, CurrentCommunityDescription)
}

//Function that populates the Community Page
func (c App) LoadAssociatedData(CurrentCommunity string, CurrentCommunityDescription string)revel.Result{
	//Opening the template that is responsible for holding the HTML for all post entries. 
	ActiveCommunity = CurrentCommunity
	ActiveComDescription = CurrentCommunityDescription
	path := "app/views/CommunityPosts.html"
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err !=nil{
		panic(err.Error())
	}

	//Clears out any existing entries in the file.
	//This is important because the database will be experiencing updates (adding & deleting)
	cleanErr := os.Truncate(path, 0)
	if cleanErr != nil{
		panic(cleanErr.Error())
	}

	//Grabbing all of the posts
	allPosts, Qerr := db.Query(`SELECT Title, Text FROM Posts WHERE Community = (SELECT Community_ID FROM Communities WHERE Name = ?) ORDER BY Post_ID DESC`, CurrentCommunity)
	if Qerr != nil{
		panic(Qerr.Error())
	}

	//Closing the file, opening a new writer
	defer file.Close()
	sqlToHtml := bufio.NewWriter(file)

	//For each entry in the post table
	//Grab the name and description of the post
	//Then spit HTML into the template in order to render the post entry
	for allPosts.Next(){
		var Title string
		var Text string
		readerr := allPosts.Scan(&Title, &Text)
		if readerr != nil{
			panic(readerr.Error())
		}	

		htmlToRender := fmt.Sprintf(`
		<div class = "postWindow">
			<b> %s </b><br/>
			%s
		</div>
		`,Title, Text)
		_, htmlRenderErr := sqlToHtml.WriteString(htmlToRender)

		if htmlRenderErr != nil{
			panic(htmlRenderErr.Error())
		}
	}
	if err := sqlToHtml.Flush(); err != nil {
		panic(err.Error())
	}

	//Opening the Events file
	EventPath := "app/views/CommunityEvents.html"
	Eventfile, EPerr := os.OpenFile(EventPath, os.O_RDWR|os.O_CREATE, 0755)
	if EPerr !=nil{
		panic(EPerr.Error())
	}

	//Clears out any existing entries in the file.
	//This is important because the database will be experiencing updates (adding & deleting)
	EcleanErr := os.Truncate(EventPath, 0)
	if EcleanErr != nil{
		panic(EcleanErr.Error())
	}


	//Querying all of the events
	allEvents, QEerr := db.Query(`SELECT Event_Name, Date, Time, Event_Location, What FROM Events WHERE Home_Community = (SELECT Community_ID FROM Communities WHERE Name = ?) ORDER BY Event_ID DESC`, CurrentCommunity)
	if QEerr != nil{
		panic(QEerr.Error())
	}

	//Closing the file, opening a new writer
	defer file.Close()
	EventToHtml := bufio.NewWriter(Eventfile)

	//For each entry in the event table
	//Grab all of the details in the event table
	//Then spit HTML into the template in order to render the post entry
	for allEvents.Next(){
		var EventName string
		var Date string
		var Time string
		var Location string
		var Details string
		readerr := allEvents.Scan(&EventName, &Date, &Time, &Location, &Details)
		if readerr != nil{
			panic(readerr.Error())
		}

		htmlToRender := fmt.Sprintf(`
		<div class = "eventWindow">
			<b> %s </b><br/>
			<i>Date : </i>
			%s
			<br/>
			<i>Time : </i>
			%s
			<br/>
			<i>Location : </i>
			%s
			<br/>
			<br/>
			%s
		</div>
		`, EventName, Date, Time, Location, Details)
		_, htmlRenderErr := EventToHtml.WriteString(htmlToRender)

		if htmlRenderErr != nil{
			panic(htmlRenderErr.Error())
		}
	}

	if err := EventToHtml.Flush(); err != nil {
		panic(err.Error())
	}

	return c.Redirect(App.Community)
}

//Renders the New Post page
func (c App) NewPost() revel.Result{{}
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	return c.Render()
}

//Renders the New Event page
func (c App) NewEvent() revel.Result{
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	return c.Render()
}

func (c App) ConstructEvent(EventTitle string, EventDay string, EventTime string, EventLocation string, EventContent string) revel.Result{
	// Error, Title, Description, and Number of Event Variables
	var err error
	var TitleExists int
	var DescriptionExists int
	var numberOfEvents int

	//Querying to check if there is an existing title: Part of guarding from reposts
	err = db.QueryRow(`SELECT COUNT(Event_Name) FROM Events WHERE Event_Name = ?`, EventTitle).Scan(&TitleExists)
	if err != nil{
		panic(err.Error())
	}

	//If there is already an event with the title, display an error message, and redirect to form. 
	if TitleExists !=0{
		c.Flash.Error("An event with that title already exists")
		return c.Redirect(App.NewPost)
	}

	//Querying to check if there is a post with the same description: Part of guarding from reposts 
	err = db.QueryRow(`SELECT COUNT(What) FROM Events WHERE What = ?`, EventContent).Scan(&DescriptionExists)
	if err != nil{
		panic(err.Error())
	}

	//If there is already a post with the description, display an error message, and redirect to form
	if DescriptionExists != 0{
		c.Flash.Error("An event with the description already exists")
		return c.Redirect(App.NewPost)
	}

	//Counting how many posts are in the database, and storing that count 
	EventCountQuery := fmt.Sprintf(`SELECT COUNT(Event_ID) FROM Events`)
	err = db.QueryRow(EventCountQuery).Scan(&numberOfEvents)
	if err != nil{
		panic(err.Error())
	}

	//Grabbing the active community
	var activeCommunityID int
	Qerr :=db.QueryRow(`SELECT Community_ID FROM Communities WHERE Name = ?`, ActiveCommunity).Scan(&activeCommunityID)
	if Qerr != nil{
		panic(Qerr.Error())
	}

	//Inserting the post into the post database.
	_, Loaderr := db.Exec(`INSERT INTO Events(Event_ID, Event_Name, Date, Time, Event_Location, What, Home_Community) VALUES (?, ?, ?, ?, ?, ?, ?)`, numberOfEvents, EventTitle, EventDay, EventTime, EventLocation, EventContent, activeCommunityID)
	if Loaderr != nil{
		panic(Loaderr.Error())
	}

	//Creation of a new post, and redirecting to the homepage
	c.Flash.Success("Event Created!")
	return c.LoadAssociatedData(ActiveCommunity, ActiveComDescription)


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

//Loads all communities from the database into the community window
//TODO: Also load their positions into the map
func LoadAllCommunities(){

    //Opening the template that is responsible for holding the HTML for all community entries. 
	path := "app/views/Communities.html"
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err !=nil{
		panic(err)
	}

	//Clears out any existing entries in the file.
	//This is important because the database will be experiencing updates (adding & deleting)
	cleanErr := os.Truncate(path, 0)
	if cleanErr != nil{
		panic(cleanErr)
	}

	//Grabbing all of the communities
	allCommunitiesQuery := `SELECT Name, Description FROM Communities ORDER BY Community_ID DESC`
	allCommunities, Qerr := db.Query(allCommunitiesQuery)
	if Qerr != nil{
		panic(Qerr)
	}

	//Closing the file, opening a new writer
	defer file.Close()
	sqlToHtml := bufio.NewWriter(file)

	//For each entry in the community table
	//Grab the name and description of the community
	//Then spit HTML into the template in order to render the community entry
	for allCommunities.Next(){
		var Name string
		var Description string
		
		readerr := allCommunities.Scan(&Name, &Description)
		if readerr != nil{
			panic(readerr.Error())
		}
		//Trims long descriptions to a length that fits inside the window
		if len(Description) > 100 {
			Description = Description[0:100] + "..."
		}
		ValueName := fmt.Sprintf("value=\"%s\"", Name)
		ValueDesc := fmt.Sprintf("value=\"%s\"", Description)
		htmlToRender := fmt.Sprintf(`
			<div class = "communityWindow">
				<b>%s</b><br/>
				%s
				<form action="/LoadAssociatedData" method="POST" right="1%%">
					<input type="hidden" name="CurrentCommunity" %s >
					<input type="hidden" name="CurrentCommunityDescription" %s>
					<button type="submit">Visit Community</button><br>
				</form>
		    </div>
			`, Name, Description, ValueName, ValueDesc)
		_, htmlRenderErr := sqlToHtml.WriteString(htmlToRender)
		if htmlRenderErr != nil{
			panic(htmlRenderErr.Error())
		}
	}
	if err := sqlToHtml.Flush(); err != nil {
        panic(err.Error())
    }

}

//Loads all posts into the Community Window
func LoadAllPosts(){
    //Opening the template that is responsible for holding the HTML for all post entries. 
	path := "app/views/LatestPosts.html"
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err !=nil{
		panic(err)
	}

	//Clears out any existing entries in the file.
	//This is important because the database will be experiencing updates (adding & deleting)
	cleanErr := os.Truncate(path, 0)
	if cleanErr != nil{
		panic(cleanErr)
	}

	//Grabbing all of the posts
	allPostsQuery := `SELECT Title, Text, Community FROM Posts ORDER BY Post_ID DESC`
	allPosts, Qerr := db.Query(allPostsQuery)
	if Qerr != nil{
		panic(err)
	}

	//Closing the file, opening a new writer
	defer file.Close()
	sqlToHtml := bufio.NewWriter(file)

	//For each entry in the post table
	//Grab the name and description of the post
	//Then spit HTML into the template in order to render the post entry
	for allPosts.Next(){
		var Title string
		var Text string
		var CommunityID int
		var Community string
		readerr := allPosts.Scan(&Title, &Text, &CommunityID)
		if readerr != nil{
			panic(readerr.Error())
		}
		//Trims long posts to a length that fits inside the window
		if len(Text) > 100 {
			Text = Text[0:100] + "..."
		}
		Cerr := db.QueryRow(`SELECT Name FROM Communities WHERE Community_ID = ?`, CommunityID).Scan(&Community)
		if Cerr != nil{
			panic(Cerr.Error())
		}

		htmlToRender := fmt.Sprintf(`
		<div class = "postWindow">
			<b> %s </b><br/>
			%s<br/>
			<br/>
			<b>From Community %s</b>
		</div>
		`, Title, Text, Community)
		_, htmlRenderErr := sqlToHtml.WriteString(htmlToRender)

		if htmlRenderErr != nil{
			panic(htmlRenderErr.Error())
		}
	}

	if err := sqlToHtml.Flush(); err != nil {
		panic(err.Error())
	}

} 