package controllers

import (
	"github.com/revel/revel"
	"image/color"
	"github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
)

type App struct {
	*revel.Controller
}

const (
	addr     = "Lubbock, TX"
	lat, lng = 33.563521, -101.879336
)

func (c App) Index() revel.Result {
	//Consider this as our "Starter Function Area"
	//Instantiate the database here!

	return c.Redirect(App.Login)
}

func (c App) Login() revel.Result {
	return c.Render()
}

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