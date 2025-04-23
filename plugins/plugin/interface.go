package plugin

import (
	"net/http"

	"example.com/auth"
	"example.com/tes"
)

// Auth is the interface plugins have to implement. To avoid calling the
// plugin for roles it doesn't support, it has to tell the plugin managers
// which roles it wants to be invoked on by implementing the Hooks() method.
type Authorizer interface {
	// Hooks returns a list of the hooks this plugin wants to register.
	// Hooks can have one of the following forms:
	//
	//  * "contents": the plugin's Authorize method will be called on
	//                the post's complete contents.
	//
	// * "role:NN": the plugin's ProcessRole method will be called with role=NN
	//              and the role's value when a :NN: role is encountered in the
	//              input.
	Hooks() []string

	// Authorize is called on the entire post contents, if requested in
	// Hooks(). It takes the contents and the post and should return the
	// transformed contents.
	Authorize(authHeader http.Header, task tes.TesTask) (auth.Auth, error)
}
