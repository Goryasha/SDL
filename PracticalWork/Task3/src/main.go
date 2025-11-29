package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"

	// "path/filepath"

	// "github.com/joho/godotenv"
)

var fileHandler *os.File

func main()  {

	// procPath, err := filepath.Abs(os.Args[0])
	// if (err !=nil){
	// 	writeLogs(os.Stderr,"Ошибка получения пути процесса.")
	// 	os.Exit(-1)
	// }

	// err = godotenv.Load(filepath.Join(filepath.Dir(filepath.Dir(procPath)), ".env"))
	// if (err !=nil){
	// 	writeLogs(os.Stderr,"Ошибка получения переменных окружения.")
	// 	os.Exit(-1)
	// }

	var err error

	filePath := os.Getenv("LOG_FILE_PATH")
	fileHandler, err = os.Create(filePath)
	if err != nil {
        writeLogs(os.Stdout, err.Error())
    }

	databasePath := os.Getenv("POSTGRES_DB")
	hostPath := os.Getenv("HOST")
	
	MaxUsernameBytes := getIntEnv("USERNAME_LENGTH_IN_BYTES")
	MaxPasswordBytes := getIntEnv("PASSWORD_LENGTH_IN_BYTES")


	connStrPref := "postgresql://"
	connStrSuff := "@" + hostPath + "/" + databasePath + "?sslmode=disable"

	username, err := readSTDIN(MaxUsernameBytes, "Введите имя пользователя: ", `[#?]`)
	if (err !=nil){
		writeLogs(os.Stderr,"Ошибка чтения STDIN.")
		os.Exit(-1)
	}

	password, err := readSTDIN(MaxPasswordBytes, "Введите пароль: ", `[#?]`)
	if (err !=nil){
		writeLogs(os.Stderr,"Ошибка чтения STDIN.")
		os.Exit(-1)
	}

	connStr := connStrPref + username + ":" + password + connStrSuff

	db, err := sql.Open("postgres", connStr)
    if err != nil {
        writeLogs(os.Stderr,"Ошибка создзания коннектора к БД.")
		os.Exit(-1)
    }
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	err = db.PingContext(ctx)
	if ctx.Err() == context.DeadlineExceeded {
		writeLogs(os.Stderr,"Запрос превысил лимит времени!")
		os.Exit(-1)
	}
	if err != nil {
        writeLogs(os.Stderr,"Ошибка доступа к БД.")
		os.Exit(-1)
    }
	cancel()

	sqlScenario(db)

}

func readSTDIN(ByteLimit int, text string, regular string) (value string, err error) {	
	reader := bufio.NewReader(io.LimitReader(os.Stdin, int64(ByteLimit)))
	
	writeLogs(os.Stdout, text)
    value, err = reader.ReadString('\n')
	value = strings.TrimSuffix(value, "\n")

    if err != nil && err != io.EOF {
        writeLogs(os.Stderr, "Ошибка чтения: STDIN")
        return "", err
	}

	re := regexp.MustCompile(regular)
	value = re.ReplaceAllString(value, "")

	return value, nil
}

func sqlScenario(db *sql.DB) () {
	tables := map[string]string{
		"1": "products",
		"2": "product_categories",
		"3": "users",
		"4": "orders",
		"5": "order_details",
	}
	SCENARIO_BYTES := 3

	for {
		scenario, err := readSTDIN(SCENARIO_BYTES, "Выберите действие :\n1 - Select\n2 - Update\n3 - Insert One\n4 - Insert Many\nВведите '-1', чтобы закончить\n", "[^0-9-]+")
		if (err !=nil){
			writeLogs(os.Stderr,"Ошибка чтения STDIN.")
			os.Exit(-1)
		}
		if (scenario == "-1"){
			return
		}

		table, err := readSTDIN(SCENARIO_BYTES, "Выберите Таблицу:\n1-products\n2-product_categories\n3-users\n4-orders\n5-order_details\n", "[^0-9-]+")
		if (err !=nil){
			writeLogs(os.Stderr,"Ошибка чтения STDIN.")
			os.Exit(-1)
		}
		tableName, exists := tables[table]
		if !exists{
			writeLogs(os.Stdout, "Такой таблицы не существует!\n")
			continue
		}


		switch scenario{
			case "1":
				selectScenario(db, tableName)
			case "2":
				updateScenario(db, tableName)
			case "3":
				insertOneScenario(db, tableName)
			case "4":
				insertManyScenario(db, tableName)
			default:
				writeLogs(os.Stdout, "Такой команды не существует!\n")
		}
	}
}

func selectScenario(db *sql.DB, table string) (){
	SCENARIO_BYTES := 3

	scenario, err := readSTDIN(SCENARIO_BYTES, "Выберите действие :\n1 - Выбрать все данные\n2 - Выбрать данные по одному параметру\n3 - Выбрать данные по нескольким параметрам\n", "[^0-9]+")
	if (err !=nil){
		writeLogs(os.Stderr,"Ошибка чтения STDIN.")
		os.Exit(-1)
	}

	switch scenario{
		case "1":
			execQueryAndPrintResult(db, fmt.Sprintf("SELECT * FROM shop.%s", table))
		
		case "2":
			paramMap, _, _ := getParamsAndValues(1,1, db, table, "select")
			query, paramValues := convertMapToQueryAndParams(paramMap[0], table, "select", make(map[string][]any))
			execQueryAndPrintResult(db, query, paramValues...)
		
		case "3":
			paramMap, _, _ := getParamsAndValues(2,5, db, table, "select")
			query, paramValues := convertMapToQueryAndParams(paramMap[0], table, "select", make(map[string][]any))
			execQueryAndPrintResult(db, query, paramValues...)
		default:
			writeLogs(os.Stdout, "Такой команды не существует!\n")
	}
}

