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

func UpdateGetHandler(c *gin.Context) {
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
    c.HTML(http.StatusOK, "update.html", gin.H{"Table": table})
}

func UpdatePostHandler(c *gin.Context) {
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
    
    var reqBody utils.RequestUpdateBody

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
            id := reqBody.Ids[0]
			paramMap, error := UpdateGetParamsAndValues(&reqBody, columns)
			if (error != nil){
                fmt.Fprintf(os.Stderr, "%s\n", err)
                c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                    "error":   "error",
                    "message": error.Error(),
                })
				return
            }
            
			paramMap[0]["id"] = fmt.Sprintf("%d", id)
			query, paramValues := utils.ConvertMapToQueryAndParams(paramMap[0], table, "update_one", make(map[string][]int))
			results, err :=utils.ExecQueryAndParceResult(dbConn, query, paramValues...)
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
            paramMap, error := UpdateGetParamsAndValues(&reqBody, columns)
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

			query, paramValues := utils.ConvertMapToQueryAndParams(paramMap[0], table, "update_many", map[string][]int{"id":reqBody.Ids})
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

func UpdateGetParamsAndValues(reqBody *utils.RequestUpdateBody, columns []string) ([]map[string]string, error){
	PARAMETER_BYTES := 20
	PARAMETER_VALUE_BYTES := 60
	paramsMap := make(map[string]string)
	var paramValues []map[string]string

	for _, f := range reqBody.Changes {
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

        if(paramName == "id"){
            return nil, fmt.Errorf("Поле id изменять нельзя!\n")
        }
	}
	paramValues = append(paramValues, paramsMap)
	return paramValues, nil
}