/* Since http is stateless, we need a way to track users
   from one page to the next.

   If a user does not have a GOSESSIONID cookie, we create one and use that id as a key for storage and retrieval of session data 
*/

package birdbrain

import (
	"os"
    "strconv"
    "strings"
    "time"
    "github.com/hoisie/web.go"
)

type SessionStore interface {
    Set(key string, val string, expire int64) os.Error
    Get(key string) (string, os.Error)
    Delete(keys ...string)
}

type Session struct {
    ctx *web.Context
    store SessionStore
    timeout int64
    expiration int64
}

func createKey(vals ...string) string {
	return strings.Join(vals, ":")
}

func NewSession(ctx *web.Context, store SessionStore) *Session {
	// create a session with a default timeout of 1 hour and
	// a default expiration of session cookies of 1 day
	return &Session{ctx: ctx, 
					store: store, 
					timeout: 60*60,
					expiration: 60*60*24}
}

func (s *Session) generateSessionID() (string, bool) {
	// FIXME: temporary. Should probably be more secure
	id := strconv.Itoa64(time.Nanoseconds())
	ok := true
	return id, ok
}

func (s *Session) getSessionID() (string, bool) {
    id, ok := s.ctx.GetSecureCookie("GOSESSIONID")
    return id, ok
}

func (s *Session) setSessionID(id string) {
	s.ctx.SetSecureCookie("GOSESSIONID", id, s.expiration)
	creationTime := strconv.Itoa64(time.Seconds())
	// Set the creation time
	key := createKey("session", id)
	val := string([]uint8(creationTime))
	s.store.Set(key, val, s.expiration)
	// Set the last activity time
	key = createKey("session", id, "last")
	// set the last activity time
	s.store.Set(key, val, s.expiration)
}

func (s *Session) removeSessionID() {
	id, ok := s.getSessionID()
	if !ok {
		/* it doesn't exist to delete, or the user cleared cookies 
		   and it'll just expire and be removed eventually 
		*/
		return
	}
	// remove the session entries corresponding to the ID
	keys := []string{createKey("session", id),
					 createKey("session", id, "last")}
	s.store.Delete(keys...)
	// set the session cookie to expire immediately
	s.ctx.SetSecureCookie("GOSESSIONID", "", -1)
}

func (s *Session) updateLastActivityTime() {
	id, ok := s.getSessionID()
	if !ok {
		return
	}
	key := createKey("session", id, "last")
	curTime := string([]uint8(strconv.Itoa64(time.Seconds())))
	s.store.Set(key, curTime, s.expiration)
}

func (s *Session) isTimedOut() (expired bool) {
	id, ok := s.getSessionID()
	if !ok {
		return true
	}
	key := createKey("session", id, "last")
	lastActivity, err := s.store.Get(key)
	if err != nil {
		return true
	}
	lastActivityTime, err := strconv.Atoi64(lastActivity)
	if err != nil {
		return true
	}
	if lastActivityTime + s.timeout < time.Seconds() {
		return true
	}
	return false
}

func (s *Session) getStoredKeys() []string {
	id, ok := s.getSessionID()
	if !ok {
		return []string{}
	}
	key := createKey("session", id, "keys")
	storedKeys, err := s.store.Get(key)
	if err != nil {
		return []string{}
	}
	return strings.Split(storedKeys, ":")
}

func (s *Session) updateStoredKeys(newKey string) {
	id, ok := s.getSessionID()
	if !ok {
		return
	}
	key := createKey("session", id, "keys")
    storedKeys, err := s.store.Get(key)
    if err != nil {
    	// this is the first stored key
    	s.store.Set(key, newKey, s.expiration)
    } else {
    	// add it to the 'list' of keys stored already
    	s.store.Set(key, storedKeys + ":" + newKey, s.expiration)
    }
}

func (s *Session) Set(key string, val string) (success bool) {
	// try to find a current session
    id, ok := s.getSessionID()
	if !ok {
		// if it doesn't exist, create one
		id, ok = s.generateSessionID()
		if !ok {
			return false
		}
		// set it as the new session
		s.setSessionID(id)
	} else if s.isTimedOut() {
		// if the current session has timed out, remove it
		// and create a new one
		s.removeSessionID()
		id, ok = s.generateSessionID()
		if !ok {
			return false
		}
		// set it as the new session
		s.setSessionID(id)
	}
	sessionKey := createKey("session", id, "key", key)
    s.store.Set(sessionKey, val, s.expiration)
    s.updateLastActivityTime()
    s.updateStoredKeys(key)
	return true
}

func (s *Session) Get(key string) (string, os.Error) {
	id, ok := s.getSessionID()
	if !ok {
		return "", os.NewError("Failed to find session ID")
	}
	if s.isTimedOut() {
		return "", os.NewError("Session has timed out")
	}
	s.updateLastActivityTime()
	sessionKey := createKey("session", id, "key", key)
    return s.store.Get(sessionKey)
}

func (s *Session) Delete(keys ...string) {
	id, ok := s.getSessionID()
	if !ok {
		return
	} 
	// TODO: check the logic of this
	if !s.isTimedOut() {
		s.updateLastActivityTime()
	}
	sessionKeys := []string{}
	for _, key := range keys {
		sessionKey := createKey("session", id, "key", key)
		sessionKeys = append(sessionKeys, sessionKey)
	}
    s.store.Delete(sessionKeys...)
}

func (s *Session) Clear() {
	keys := s.getStoredKeys()
	s.Delete(keys...)
}
