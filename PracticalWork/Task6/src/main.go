package main

import (
	"fmt"
	"context"
	"time"
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	"github.com/hashicorp/vault/api"
	approle "github.com/hashicorp/vault/api/auth/approle"
	cron "github.com/robfig/cron/v3"
)

var fileHandler *os.File

func main()  {
	var err error

	filePath := os.Getenv("LOG_FILE_PATH")
	fileHandler, err = os.Create(filePath)
	if err != nil {
        writeLogs(os.Stdout, err.Error())
    }

	config := api.DefaultConfig()
    err = config.ReadEnvironment()
    if err != nil {
        writeLogs(os.Stdout, err.Error())
    }

	client, err := api.NewClient(api.DefaultConfig())
    if err != nil {
        writeLogs(os.Stdout, err.Error())
    }

	appRoleAuth, err := approle.NewAppRoleAuth(
		os.Getenv("VAULT_HEALTHCHECK_ROLE_ID"),
		&approle.SecretID{FromEnv: "VAULT_HEALTHCHECK_SECRET_ID"},
	)
	if err != nil {
        writeLogs(os.Stdout, err.Error())
    }

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()

    secret, err := appRoleAuth.Login(context.Background(), client)
	if ctx.Err() == context.DeadlineExceeded {
		writeLogs(os.Stdout, "Запрос превысил лимит времени!")
	}
    if err != nil {
        writeLogs(os.Stdout, err.Error())
    }

    if secret == nil || secret.Auth == nil || secret.Auth.ClientToken == "" {
        writeLogs(os.Stdout, "Не удалось получить токен после аутентификации")
    }

    client.SetToken(secret.Auth.ClientToken)


	mountPath := os.Getenv("VAULT_MOUNT_POINT")
    secretPath := os.Getenv("VAULT_HEALTHCHECK_SECRET_PATH")
	
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()

    readSecret, err := client.Logical().ReadWithContext(ctx, mountPath + "/data/" + secretPath)
	if ctx.Err() == context.DeadlineExceeded {
		writeLogs(os.Stdout, "Запрос превысил лимит времени!")
	}
    if err != nil {
        writeLogs(os.Stdout, err.Error())
    }
    
    if readSecret == nil {
        writeLogs(os.Stdout, "Секрет не найден.")
    }

    secretData, ok := readSecret.Data["data"].(map[string]any)
    if !ok {
        writeLogs(os.Stdout, "Не удалось распарсить содержимое секрета.")
    }

    user := getSecret(secretData, "user")
	password := getSecret(secretData, "password")
	hostPath := getSecret(secretData, "host")
    databasePath := getSecret(secretData, "database")

	healthcheckPeriod := os.Getenv("HEALTHCHECK_PERIOD_IN_MINUTES")
	cr := cron.New(cron.WithSeconds())
	taskTimer := fmt.Sprintf("*/%s * * * * *", healthcheckPeriod)
	id, err := cr.AddFunc(taskTimer, func () {task(user, password, hostPath, databasePath)})
	if (err !=nil){
		writeLogs(os.Stdout, err.Error())
		return
	}

	defer cr.Remove(id)

	cr.Start()

	select {}
}

func task(username string, password string, hostPath string, databasePath string) {

	connStr := "postgresql://" + username + ":" + password + "@" + hostPath + "/" + databasePath + "?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
    if err != nil {
        writeLogs(os.Stdout, err.Error())
    }
	defer db.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()

	var version string
	err = db.QueryRowContext(ctx, "SELECT VERSION();").Scan(&version)
	if ctx.Err() == context.DeadlineExceeded {
		writeLogs(os.Stdout, "Запрос превысил лимит времени!")
	}
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

func getSecret(secretData map[string]interface{}, secretName string) (string){
	secretValue, ok := secretData[secretName].(string)
	if !ok {
        writeLogs(os.Stdout, "Не удалось найти секрет.")
		return ""
    }
	return secretValue
}