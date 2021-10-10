package controllers

import (
	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

func (c App) Index(LoginUserName string, LoginPassword string) revel.Result {
	c.Validation.Required(LoginUserName).Message("Username is required.")
    c.Validation.MinSize(LoginUserName, 3).Message("Username is not long enough, it must have at least 3 characters.")
	c.Validation.Required(LoginPassword).Message("Password is required.")
	c.Validation.MinSize(LoginPassword, 5).Message("Password must be at least 5 characters.")

    if c.Validation.HasErrors() {
        c.Validation.Keep()
        c.FlashParams()
        return c.Redirect(App.Login)
    }

	return c.Render(LoginUserName, LoginPassword)
}

func (c App) Login() revel.Result {
	return c.Render()
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