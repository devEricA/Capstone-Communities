/*
	This controller handles the home page, along with global operations and variables
*/

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
   // "fmt"
   // "bufio"
   // "os"
	 "strconv"
   // "time"
   // "log"
)

// Defining the controller used in this file
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