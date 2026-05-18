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

func InsertOneGetHandler(c *gin.Context) {
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
	

    c.HTML(http.StatusOK, "insertOne.html", gin.H{"Table": table, "Columns": columns, "NotNullColumns": notNullCollumns,"ConnectedTables": result})
}

func InsertOnePostHandler(c *gin.Context) {
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
    
    var reqBody utils.RequestInputOneBody

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
			paramMap, error := InsertOneGetParamsAndValues(&reqBody, columns)
			if (error != nil){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": error.Error(),
                })
				return
            }

            query, paramValues := utils.ConvertMapToQueryAndParams(paramMap[0], table, "insert_one", nil)
            results, err := utils.ExecQueryAndParceResult(dbConn, query, paramValues...)
            if (err != ""){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": err,
                })
				return
            }

            c.JSON(http.StatusOK, results)
        
        case 2:

            paramMap, error := InsertOneGetParamsAndValues(&reqBody, columns)
			if (error != nil){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": error.Error(),
                })
				return
            }

			query, paramValues := utils.ConvertMapToQueryAndParams(paramMap[1], reqBody.TableName, "insert_one_connected", nil)
			results, err := utils.ExecQueryAndParceResult(dbConn, query, paramValues...)
            if (err != ""){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": err,
                })
				return
            }
            
            paramMap[0][reqBody.ConnectedField] = fmt.Sprintf("%d", results[0]["id"])

            query, paramValues = utils.ConvertMapToQueryAndParams(paramMap[0], table, "insert_one", nil)
			results, err = utils.ExecQueryAndParceResult(dbConn, query, paramValues...)
            if (err != ""){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": err,
                })
				return
            }
            

            c.JSON(http.StatusOK, results)
            
        default:
            fmt.Fprint(os.Stderr, "Такой команды не существует!\n")
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "error":   "error",
                "message": "Такой команды не существует!\n",
            })
			return
    }
}

func InsertOneGetParamsAndValues(reqBody *utils.RequestInputOneBody, columns []string) ([]map[string]string, error){
	PARAMETER_BYTES := 20
	PARAMETER_VALUE_BYTES := 60
	paramsMap := make(map[string]string)
	var paramValues []map[string]string

	for _, f := range reqBody.InputMain {
		paramName, err := utils.ReadString(f.Key, PARAMETER_BYTES,"[^a-z_]")
		if (err != nil){
			return nil, err
		}
		paramValue, err := utils.ReadString(f.Value, PARAMETER_VALUE_BYTES,"")
		if (err != nil){
			return nil, err
		}
		paramsMap[paramName] = paramValue

		if(!slices.Contains(columns, paramName)){
			return nil, fmt.Errorf("Такого столбца в таблице нет. Есть только %s\n", columns)
		}
	}
	paramValues = append(paramValues, paramsMap)

    paramsMap = make(map[string]string)
    for _, f := range reqBody.InputConnected {
		paramName, err := utils.ReadString(f.Key, PARAMETER_BYTES,"[^a-z_]")
		if (err != nil){
			return nil, err
		}
		paramValue, err := utils.ReadString(f.Value, PARAMETER_VALUE_BYTES,"")
		if (err != nil){
			return nil, err
		}
		paramsMap[paramName] = paramValue

		if(!slices.Contains(columns, paramName)){
			return nil, fmt.Errorf("Такого столбца в таблице нет. Есть только %s\n", columns)
		}
	}
    paramValues = append(paramValues, paramsMap)

	return paramValues, nil
}