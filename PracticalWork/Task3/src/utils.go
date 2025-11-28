package main

import (
	"fmt"
	"slices"
	"database/sql"
	"os"
	"strconv"
	"strings"
)

func getIntEnv(paramName string) (int){
	param, err := strconv.Atoi(os.Getenv(paramName))
	if (err !=nil){
		writeLogs(os.Stderr,"Ошибка конвертации переменных окружения.")
		os.Exit(-1)
	}
	return param
}

func getParamsAndValues(minValues int, maxValues int, db *sql.DB, table string, action string) ([]map[string]string, string, []any){
	ROUND_BYTES := 4
	PARAMETER_BYTES := 20
	PARAMETER_VALUE_BYTES := 60
	ACTIONS_MAP := map[string]string{
		"select":", по которому будет происходить поиск",
		"update_one":", который будет изменяться",
		"update_many":", который будет изменяться",
		"insert_one":", который будет добавлен",
		"insert_many":", который будет добавлен",
	}

	var valuesCounter int
	paramsMap := make(map[string]string)
	var paramValues []map[string]string
	rounds := 1

	columns := getTableStructure(db, table)
	if (action == "insert_one" || action == "insert_many"){
		announceNotNull(db, table)
	}

	if (action == "insert_many"){
		roundStr, err :=readSTDIN(ROUND_BYTES, "Введите количество добавляемых строк.", "[^0-9]")
		if (err !=nil){
			writeLogs(os.Stderr,"Ошибка чтения STDIN.")
			os.Exit(-1)
		}
		rounds, err = strconv.Atoi(roundStr)
		if err != nil {
			writeLogs(os.Stderr, "Ошибка конвертации в int.")
			os.Exit(-1)
		}
	}
	
	for i := 0; i < rounds; i++ {
		if (action == "insert_many"){
			writeLogs(os.Stdout, fmt.Sprintf("Введите значения для %d строки.\n", i + 1))
		}

		paramsMap = map[string]string{}
		valuesCounter = 0
		
		for valuesCounter < maxValues {
			paramName, err :=readSTDIN(PARAMETER_BYTES, fmt.Sprintf("Введите параметр%s. Если необходимо закончить, введите '_'.", ACTIONS_MAP[action]), "[^a-z_]")
			if (err !=nil){
				writeLogs(os.Stderr,"Ошибка чтения STDIN.")
				os.Exit(-1)
			}
			if (paramName == "_"){
				if (valuesCounter < minValues){
					writeLogs(os.Stdout, fmt.Sprintf("Аргументов должно быть больше %d.\n", minValues - 1))
					continue
				}else{
					break
				}
			}

			if(!slices.Contains(columns, paramName)){
				writeLogs(os.Stdout, fmt.Sprintf("Такого столбца в таблице нет. Есть только %s\n", columns))
				continue
			}

			if(action == "update_one" && paramName == "id"){
				writeLogs(os.Stdout, "Поле id изменять нельзя!\n")
				continue
			}
			
			paramValue, err :=readSTDIN(PARAMETER_VALUE_BYTES, fmt.Sprintf("Введите значение параметра%s: ", ACTIONS_MAP[action]), "")
			if (err !=nil){
				writeLogs(os.Stderr,"Ошибка чтения STDIN.")
				os.Exit(-1)
			}
			
			paramsMap[paramName] = paramValue
			valuesCounter++
		}
		if (action == "insert_many"){
			paramValues = append(paramValues, paramsMap)
		}
	}
	if (action == "insert_many"){
		return paramValues, "", []any{}
	}
	
	searchValues := []any{}
	if (action == "update_many"){

		for {
			paramName, err :=readSTDIN(PARAMETER_BYTES, "Введите параметр, по которому будет происходить изменение.", "[^a-z_]")
			if (err !=nil){
				writeLogs(os.Stderr,"Ошибка чтения STDIN.")
				os.Exit(-1)
			}
			if(!slices.Contains(columns, paramName)){
				writeLogs(os.Stdout, fmt.Sprintf("Такого столбца в таблице нет. Есть только %s\n", columns))
				continue
			}
			for{
				paramValue, err :=readSTDIN(PARAMETER_VALUE_BYTES, "Введите значение параметра, по которому будет происходить изменение. Если необходимо закончить, введите '_'.", "")
				if (err !=nil){
					writeLogs(os.Stderr,"Ошибка чтения STDIN.")
					os.Exit(-1)
				}
				if (paramValue == "_"){
					paramValues = append(paramValues, paramsMap)
					return paramValues, paramName, searchValues
				}
				searchValues = append(searchValues, paramValue)
			}
		}
	}
	paramValues = append(paramValues, paramsMap)
	return paramValues, "", []any{}
}

