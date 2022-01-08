/*
	This controller handles community activities
*/

package controllers

// Packages used in our project
import (
	"github.com/revel/revel"
	// "database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
    "bufio"
    "os"
	// "strconv"
)

// Defining the controller used in this file
type Community struct {
	*revel.Controller
}

//Renders the community creation page
func (c Community) CreateCommunity() revel.Result{
	//If an attempt is made to access the page without being logged in, remain in Login page
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	return c.Render()
}

//Function that constructs the community
//Called whenever a "ConstructCommunity" form is submitted in CreateCommunity.html
func (c Community) ConstructCommunity(NewCommunityName string, CommunityDescription string) revel.Result{
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
		return c.Redirect(Community.CreateCommunity)
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
func (c Community) ConstructPost(PostTitle string, PostContent string, CurrentCommunity string) revel.Result{
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
		return c.Redirect(Community.NewPost)
	}

	//Querying to check if there is a post with the same description: Part of guarding from reposts 
	err = db.QueryRow(`SELECT COUNT(Text) FROM Posts WHERE Text = ?`, PostContent).Scan(&DescriptionExists)
	if err != nil{
		panic(err.Error())
	}

	//If there is already a post with the description, display an error message, and redirect to form
	if DescriptionExists != 0{
		c.Flash.Error("A post with the description already exists")
		return c.Redirect(Community.NewPost)
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
func (c Community) Community(CurrentCommunity string, CurrentCommunityDescription string) revel.Result{
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	CurrentCommunity = ActiveCommunity
	CurrentCommunityDescription = ActiveComDescription
	return c.Render(CurrentCommunity, CurrentCommunityDescription)
}

//Function that populates the Community Page
func (c Community) LoadAssociatedData(CurrentCommunity string, CurrentCommunityDescription string)revel.Result{
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

	return c.Redirect(Community.Community)
}

//Renders the New Post page
func (c Community) NewPost() revel.Result{{}
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	return c.Render()
}

//Renders the New Event page
func (c Community) NewEvent() revel.Result{
	if(!LoggedIn){
		return c.Redirect(Auth.Login);
	}
	return c.Render()
}

func (c Community) ConstructEvent(EventTitle string, EventDay string, EventTime string, EventLocation string, EventContent string) revel.Result{
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
		return c.Redirect(Community.NewPost)
	}

	//Querying to check if there is a post with the same description: Part of guarding from reposts 
	err = db.QueryRow(`SELECT COUNT(What) FROM Events WHERE What = ?`, EventContent).Scan(&DescriptionExists)
	if err != nil{
		panic(err.Error())
	}

	//If there is already a post with the description, display an error message, and redirect to form
	if DescriptionExists != 0{
		c.Flash.Error("An event with the description already exists")
		return c.Redirect(Community.NewPost)
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
//Grabbing the active community//Loads all communities from the database into the community window
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