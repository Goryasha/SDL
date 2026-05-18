package utils


import (
	"fmt"
	"os"
	"database/sql"
	"time"
	"context"
)


func ExecQueryAndParceResult(db *sql.DB, query string, placeholders ...any) ([]map[string]any, string){
	rows, cancel, error := executeQuery(db, query, placeholders...)
	if (error != ""){
		return nil, error
	}
	defer cancel()
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, "Ошбика чтения названия колонок в ответе от БД.\n"
	}

	values := make([]any, len(columns))
	scans := make([]any, len(values))
	for i := range values {
		scans[i] = &values[i]
	}

	var results []map[string]interface{}

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			return nil, "Ошибка чтения ответов от БД.\n"
		}

		rowMap := make(map[string]interface{})
        for i, col := range columns {
            rowMap[col] = values[i]
        }
        results = append(results, rowMap)
    }

	if err = rows.Err(); err != nil {
		fmt.Fprint(os.Stderr,"Ошибка итерации по ответам от БД.\n")
		os.Exit(-1)
	}

	return results, ""
}

func GetTableStructure(db *sql.DB, table string) ([]string, string){
	rows, cancel, err := executeQuery(db, "SELECT column_name FROM information_schema.columns WHERE table_name = $1", table)
	if (err != ""){
		return []string{}, err
	}
	defer cancel()
	defer rows.Close()

	columns := make([]string, 0)
    for rows.Next() {
        var col string
        err := rows.Scan(&col)
        if err != nil {
            fmt.Fprint(os.Stderr,"Ошибка чтения ответов от БД.\n")
			return []string{}, "Ошибка чтения ответов от БД.\n"
        }
        columns = append(columns, col)
    }

	return columns, ""
}

func GetConnectedTables(db *sql.DB, table string) (map[string][]string, string){
	rows, cancel, err := executeQuery(db, "SELECT cls.relname, att.attname FROM pg_constraint con JOIN pg_class cls ON cls.oid = con.confrelid JOIN pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey) WHERE con.conrelid = ($1)::regclass AND con.contype = 'f'", "shop."+table)
	if (err != ""){
		return nil, err
	}
	defer cancel()
	defer rows.Close()

	connectedCounter := 1
	connectedTables := map[string][]string{}
    for rows.Next() {
        var col string
        var col2 string
        err := rows.Scan(&col, &col2)
        if err != nil {
            fmt.Fprint(os.Stderr,"Ошибка чтения ответов от БД.")
			os.Exit(-1)
        }
		connectedTables[fmt.Sprintf("%d", connectedCounter)] = []string{col, col2}
		connectedCounter++
    }

	if (len(connectedTables) == 0){
		return map[string][]string{}, ""
	}

	return connectedTables, ""
}

func executeQuery(db *sql.DB, query string, placeholders ...any)(*sql.Rows, context.CancelFunc, string){
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, query)
	if ctx.Err() == context.DeadlineExceeded {
		return nil, nil, "Запрос превысил лимит времени!\n"
	}
	if err != nil {
		return nil, nil, "Ошибка создания Prepare Statement.\n"
	}
	defer stmt.Close()
	
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	rows, err := stmt.QueryContext(ctx, placeholders...)
	if ctx.Err() == context.DeadlineExceeded {
		return nil, cancel, "Запрос превысил лимит времени!\n"
	}
	if err != nil{
		return nil, cancel, "Ошибка выполнения запроса.\n"
	}

	return rows, cancel, ""
}

func ConnectToDB(username string, password string)(*sql.DB, string){
	databasePath := os.Getenv("POSTGRES_DB")
	hostPath := os.Getenv("HOST")

	connStrPref := "postgresql://"
	connStrSuff := "@" + hostPath + "/" + databasePath + "?sslmode=disable"
	connStr := connStrPref + username + ":" + password + connStrSuff

	db, err := sql.Open("postgres", connStr)
    if err != nil {
        fmt.Fprint(os.Stderr,"Ошибка создзания коннектора к БД.\n")
		return nil, "Ошибка создзания коннектора к БД.\n"
    }

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = db.PingContext(ctx)
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Fprint(os.Stderr,"Запрос превысил лимит времени!")
		return nil, "Запрос превысил лимит времени!"
	}
	if err != nil {
        fmt.Fprint(os.Stderr,"Ошибка доступа к БД.")
		return nil, "Ошибка доступа к БД."
    }

	return db, ""
}

func GetNotNull(db *sql.DB, table string)([]string, string){
	rows, cancel, err := executeQuery(db, "SELECT column_name FROM information_schema.columns WHERE table_name = $1 AND is_nullable = 'NO'", table)
	if (err != ""){
		return nil, err
	}
	defer cancel()
	defer rows.Close()

	columns := make([]string, 0)
    for rows.Next() {
        var col string
        err := rows.Scan(&col)
        if err != nil {
			return nil, "Ошибка чтения ответов от БД."
        }
        columns = append(columns, col)
    }

	return columns, ""
}
