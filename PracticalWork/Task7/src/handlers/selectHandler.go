package handlers
import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"slices"

	"github.com/gin-gonic/gin"

	"connectToDB/utils"
)

func SelectGetHandler(c *gin.Context) {
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
    c.HTML(http.StatusOK, "select.html", gin.H{"Table": table})
}

func SelectPostHandler(c *gin.Context) {
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
    
    var reqBody utils.RequestSelectBody

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
            results, err := utils.ExecQueryAndParceResult(dbConn, fmt.Sprintf("SELECT * FROM shop.%s", table))
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
            

            paramMap, error := SelectGetParamsAndValues(&reqBody, columns)
			if (error != nil){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": error.Error(),
                })
				return
            }

            if (len(paramMap[0]) > 1){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": "Неверное количество значений",
                })
				return
            }
            query, paramValues := utils.ConvertMapToQueryAndParams(paramMap[0], table, "select", make(map[string][]int))
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
        case 3:

            paramMap, error := SelectGetParamsAndValues(&reqBody, columns)
			if (error != nil){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": error.Error(),
                })
				return
            }

            query, paramValues := utils.ConvertMapToQueryAndParams(paramMap[0], table, "select", make(map[string][]int))
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
        default:
            fmt.Fprint(os.Stderr, "Такой команды не существует!\n")
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "error":   "error",
                "message": "Такой команды не существует!\n",
            })
			return
    }
}

func SelectGetParamsAndValues(reqBody *utils.RequestSelectBody, columns []string) ([]map[string]string, error){
	PARAMETER_BYTES := 20
	PARAMETER_VALUE_BYTES := 60
	paramsMap := make(map[string]string)
	var paramValues []map[string]string

	for _, f := range reqBody.Filters {
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