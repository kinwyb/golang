package gosql

import (
	"bytes"
	"reflect"
)

//SetSQL 转换成插入语句
func SetSQL(obj interface{}) (string, []interface{}) {
	vp := reflect.ValueOf(obj)
	if vp.CanInterface() {
		vp = vp.Elem()
	}
	retinterface := make([]interface{}, 0)
	buf := &bytes.Buffer{}
	rtype := reflect.TypeOf(vp.Interface())
	tp, _ := dbTag(rtype)
	if tp != nil {
		for k, v := range tp {
			d := vp.FieldByName(k)
			if d.IsNil() {
				continue
			} else {
				buf.WriteString(v)
				buf.WriteString(" = ?,")
				if d.CanInterface() {
					retinterface = append(retinterface, d.Elem().Interface())
				} else {
					retinterface = append(retinterface, d.Interface())
				}
			}
		}
	}
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 1)
	}
	return buf.String(), retinterface
}

//Update 更新数据
func Update(table string, obj interface{}) (string, []interface{}, Error) {
	vp := reflect.ValueOf(obj)
	if vp.CanInterface() {
		vp = vp.Elem()
	}
	retinterface := make([]interface{}, 0)
	rtype := reflect.TypeOf(vp.Interface())
	tp, primary := dbTag(rtype)
	if primary == "" {
		return "", nil, &err{
			code: -1,
			msg:  "更新数据无主键",
		}
	}
	buf := bytes.NewBufferString("UPDATE ")
	buf.WriteString(table)
	buf.WriteString(" SET ")
	var paramkey interface{}
	if tp != nil {
		for k, v := range tp {
			d := vp.FieldByName(k)
			if d.IsNil() {
				continue
			} else if primary == v {
				paramkey = getPtrInterface(d)
			} else {
				buf.WriteString(v)
				buf.WriteString(" = ?,")
				if d.CanInterface() {
					retinterface = append(retinterface, d.Elem().Interface())
				} else {
					retinterface = append(retinterface, d.Interface())
				}
			}
		}
	}
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 1)
	}
	buf.WriteString(" WHERE ")
	buf.WriteString(primary)
	buf.WriteString(" = ?")
	retinterface = append(retinterface, paramkey)
	return buf.String(), retinterface, nil
}

//SelectSQL 获取对象
func SelectSQL(obj interface{}, tablename ...string) string {
	vp := reflect.ValueOf(obj)
	if vp.CanInterface() {
		vp = vp.Elem()
	}
	buf := &bytes.Buffer{}
	if tablename != nil && len(tablename) > 0 {
		buf.WriteString("SELECT ")
	}
	rtype := reflect.TypeOf(vp.Interface())
	tp, _ := dbTag(rtype)
	if tp != nil {
		for k, v := range tp {
			buf.WriteString(v)
			buf.WriteString(" ")
			buf.WriteString(k)
			buf.WriteString(",")
		}
	}
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 1)
	}
	if tablename != nil && len(tablename) > 0 {
		buf.WriteString(" FROM ")
		buf.WriteString(tablename[0])
	}
	return buf.String()
}
