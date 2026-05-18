package handlers
import (
	"fmt"
	"net/http"
	"os"
	"slices"
    "database/sql"

	"github.com/gin-gonic/gin"

	"connectToDB/utils"
)

func InsertManyGetHandler(c *gin.Context) {
	dbConn := c.MustGet("db").(*sql.DB)

    table, err := utils.ReadString(c.Param("table"), TABLE_BYTES,"[^a-z_]")
    if (err != nil){
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": err,
        })
        return
    }

    if (!slices.Contains(TABLES, table)){
        fmt.Fprint(os.Stderr, "Таблицы не существует.\n")
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": "Таблицы не существует.\n",
        })
        return
    }

	columns, error := utils.GetTableStructure(dbConn, table)
    if (error != ""){
        fmt.Fprintf(os.Stderr, "%s\n", error)
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": error,
        })
        return
    }

    notNullCollumns, error := utils.GetNotNull(dbConn, table)
    if (error != ""){
        fmt.Fprintf(os.Stderr, "%s\n", error)
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": error,
        })
        return
    }

	mapping, error := utils.GetConnectedTables(dbConn, table)
	if (error != ""){
        fmt.Fprintf(os.Stderr, "%s\n", error)
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": error,
        })
        return
    }

	var result utils.Result

    for _, tableParams := range mapping {
		tableName := tableParams[0]
		connectedField := tableParams[1]
        columns, error := utils.GetTableStructure(dbConn, tableName)
        if (error != ""){
			fmt.Fprintf(os.Stderr, "%s\n", error)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "error",
				"message": error,
			})
			return
		}

        notNullCollumns, error := utils.GetNotNull(dbConn, tableName)
        if (error != ""){
            fmt.Fprintf(os.Stderr, "%s\n", error)
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "error":   "error",
                "message": error,
            })
            return
        }

        result.Tables = append(result.Tables, utils.TableInfo{
            TableName:           tableName,
            Columns:             columns,
            NotNullColumns:      notNullCollumns,
            ConnectedField:      connectedField,
        })
    }
	

    c.HTML(http.StatusOK, "insertMany.html", gin.H{"Table": table, "Columns": columns, "NotNullColumns": notNullCollumns,"ConnectedTables": result})
}

func InsertManyPostHandler(c *gin.Context) {
    dbConn := c.MustGet("db").(*sql.DB)

    table, err := utils.ReadString(c.Param("table"), TABLE_BYTES,"[^a-z_]")
    if (err != nil){
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": err,
        })
        return
    }

    if (!slices.Contains(TABLES, table)){
        fmt.Fprint(os.Stderr, "Таблицы не существует.\n")
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": "Таблицы не существует.\n",
        })
        return
    }
    
    var reqBody utils.RequestInputManyBody

    if err := c.ShouldBindJSON(&reqBody); err != nil {
        fmt.Fprint(os.Stderr, "Невалидный запрос.\n")
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": "Невалидный запрос.\n",
        })
        return
    }

    columns, error := utils.GetTableStructure(dbConn, table)
    if (error != ""){
        fmt.Fprintf(os.Stderr, "%s\n", error)
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "error":   "error",
            "message": error,
        })
        return
    }

    switch reqBody.Scenario{
        case 1:
			paramMap, _, error := InsertManyGetParamsAndValues(&reqBody, columns)
			if (error != nil){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": error.Error(),
                })
				return
            }

            response := make(map[int]string)
            for i, line := range paramMap{
                query, paramValues := utils.ConvertMapToQueryAndParams(line, table, "insert_one", nil)
                _, err := utils.ExecQueryAndParceResult(dbConn, query, paramValues...)
                if (err != ""){
                    fmt.Fprintf(os.Stderr, "%s\n", err)
                    response[i] = err
                } else{
                    response[i] = "OK"
                }
            }

            c.JSON(http.StatusOK, response)
        
        case 2:
            paramMap, paramConnectedMap, error := InsertManyGetParamsAndValues(&reqBody, columns)
			if (error != nil){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": error.Error(),
                })
				return
            }

            if (len(paramMap) != len(paramConnectedMap)){
                fmt.Fprintf(os.Stderr, "%s\n", "Количество введенных строк для таблиц не совпадает.\n")
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": "Количество введенных строк для таблиц не совпадает.",
                })
				return
            }

            response := make(map[int]string)
            for i:=0; i < len(paramMap); i++ {
                query, paramValues := utils.ConvertMapToQueryAndParams(paramConnectedMap[i], reqBody.TableName, "insert_one_connected", nil)
                results, err := utils.ExecQueryAndParceResult(dbConn, query, paramValues...)
                if (err != ""){
                    fmt.Fprintf(os.Stderr, "%s\n", err)
                    response[i] = err;
                    continue
                }
                
                paramMap[i][reqBody.ConnectedField] = fmt.Sprintf("%d", results[0]["id"])

                query, paramValues = utils.ConvertMapToQueryAndParams(paramMap[i], table, "insert_one", nil)
                results, err = utils.ExecQueryAndParceResult(dbConn, query, paramValues...)
                if (err != ""){
                    fmt.Fprintf(os.Stderr, "%s\n", err)
                    response[i] = err;
                    continue
                }
                response[i] = "OK"
            }

            c.JSON(http.StatusOK, response)
            
        default:
            fmt.Fprint(os.Stderr, "Такой команды не существует!\n")
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "error":   "error",
                "message": "Такой команды не существует!\n",
            })
			return
    }
}

func InsertManyGetParamsAndValues(reqBody *utils.RequestInputManyBody, columns []string) ([]map[string]string, []map[string]string, error){
	PARAMETER_BYTES := 20
	PARAMETER_VALUE_BYTES := 60
	var paramValues []map[string]string
	var paramConnectedValues []map[string]string

    for _, input := range reqBody.InputMain{
        paramsMap := make(map[string]string)
        for _, f := range input {
            paramName, err := utils.ReadString(f.Key, PARAMETER_BYTES,"[^a-z_]")
            if (err != nil){
                return nil, nil, err
            }
            paramValue, err := utils.ReadString(f.Value, PARAMETER_VALUE_BYTES,"")
            if (err != nil){
                return nil, nil, err
            }
            paramsMap[paramName] = paramValue

            if(!slices.Contains(columns, paramName)){
                return nil, nil, fmt.Errorf("Такого столбца в таблице нет. Есть только %s\n", columns)
            }
        }
        paramValues = append(paramValues, paramsMap)
    }

    for _, input := range reqBody.InputConnected {
        paramsMap := make(map[string]string)
        for _, f := range input {
            paramName, err := utils.ReadString(f.Key, PARAMETER_BYTES,"[^a-z_]")
            if (err != nil){
                return nil, nil, err
            }
            paramValue, err := utils.ReadString(f.Value, PARAMETER_VALUE_BYTES,"")
            if (err != nil){
                return nil, nil, err
            }
            paramsMap[paramName] = paramValue

            if(!slices.Contains(columns, paramName)){
                return nil, nil, fmt.Errorf("Такого столбца в таблице нет. Есть только %s\n", columns)
            }
        }
        paramConnectedValues = append(paramConnectedValues, paramsMap)
	}

	return paramValues, paramConnectedValues, nil
}