package utils

import (
	"bufio"
	"io"
	"fmt"
	"os"
	"strconv"
	"strings"
	"regexp"
)

func GetIntEnv(paramName string) (int){
	param, err := strconv.Atoi(os.Getenv(paramName))
	if (err !=nil){
		fmt.Fprint(os.Stderr,"Ошибка конвертации переменных окружения.")
		os.Exit(-1)
	}
	return param
}

func ConvertMapToQueryAndParams(paramMap map[string]string, table string, action string, searchMap map[string][]int) (string, []any){
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
			fmt.Fprint(os.Stderr,"Невозможный сценарий!")
			os.Exit(-1)
	}
	
	return query, paramValues
}

func ReadString(input string, ByteLimit int, regular string) (value string, err error) {	
	reader := bufio.NewReader(io.LimitReader(strings.NewReader(input), int64(ByteLimit)))
	
    value, err = reader.ReadString('\n')
	value = strings.TrimSuffix(value, "\n")

    if err != nil && err != io.EOF {
        return "", err
	}

	re := regexp.MustCompile(regular)
	value = re.ReplaceAllString(value, "")

	return value, nil
}

func LogOutput(f *os.File, ef *os.File) func() {
	out := os.Stdout
	err := os.Stderr
	mwOut := io.MultiWriter(out, f)
	mwErr := io.MultiWriter(err, f)

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	exitOut := make(chan bool)
	exitErr := make(chan bool)


	go func() {
		_,_ = io.Copy(mwOut, rOut)
		exitOut <- true
	}()

	go func() {
		_,_ = io.Copy(mwErr, rErr)
		exitErr <- true
	}()

	return func() {
		_ = wOut.Close()
		<-exitOut
		_ = wErr.Close()
		<-exitErr
		_ = f.Close()
		_ = ef.Close()
	}
}

func RemoveString(ss []string, s string) []string {
    for i, v := range ss {
        if v == s {
            return append(ss[:i], ss[i+1:]...)
        }
    }
    res := make([]string, len(ss))
    copy(res, ss)
    return res
}
