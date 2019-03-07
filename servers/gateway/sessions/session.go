package sessions

import (
	"errors"
	"net/http"
	"strings"
)

const headerAuthorization = "Authorization"
const paramAuthorization = "auth"
const schemeBearer = "Bearer "

//ErrNoSessionID is used when no session ID was found in the Authorization header
var ErrNoSessionID = errors.New("no session ID found in " + headerAuthorization + " header")

//ErrInvalidScheme is used when the authorization scheme is not supported
var ErrInvalidScheme = errors.New("authorization scheme not supported")

//BeginSession creates a new SessionID, saves the `sessionState` to the store, adds an
//Authorization header to the response with the SessionID, and returns the new SessionID
func BeginSession(signingKey string, store Store, sessionState interface{}, w http.ResponseWriter) (SessionID, error) {
	sid, err := NewSessionID(signingKey) // make new SessionID
	if err != nil {
		return InvalidSessionID, err
	}

	err = Store.Save(store, sid, sessionState) // save data in store
	if err != nil {
		return InvalidSessionID, err
	}

	w.Header().Set(headerAuthorization, schemeBearer+string(sid)) // add auth header

	return sid, nil
}

//GetSessionID extracts and validates the SessionID from the request headers
func GetSessionID(r *http.Request, signingKey string) (SessionID, error) {
	auth := r.Header.Get(headerAuthorization)
	if len(auth) == 0 { // no Auth header
		auth = r.URL.Query().Get(paramAuthorization)
	}

	vals := strings.Fields(auth) // get scheme and SessionID
	if len(vals) < 2 {           // only got scheme Bearer
		return InvalidSessionID, ErrNoSessionID
	}
	if vals[0] != strings.TrimSpace(schemeBearer) { // Invalid Scheme Bearer
		return InvalidSessionID, ErrInvalidScheme
	}

	sid, err := ValidateID(vals[1], signingKey) // validate SessionID
	if err != nil {
		return InvalidSessionID, err
	}
	return sid, nil
}

//GetState extracts the SessionID from the request,
//gets the associated state from the provided store into
//the `sessionState` parameter, and returns the SessionID
func GetState(r *http.Request, signingKey string, store Store, sessionState interface{}) (SessionID, error) {
	sid, err := GetSessionID(r, signingKey) // get session id from request
	if err != nil {
		return InvalidSessionID, err
	}

	err = store.Get(sid, sessionState) // get state from store
	if err != nil {
		return InvalidSessionID, err
	}
	return sid, nil
}

//EndSession extracts the SessionID from the request,
//and deletes the associated data in the provided store, returning
//the extracted SessionID.
func EndSession(r *http.Request, signingKey string, store Store) (SessionID, error) {
	sid, err := GetSessionID(r, signingKey) // get SessionID from request
	if err != nil {
		return InvalidSessionID, err
	}

	err = store.Delete(sid) // delete data from store
	if err != nil {
		return InvalidSessionID, err
	}
	return sid, nil
}
