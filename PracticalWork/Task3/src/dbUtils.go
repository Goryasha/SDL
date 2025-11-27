package main


import (
	"fmt"
	"os"
	"database/sql"
	"encoding/json"
	"time"
	"context"
	"reflect"
)


func execQueryAndPrintResult(db *sql.DB, query string, placeholders ...any) (map[string]any){
	rows, cancel := executeQuery(db, query, placeholders...)
	defer cancel()
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Fprintln(os.Stderr,"Ошбика чтения названия колонок в ответе от БД.")
		os.Exit(-1)
	}

	values := make([]any, len(columns))
	scans := make([]any, len(values))
	for i := range values {
		scans[i] = &values[i]
	}

	var record map[string]any

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			fmt.Fprintln(os.Stderr,"Ошибка чтения ответов от БД.")
			os.Exit(-1)
		}

		record = make(map[string]any)
		for k, v := range columns {
            value := values[k]
            if value == nil || reflect.ValueOf(value).IsZero() {
                record[v] = nil
            } else if b, ok := value.([]byte); ok {
                record[v] = string(b)
            } else {
                record[v] = value
            }
        }
		
		jsonData, err := json.MarshalIndent(record, "", "  ")
		if (err != nil) {
			fmt.Fprintln(os.Stderr,"Ошибка формирования json.")
			os.Exit(-1)
		}
		if (len(jsonData) == 0){
			fmt.Print("Данных по запросу не нашлось!\n")
		}
		fmt.Println(string(jsonData))
	}

	if err = rows.Err(); err != nil {
		fmt.Fprintln(os.Stderr,"Ошибка итерации по ответам от БД.")
		os.Exit(-1)
	}

	return record
}

func announceNotNull(db *sql.DB, table string)(){
	rows, cancel := executeQuery(db, "SELECT column_name FROM information_schema.columns WHERE table_name = $1 AND is_nullable = 'NO'", table)
	defer cancel()
	defer rows.Close()

	columns := make([]string, 0)
    for rows.Next() {
        var col string
        err := rows.Scan(&col)
        if err != nil {
            fmt.Fprintln(os.Stderr,"Ошибка чтения ответов от БД.")
			os.Exit(-1)
        }
        columns = append(columns, col)
    }

	fmt.Printf("Внимание, данные параметры %s обязательны для данной таблицы!\n",columns)
}

func getTableStructure(db *sql.DB, table string) ([]string){
	rows, cancel := executeQuery(db, "SELECT column_name FROM information_schema.columns WHERE table_name = $1", table)
	defer cancel()
	defer rows.Close()

	columns := make([]string, 0)
    for rows.Next() {
        var col string
        err := rows.Scan(&col)
        if err != nil {
            fmt.Fprintln(os.Stderr,"Ошибка чтения ответов от БД.")
			os.Exit(-1)
        }
        columns = append(columns, col)
    }

	return columns
}

func chooseConnectedTables(db *sql.DB, table string) ([]string){
	TABLE_BYTES := 4

	rows, cancel := executeQuery(db, "SELECT cls.relname, att.attname FROM pg_constraint con JOIN pg_class cls ON cls.oid = con.confrelid JOIN pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey) WHERE con.conrelid = ($1)::regclass AND con.contype = 'f'", "shop."+table)
	defer cancel()
	defer rows.Close()

	connectedCounter := 1
	connectedTables := map[string][]string{}
    for rows.Next() {
        var col string
        var col2 string
        err := rows.Scan(&col, &col2)
        if err != nil {
            fmt.Fprintln(os.Stderr,"Ошибка чтения ответов от БД.")
			os.Exit(-1)
        }
		connectedTables[fmt.Sprintf("%d", connectedCounter)] = []string{col, col2}
		connectedCounter++
    }

	if (len(connectedTables) != 0){
		fmt.Print("С данной таблицей связаны таблицы (Выберите одну из них):\n")
		for k, v := range connectedTables{
			fmt.Printf("%s - Таблица %s - Поле %s\n", k, v[0], v[1])
		}
	} else {
		fmt.Print("У данной таблицы нет связанных таблиц!\n")
		return nil
	}

	for {
		tableNumber, err := readSTDIN(TABLE_BYTES, "Выберите одну из таблиц\n", "[^0-9]+")
		if (err !=nil){
			fmt.Fprintln(os.Stderr,"Ошибка чтения STDIN.")
			os.Exit(-1)
		}
		
		table2Param, res := connectedTables[tableNumber]
		if (!res){
			fmt.Print("Такой записи нет!\n")
			continue
		}

		return table2Param
	}
}

func executeQuery(db *sql.DB, query string, placeholders ...any)(*sql.Rows, context.CancelFunc){
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, query)
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Fprintln(os.Stderr,"Запрос превысил лимит времени!")
		os.Exit(-1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr,"Ошибка создания Prepare Statement.")
		os.Exit(-1)
	}
	defer stmt.Close()
	
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	rows, err := stmt.QueryContext(ctx, placeholders...)
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Fprintln(os.Stderr,"Запрос превысил лимит времени!")
		os.Exit(-1)
	}
	if err != nil{
		fmt.Fprintln(os.Stderr,"Ошибка выполнения запроса.")
		os.Exit(-1)
	}

	return rows, cancel
}
