package handler

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SetFlash stores a one-time message in the session.
func SetFlash(c *gin.Context, key, value string) {
	sess := sessions.Default(c)
	sess.Set("flash_"+key, value)
	_ = sess.Save()
}

// PopFlash returns the message and deletes it from the session.
func PopFlash(c *gin.Context, key string) (string, bool) {
	sess := sessions.Default(c)
	v := sess.Get("flash_" + key)
	if v == nil {
		return "", false
	}
	sess.Delete("flash_" + key)
	_ = sess.Save()
	s, ok := v.(string)
	return s, ok
}

// PopFlashCheck peeks at a flash value without removing it.
func PopFlashCheck(c *gin.Context, key string) string {
	sess := sessions.Default(c)
	v := sess.Get("flash_" + key)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

