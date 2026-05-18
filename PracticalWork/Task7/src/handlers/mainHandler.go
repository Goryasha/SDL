package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"

	"connectToDB/utils"
)

var	TABLES = []string{
		"products",
		"product_categories",
		"users",
		"orders",
		"order_details",
	}

var TABLE_BYTES = 20


var InternalCache = cache.New(24*time.Hour, 48*time.Hour)

func LoginHandler(c *gin.Context) {
    c.SetCookie("auth", "", -1, "/", "localhost", false, true)
    c.HTML(http.StatusOK, "login.html", nil)
}


func LoginHandlerPost(c *gin.Context) {

    MaxUsernameBytes := utils.GetIntEnv("USERNAME_LENGTH_IN_BYTES")
	MaxPasswordBytes := utils.GetIntEnv("PASSWORD_LENGTH_IN_BYTES")

    username, err := utils.ReadString(c.PostForm("username"), MaxUsernameBytes, `[#?]`)
	if (err !=nil){
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": "Ошибка чтения ввода.",
        })
        return
	}

	password, err := utils.ReadString(c.PostForm("password"), MaxPasswordBytes, `[#?]`)
	if (err !=nil){
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": "Ошибка чтения ввода.",
        })
        return
	}

    dbConn, error := utils.ConnectToDB(username, password)
    
    if (error != ""){
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": error,
        })
        return
    } else{

        sessionUUID := uuid.New().String()
        cacheKey := "session:" + sessionUUID

        InternalCache.Set(cacheKey, dbConn, cache.DefaultExpiration)
    
        c.SetCookie("auth", sessionUUID, 3600, "/", "localhost", true, true)
        
        c.Redirect(http.StatusSeeOther, "/")
    }
}

func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        cookie, err := c.Request.Cookie("auth")
        if err != nil {
            c.Redirect(http.StatusSeeOther, "/login")
            c.Abort()
            return
        }

        sessionUUID := cookie.Value
        cacheKey := "session:" + sessionUUID


        if cachedValue, found := InternalCache.Get(cacheKey); found {
            if dbConn, ok := cachedValue.(*sql.DB); ok {
                c.Set("db", dbConn)
                c.Next()
                return
            }
        }

        c.SetCookie("auth", "", -1, "/", "localhost", false, true)
        c.Redirect(http.StatusSeeOther, "/login")
        c.Abort()
    }
}

func MainHandler(c *gin.Context) {
    c.HTML(http.StatusOK, "main.html", nil)
}