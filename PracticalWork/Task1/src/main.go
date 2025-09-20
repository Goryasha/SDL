package main

import (
	"fmt"
	"context"
	"time"
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	"bufio"
	"io"
	"strings"
	"regexp"
	"github.com/BurntSushi/toml"
)

type Config struct {
    Title string `toml:"title"`
    Server ReaderConfig `toml:"reader"` // вложенная секция server
    Database DatabaseConfig `toml:"database"` // вложенная секция database
}

type DatabaseConfig struct {
    Host string `toml:"host"`
    PostgresDB string `toml:"postgres_db"`
}

type ReaderConfig struct {
    MaxUsernameBytes int `toml:"username_length_in_bytes"`
    MaxPasswordBytes int `toml:"password_length_in_bytes"`
}

func main()  {
	var config Config

	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
        panic(err)
    }

// Возможные символы для userinfo
// 		if 'A' <= r && r <= 'Z' {
// 		if 'a' <= r && r <= 'z' {
// 		if '0' <= r && r <= '9' {
// 		case '-', '.', '_', ':', '~', '!', '$', '&', '\'',
// 			'(', ')', '*', '+', ',', ';', '=', '%', '@':

	connStrPref := "postgresql://"
	connStrSuff := "@" + config.Database.Host + "/" + config.Database.PostgresDB + "?sslmode=disable"
	
	// https://github.com/lib/pq/blob/master/conn.go - описана регистрация драйвера sql.Register("postgres", &Driver{})
	// Далее идем по методам. Open -> DialOpen -> NewConnector, в котором и содержится логика мапинга dsn в когфиг для БД
	// Логика описана в  https://github.com/lib/pq/blob/master/connector.go
	// За первичный анализ строки отвечает ParseURL() из https://github.com/lib/pq/blob/master/url.go
	// Вся логика парсинга лежит на базово библиотеке net/url
	// 1. Проверка на наличие экранированных символов через # <- нет смысла давать ему возможность избежать базовой логики.
	// 2. Проверка на наличие ASCII control character (есть = err)
	// 3. Получение схемы (postgresql://) (по первому вхождению :)
	// 4. Разделение rest и querry (по первому вхождению ?) <- единственное место, где пользователь при вводе пароля сможет изменить свой host, path и querry.
	// 5. Определение authority и path частей  (по первому вхождению /)
	// 6. Разделение host и userinfo частей (по последнему вхождению @)
	// 7. Разделение логина и пароля (по первому вхождению :)
	// 8. Все это записывается в url.Url
	// 9. Через accrue соединяем параметры в строку key=value
	// 10. Создаем сущность Connector с указанными параметрами, которая и участвует в открытии сессии до БД

	username, err := readSTDIN(config.Server.MaxUsernameBytes, "имя пользователя")

	if (err !=nil){
		panic(err)
	}

	password, err := readSTDIN(config.Server.MaxPasswordBytes, "пароль")

	if (err !=nil){
		panic(err)
	}

	connStr := connStrPref + username + ":" + password + connStrSuff

	db, err := sql.Open("postgres", connStr)
    if err != nil {
        panic(err)
    }
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()

	var version string
	err = db.QueryRowContext(ctx, "SELECT VERSION();").Scan(&version)
	if ctx.Err() == context.DeadlineExceeded {
		panic("Запрос превысил лимит времени!")
	}
    if err != nil{
        panic(err)
    }

	fmt.Printf("%s", version)

}

func readSTDIN(ByteLimit int, text string) (value string, err error) {	
	reader := bufio.NewReader(io.LimitReader(os.Stdin, int64(ByteLimit)))
	
	fmt.Printf("Введите %s: ", text)
    value, err = reader.ReadString('\n')
	value = strings.TrimSuffix(value, "\n")

    if err != nil && err != io.EOF {
        fmt.Printf("Ошибка чтения: %v\n", err)
        return "", err
	}

	re := regexp.MustCompile(`[#?]`)
	value = re.ReplaceAllString(value, "")

	return value, nil
}