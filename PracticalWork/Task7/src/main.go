package main

import (
	"fmt"
	"os"
	"time"
	"net/http"
	

	"connectToDB/handlers"
	"connectToDB/utils"

	_ "github.com/lib/pq"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/timeout"
)

func main()  {

	var err error

	filePath := os.Getenv("LOG_FILE_PATH")
	errorFilePath := os.Getenv("LOG_ERROR_FILE_PATH")
	fileHandler, err := os.Create(filePath)
	if err != nil {
        fmt.Fprint(os.Stdout, err.Error())
    }
	errorFileHandler, err := os.Create(errorFilePath)
	if err != nil {
        fmt.Fprint(os.Stdout, err.Error())
    }
	fn := utils.LogOutput(fileHandler, errorFileHandler)
	defer fn()
    

	gin.SetMode(gin.ReleaseMode)
    r := gin.Default()

	r.Use(timeout.New(
        timeout.WithTimeout(10*time.Second),
        timeout.WithResponse(func(c *gin.Context) {
			fmt.Fprint(os.Stderr, "Request took too long.\n")
            c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request took too long"})
        }),
    ))

    r.LoadHTMLGlob("templates/*")

    r.GET("/login", handlers.LoginHandler)
    r.POST("/login", handlers.LoginHandlerPost)

    authorized := r.Group("/")
    authorized.Use(handlers.AuthRequired())
    {
        authorized.GET("/", handlers.MainHandler)
        
        
        selectGroup := authorized.Group("/select")
		{
            selectGroup.GET("/:table", handlers.SelectGetHandler) 
            selectGroup.POST("/:table", handlers.SelectPostHandler) 
        }

		updateGroup := authorized.Group("/update")
		{
			updateGroup.GET("/:table", handlers.UpdateGetHandler) 
            updateGroup.POST("/:table", handlers.UpdatePostHandler) 
		}
        
        insertOneGroup := authorized.Group("/insertOne")
		{
			insertOneGroup.GET("/:table", handlers.InsertOneGetHandler) 
            insertOneGroup.POST("/:table", handlers.InsertOnePostHandler) 
		}

		insertManyGroup := authorized.Group("/insertMany")
		{
			insertManyGroup.GET("/:table", handlers.InsertManyGetHandler) 
            insertManyGroup.POST("/:table", handlers.InsertManyPostHandler) 
		}
    }

    fmt.Println("Сервер запущен на :8090")
    if err := r.Run("0.0.0.0:8090"); err != nil {
        fmt.Fprint(os.Stderr, err)
    }

}
