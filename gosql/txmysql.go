package gosql

import (
	"database/sql"
	"regexp"
	"strconv"
)

//MySQLTx 事务操作
type MySQLTx struct {
	tx *sql.Tx
}

//RowsCallbackResult 查询多条数据,结果以回调函数处理
//
//@param sql string SQL
//
//@param callback func(*sql.Rows) 回调函数指针
//
//@param args... interface{} SQL参数
func (m *MySQLTx) RowsCallbackResult(sql string, callback RowsCallback, args ...interface{}) Error {
	rows, err := m.tx.Query(sql, args...)
	if err != nil {
		return formatMySQLError(err)
	}
	defer rows.Close()
	if callback != nil {
		callback(rows)
	}
	return nil
}

//Rows 查询多条数据,结果以[]map[string]interface{}方式返回
//
//返回结果,使用本package中的类型函数进行数据解析
//eg:
//		result := Rows(...)
//		for _,mp := range result {
//			Int(mp["colum"])
//			String(mp["colum"])
//			.......
//		}
//@param sql string SQL
//
//@param args... interface{} SQL参数
func (m *MySQLTx) Rows(sql string, args ...interface{}) ([]map[string]interface{}, Error) {
	rows, err := m.tx.Query(sql, args...)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	defer rows.Close()
	var result []map[string]interface{}
	colums, _ := rows.Columns()
	for rows.Next() {
		var colmap = make(map[string]interface{}, 1)
		var refs = make([]interface{}, len(colums))
		for index, c := range colums {
			var ref interface{}
			colmap[c] = &ref
			refs[index] = &ref
		}
		err = rows.Scan(refs...)
		if err != nil {
			return nil, formatMySQLError(err)
		}
		for k, v := range colmap {
			colmap[k] = *v.(*interface{})
		}
		result = append(result, colmap)
	}
	return result, nil
}

//Row 查询单条语句,返回结果
//@param sql string SQL
//@param args... interface{} SQL参数
func (m *MySQLTx) Row(sql string, args ...interface{}) (*sql.Row, Error) {
	if ok, _ := regexp.MatchString("(?i)(.*?) LIMIT (.*?)\\s?(.*)?", sql); ok {
		sql = regexp.MustCompile("(?i)(.*?) LIMIT (.*?)\\s?(.*)?").ReplaceAllString(sql, "$1")
	} else {
		sql += " LIMIT 1 "
	}
	return m.tx.QueryRow(sql, args...), nil
}

//Exec 执行一条SQL
//@param sql string SQL
//@param args... interface{} SQL参数
func (m *MySQLTx) Exec(sql string, args ...interface{}) (sql.Result, Error) {
	result, err := m.tx.Exec(sql, args...)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	return result, nil
}

//Count SQL语句条数统计
//@param sql string SQL
//@param args... interface{} SQL参数
func (m *MySQLTx) Count(sql string, args ...interface{}) (int64, Error) {
	if ok, _ := regexp.MatchString("(?i)(.*?) LIMIT (.*?)\\s?(.*)?", sql); ok {
		sql = "SELECT COUNT(1) FROM (" + sql + ") as tmp"
	}
	if ok, _ := regexp.MatchString("(?i).* group by .*", sql); ok {
		sql = "SELECT COUNT(1) FROM (" + sql + ") as tmp"
	}
	sql = regexp.MustCompile("^(?i)select .*? from (.*) order by (.*)").ReplaceAllString(sql, "SELECT count(1) FROM $1")
	sql = regexp.MustCompile("^(?i)select .*? from (.*)").ReplaceAllString(sql, "SELECT count(1) FROM $1")
	result := m.tx.QueryRow(sql, args...)
	var count int64
	err := result.Scan(&count)
	if err != nil {
		return 0, formatMySQLError(err)
	}
	return count, nil
}

//GetTx 获取事务对象
func (m *MySQLTx) GetTx() *sql.Tx {
	return m.tx
}

//RowsPage 分页查询
func (m *MySQLTx) RowsPage(sql string, page *PageObj, args ...interface{}) ([]map[string]interface{}, Error) {
	if page == nil {
		return m.Rows(sql, args...)
	}
	countsql := "select count(0) from (" + sql + ") as total"
	result := m.tx.QueryRow(countsql, args...)
	var count int64
	err := result.Scan(&count)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	if page == nil {
		page = &PageObj{
			Page: 1,
			Rows: 1000,
		}
	}
	page.setTotal(count)
	currentpage := 0
	if page.Page-1 > 0 {
		currentpage = page.Page - 1
	}
	sql = sql + " LIMIT " + strconv.FormatInt(int64(currentpage*page.Rows), 10) + "," + strconv.FormatInt(int64(page.Rows), 10)
	return m.Rows(sql, args...)
}
