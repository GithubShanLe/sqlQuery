package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func handleScheduleTask(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/tcc_cloud_apiserver")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM schedule_task")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	result := make(map[string]string)
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		for i, col := range values {
			if col == nil {
				result[columns[i]] = ""
			} else {
				result[columns[i]] = string(col)
			}
		}
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func main() {
	http.HandleFunc("/json/schedule_task", handleScheduleTask)
	http.HandleFunc("/table/", handleScheduleTaskTable)
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func handleScheduleTaskTable(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/tcc_cloud_apiserver")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	//格式必须http://127.0.0.1:9000/table/t_resource_info?resource_type=k8s
	dbTableName := strings.Split(strings.Split(r.RequestURI, "?")[0], "/")
	params := r.URL.Query().Encode()
	dbParams, err := ParaseParams(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	querySql := fmt.Sprintf("SELECT * FROM %s", dbTableName[2]+" where 1=1")
	for k, v := range dbParams {
		querySql += fmt.Sprintf(` AND %s = "%s"`, k, v)
	}
	if strings.Contains(querySql, ";") {
		http.Error(w, fmt.Sprintf("%s不允许", ";"), http.StatusInternalServerError)
		return
	}
	rows, err := db.Query(querySql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var data [][]string
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var row []string
		for _, col := range values {
			if col == nil {
				row = append(row, "")
			} else {
				row = append(row, string(col))
			}
		}
		data = append(data, row)
	}

	table := "<table border='1'><thead><tr>"
	for _, col := range columns {
		table += fmt.Sprintf("<th>%s</th>", col)
	}
	table += "</tr></thead><tbody>"
	for _, row := range data {
		table += "<tr>"
		for _, col := range row {
			table += fmt.Sprintf("<td>%s</td>", col)
		}
		table += "</tr>"
	}
	table += "</tbody></table>"

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, table)
}

func ParaseParams(params string) (result map[string]string, err error) {
	result = make(map[string]string)
	for _, item := range strings.Split(params, "&") {
		p := strings.Split(item, "=")
		fmt.Println(len(p), p, item)
		if len(p) != 2 {
			return nil, fmt.Errorf("param format is a=b&c=d")
		}
		k, v := strings.TrimSpace(p[0]), strings.TrimSpace(p[1])
		if _, exit := result[k]; !exit {
			result[k] = v
		}
	}
	return
}