func updateScenario(db *sql.DB, table string) (){
	SCENARIO_BYTES := 3

	scenario, err := readSTDIN(SCENARIO_BYTES, "Выберите действие :\n1 - Изменить одну запись\n2 - Изменить несколько записей\n", "[^0-9]+")
	if (err !=nil){
		writeLogs(os.Stderr,"Ошибка чтения STDIN.")
		os.Exit(-1)
	}

	switch scenario{
		case "1":
			id := getId()
			paramMap, _, _ := getParamsAndValues(1,4, db, table, "update_one")
			paramMap[0]["id"] = id
			query, paramValues := convertMapToQueryAndParams(paramMap[0], table, "update_one", make(map[string][]any))
			execQueryAndPrintResult(db, query, paramValues...)

		case "2":
			paramMap, searchName, searchValues := getParamsAndValues(1,1, db, table, "update_many")
			query, paramValues := convertMapToQueryAndParams(paramMap[0], table, "update_many", map[string][]any{searchName:searchValues})
			execQueryAndPrintResult(db, query, paramValues...)
		
		default:
			writeLogs(os.Stdout, "Такой команды не существует!\n")
	}
}

func insertOneScenario(db *sql.DB, table string) (){
	SCENARIO_BYTES := 3

	scenario, err := readSTDIN(SCENARIO_BYTES, "Выберите действие :\n1 - Добавить одну строку в одну таблицу\n2 - Добавить одну строку в несколько связанных таблиц\n", "[^0-9]+")
	if (err !=nil){
		writeLogs(os.Stderr,"Ошибка чтения STDIN.")
		os.Exit(-1)
	}

	switch scenario{
		case "1":
			paramMap, _, _ := getParamsAndValues(1,5, db, table, "insert_one")
			query, paramValues := convertMapToQueryAndParams(paramMap[0], table, "insert_one", make(map[string][]any))
			execQueryAndPrintResult(db, query, paramValues...)

		case "2":
			table2Param := chooseConnectedTables(db, table)
			if (table2Param != nil){
				paramMap, _, _ := getParamsAndValues(1,5, db, table2Param[0], "insert_one")
				query, paramValues := convertMapToQueryAndParams(paramMap[0], table2Param[0], "insert_one_connected", make(map[string][]any))
				output := execQueryAndPrintResult(db, query, paramValues...)

				paramMap, _, _ = getParamsAndValues(1,5, db, table, "insert_one")
				paramMap[0][table2Param[1]] = fmt.Sprintf("%d", output["id"])
				query, paramValues = convertMapToQueryAndParams(paramMap[0], table, "insert_one", make(map[string][]any))
				execQueryAndPrintResult(db, query, paramValues...)
			} else {
				writeLogs(os.Stdout, "Невозможно выполнить сценарий!\n")
			}
		
		default:
			writeLogs(os.Stdout, "Такой команды не существует!\n")
	}
}

func insertManyScenario(db *sql.DB, table string) (){
	SCENARIO_BYTES := 3

	scenario, err := readSTDIN(SCENARIO_BYTES, "Выберите действие :\n1 - Добавить несколько строк в одну таблицу\n2 - Добавить несколько строк в несколько связанных таблиц\n", "[^0-9]+")
	if (err !=nil){
		writeLogs(os.Stderr,"Ошибка чтения STDIN.")
		os.Exit(-1)
	}

	switch scenario{
		case "1":
			paramMap, _, _ := getParamsAndValues(1,5, db, table, "insert_many")
			for _, values := range paramMap{
				query, paramValues := convertMapToQueryAndParams(values, table, "insert_one", make(map[string][]any))
				execQueryAndPrintResult(db, query, paramValues...)
			}

		case "2":
			table2Param := chooseConnectedTables(db, table)
			if (table2Param != nil){
				paramMap, _, _ := getParamsAndValues(1,5, db, table2Param[0], "insert_many")
				for _, values := range paramMap{
					query, paramValues := convertMapToQueryAndParams(values, table2Param[0], "insert_one_connected", make(map[string][]any))
					output := execQueryAndPrintResult(db, query, paramValues...)
					paramMap2, _, _ := getParamsAndValues(1,5, db, table, "insert_one")
					paramMap2[0][table2Param[1]] = fmt.Sprintf("%d", output["id"])
					query, paramValues = convertMapToQueryAndParams(paramMap2[0], table, "insert_one", make(map[string][]any))
					execQueryAndPrintResult(db, query, paramValues...)
				}
			} else {
				writeLogs(os.Stdout, "Невозможно выполнить сценарий!\n")
			}
			
		default:
			writeLogs(os.Stdout, "Такой команды не существует!\n")
	}
}
