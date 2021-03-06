package sessions

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"reflect"
)

//InvalidSessionID represents an empty, invalid session ID
const InvalidSessionID SessionID = ""

//idLength is the length of the ID portion
const idLength = 32

//signedLength is the full length of the signed session ID
//(ID portion plus signature)
const signedLength = idLength + sha256.Size

//SessionID represents a valid, digitally-signed session ID.
//This is a base64 URL encoded string created from a byte slice
//where the first `idLength` bytes are crytographically random
//bytes representing the unique session ID, and the remaining bytes
//are an HMAC hash of those ID bytes (i.e., a digital signature).
//The byte slice layout is like so:
//+-----------------------------------------------------+
//|...32 crypto random bytes...|HMAC hash of those bytes|
//+-----------------------------------------------------+
type SessionID string

//ErrInvalidID is returned when an invalid session id is passed to ValidateID()
var ErrInvalidID = errors.New("Invalid Session ID")

//NewSessionID creates and returns a new digitally-signed session ID,
//using `signingKey` as the HMAC signing key. An error is returned only
//if there was an error generating random bytes for the session ID
func NewSessionID(signingKey string) (SessionID, error) {
	if len(signingKey) == 0 {
		return InvalidSessionID, errors.New("Signing Key must not have size 0")
	}
	bytes := []byte{}
	rand, err := randCrypto(idLength)
	if err != nil {
		return InvalidSessionID, err
	}
	bytes = append(bytes, rand...)

	sig := hmacHash([]byte(signingKey), bytes[0:31])
	bytes = append(bytes, sig...)

	return SessionID(base64.URLEncoding.EncodeToString(bytes)), nil
}

//hmacHash creates an HMAC digital signature using a byte slice key
//and the message msg to use. Returns the digital signature as a byte slice
func hmacHash(key []byte, msg []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(msg[0:31])
	return h.Sum(nil)
}

//randCrypto takes a size integer and returns a cryptographically
//random generated byte slice of that size. An error is returned if size
//is 0
func randCrypto(size int) ([]byte, error) {
	if size < 1 {
		return nil, errors.New("invalid size parameter")
	}
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, errors.New("Error generating random bytes")
	}
	return bytes, nil
}

//ValidateID validates the string in the `id` parameter
//using the `signingKey` as the HMAC signing key
//and returns an error if invalid, or a SessionID if valid
func ValidateID(id string, signingKey string) (SessionID, error) {

	//TODO: validate the `id` parameter using the provided `signingKey`.
	//base64 decode the `id` parameter, HMAC hash the
	//ID portion of the byte slice, and compare that to the
	//HMAC hash stored in the remaining bytes. If they match,
	//return the entire `id` parameter as a SessionID type.
	//If not, return InvalidSessionID and ErrInvalidID.

	bytes, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		return "", err
	}
	sig := hmacHash([]byte(signingKey), bytes[0:31])
	if !reflect.DeepEqual(sig, bytes[32:]) {
		return InvalidSessionID, ErrInvalidID
	}
	return SessionID(id), nil
}

//String returns a string representation of the sessionID
func (sid SessionID) String() string {
	return string(sid)
}
