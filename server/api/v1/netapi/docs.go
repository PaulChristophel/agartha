package netapi

import (
	_ "github.com/PaulChristophel/agartha/server/httputil"
)

type Accept struct {
	Accept string `json:"accept" example:"application/x-yaml"`
}

type Tag struct {
	Tag string `json:"tag" example:"mycompany/myapp/mydata"`
}

type AuthToken struct {
	XAuthToken string `json:"X-Auth-Token" binding:"required" example:"6572118dfaaaa84363ebf491c98e669105f2b1db"`
}

type GenericReturn struct {
	Return []string `json:"return" example:"{dict of return data}"`
}

type WelcomeReturn struct {
	Return  string   `json:"return" example:"Welcome"`
	Clients []string `json:"clients" example:"local"`
}

type Success struct {
	Status bool `json:"success" example:"true"`
}

type LogoutReturn struct {
	Return string `json:"return" example:"Your token has been cleared"`
}

type SaltRequestBody struct {
	Client  string `json:"client" example:"local_async"`
	Fun     string `json:"fun" example:"test.ping"`
	TGT     string `json:"tgt" example:"*"`
	TGTType string `json:"tgt_type" example:"glob"`
}

type HookEvent struct {
	Foo string `json:"foo" example:"Hello"`
	Bar string `json:"bar" example:"World!"`
}

// class salt.netapi.rest_cherrypy.app.LowDataAdapter
//
//	@ID				LowDataAdapter.GET()
//	@Summary		Send one or more Salt commands in the request body.
//	@Description	The primary entry point to Salt's REST API. Send one or more Salt commands in the request body. https://docs.saltproject.io/en/latest/ref/netapi/all/salt.netapi.rest_cherrypy.html#salt.netapi.rest_cherrypy.app.LowDataAdapter.POST
//	@Tags			NetAPI
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	WelcomeReturn
//	@Failure		401				{object}	httputil.HTTPError401
//	@Failure		406				{object}	httputil.HTTPError406
//	@Failure		500				{object}	httputil.HTTPError500
//	@Param			X-Auth-Token	header		AuthToken	true	"a session token from Login"
//	@Param			Accept			header		Accept		false	"the desired response format"
//	@router			/api/v1/netapi/ [get]
//	@Security		Bearer
func RootGet() {}

// class salt.netapi.rest_cherrypy.app.LowDataAdapter
//
//	@ID				LowDataAdapter.POST()
//	@Summary		Send one or more Salt commands in the request body.
//	@Description	The primary entry point to Salt's REST API. Send one or more Salt commands in the request body. https://docs.saltproject.io/en/latest/ref/netapi/all/salt.netapi.rest_cherrypy.html#salt.netapi.rest_cherrypy.app.LowDataAdapter.POST
//	@Tags			NetAPI
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	GenericReturn
//	@Failure		401				{object}	httputil.HTTPError401
//	@Failure		406				{object}	httputil.HTTPError406
//	@Failure		500				{object}	httputil.HTTPError500
//	@Param			X-Auth-Token	header		AuthToken			true	"a session token from Login"
//	@Param			Accept			header		Accept				false	"the desired response format"
//	@Param			req				body		[]SaltRequestBody	true	"Request Body"
//	@router			/api/v1/netapi/ [post]
//	@Security		Bearer
func RootPost() {}

// class salt.netapi.rest_cherrypy.app.Logout(*args, **kwargs)
//
//	@ID				Logout.Post()
//	@Summary		Log out to expire the session token.
//	@Description	Log out to expire the session token. https://docs.saltproject.io/en/latest/ref/netapi/all/salt.netapi.rest_cherrypy.html#salt.netapi.rest_cherrypy.app.Logout.POST
//	@Tags			NetAPI
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	LogoutReturn
//	@Failure		401				{object}	httputil.HTTPError401
//	@Failure		406				{object}	httputil.HTTPError406
//	@Failure		500				{object}	httputil.HTTPError500
//	@Param			X-Auth-Token	header		AuthToken	true	"a session token from Login"
//	@router			/api/v1/netapi/logout [post]
//	@Security		Bearer
func LogoutPost() {}

// class salt.netapi.rest_cherrypy.app.Hook(*args, **kwargs)
//
//	@ID				Hook.Post()
//	@Summary		Send an event to Salt's event bus.
//	@Description	Send an event to Salt's event bus. https://docs.saltproject.io/en/latest/ref/netapi/all/salt.netapi.rest_cherrypy.html#salt.netapi.rest_cherrypy.app.Hook.POST
//	@Tags			NetAPI
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	Success
//	@Failure		401				{object}	httputil.HTTPError401
//	@Failure		406				{object}	httputil.HTTPError406
//	@Failure		413				{object}	httputil.HTTPError413
//	@Failure		500				{object}	httputil.HTTPError500
//	@Param			X-Auth-Token	header		AuthToken	true	"a session token from Login"
//	@Param			Accept			header		Accept		false	"the desired response format"
//	@Param			tag				path		Tag			false	"optional tag for the request"
//	@Param			req				body		HookEvent	true	"Hook Event Data"
//	@router			/api/v1/netapi/hook/{tag} [post]
//	@Security		Bearer
func HookPost() {}

// class salt.netapi.rest_cherrypy.app.Stats(*args, **kwargs)
//
//	@ID				Stats.Get()
//	@Summary		Return statistics about the running CherryPy process.
//	@Description	Return statistics about the running CherryPy process. https://docs.saltproject.io/en/latest/ref/netapi/all/salt.netapi.rest_cherrypy.html#salt.netapi.rest_cherrypy.app.Stats.GET
//	@Tags			NetAPI
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	Stats
//	@Failure		401				{object}	httputil.HTTPError401
//	@Failure		406				{object}	httputil.HTTPError406
//	@Failure		500				{object}	httputil.HTTPError500
//	@Param			X-Auth-Token	header		AuthToken	true	"a session token from Login"
//	@Param			Accept			header		Accept		false	"the desired response format"
//	@Security		Bearer
func StatsGet() {}
