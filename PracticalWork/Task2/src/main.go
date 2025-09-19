package main

import (
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	// "path/filepath"
	// "github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
)

var fileHandler *os.File

func main()  {

	// procPath, err := filepath.Abs(os.Args[0])
	// if (err !=nil){
	// 	panic(err)
	// }

	// err = godotenv.Load(filepath.Join(filepath.Dir(filepath.Dir(procPath)), ".env"))
	// if (err !=nil){
	// 	panic(err)
	// }
	var err error

	filePath := os.Getenv("LOG_FILE_PATH")
	fileHandler, err = os.Create(filePath)
	if err != nil {
        writeLogs(os.Stdout, err.Error())
    }

	healthcheckPeriod := os.Getenv("HEALTHCHECK_PERIOD_IN_MINUTES")
	cr := cron.New(cron.WithSeconds())
	taskTimer := fmt.Sprintf("*/%s * * * * *", healthcheckPeriod)
	id, err := cr.AddFunc(taskTimer, task)
	if (err !=nil){
		writeLogs(os.Stdout, err.Error())
		return
	}

	defer cr.Remove(id)

	cr.Start()

	select {}
}

func task() {
	databasePath := os.Getenv("POSTGRES_DB")
	hostPath := os.Getenv("HOST")
	username := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	connStr := "postgresql://" + username + ":" + password + "@" + hostPath + "/" + databasePath + "?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
    if err != nil {
        writeLogs(os.Stdout, err.Error())
    }
	defer db.Close()
	
	var version string
	err = db.QueryRow("SELECT VERSION();").Scan(&version)
    if err != nil{
        writeLogs(os.Stdout, err.Error())
    }

	writeLogs(os.Stdout, fmt.Sprintf("%s\n", version))
}

func writeLogs(baseHandler *os.File, text string){
	fmt.Fprint(baseHandler, text)
	if fileHandler !=nil {
		fmt.Fprint(fileHandler, text)
	}
}