func getId() (string){
	ID_BYTES := 10
	id, err :=readSTDIN(ID_BYTES, "Введите id изменяемой записи.\n", "[^0-9]+")
	if (err !=nil){
		writeLogs(os.Stderr,"Ошибка чтения STDIN.")
		os.Exit(-1)
	}
	return id
}

func convertMapToQueryAndParams(paramMap map[string]string, table string, action string, searchMap map[string][]any) (string, []any){
	ACTIONS_MAP := map[string]string{
		"select":"SELECT * FROM shop.%s WHERE ",
		"update_one":"UPDATE shop.%s SET ",
		"update_many":"UPDATE shop.%s SET ",
		"insert_one":"INSERT INTO shop.%s( ",
		"insert_one_connected":"INSERT INTO shop.%s( ",
	}

	var id string
	if (action == "update_one"){
		id = paramMap["id"]
		delete(paramMap, "id")
	}

	query := fmt.Sprintf(ACTIONS_MAP[action], table)
	paramsCounter := 1
	parts := make([]string, 0, len(paramMap))
	partsEnd := make([]string, 0, len(paramMap))
	paramValues := make([]any, 0, len(paramMap))

	if (strings.HasPrefix(action, "insert")){
		for k, v := range paramMap {
			partsEnd = append(partsEnd, fmt.Sprintf("$%d", paramsCounter))
			parts = append(parts, k)

			paramValues = append(paramValues, v)
			paramsCounter++
		}
	}else{
		for k, v := range paramMap {
			part := fmt.Sprintf("%s = $%d", k, paramsCounter)
			parts = append(parts, part)

			paramValues = append(paramValues, v)
			paramsCounter++
		}
	}

	switch action{
		case "select":
			query = query + strings.Join(parts, " AND ")
		case "update_one":
			query = query + strings.Join(parts, ", ") + fmt.Sprintf(" WHERE id = $%d", paramsCounter)
			paramValues = append(paramValues, id)
		case "update_many":
			query = query + parts[0] +" WHERE "

			parts := make([]string, 0, len(paramMap))

			for k, v := range searchMap {
				query = query + k + " IN ("
				for _, val := range v{
					part := fmt.Sprintf("$%d", paramsCounter)
					parts = append(parts, part)

					paramValues = append(paramValues, val)
					paramsCounter++
				}
				query = query + strings.Join(parts, ", ") + " )"
			}
		
		case "insert_one":
			query = query + strings.Join(parts, ", ") + " ) VALUES ( " + strings.Join(partsEnd, ", ") + " )"

		case "insert_one_connected":
			query = query + strings.Join(parts, ", ") + " ) VALUES ( " + strings.Join(partsEnd, ", ") + " ) RETURNING id"

		default:
			writeLogs(os.Stderr,"Невозможный сценарий!")
			os.Exit(-1)
	}
	
	return query, paramValues
}

func writeLogs(baseHandler *os.File, text string){
	if fileHandler !=nil {
		fmt.Fprint(fileHandler, text + "\n")
	}
	fmt.Fprint(baseHandler, text)
}