package utils

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type operation int

const (
	defaultOperation operation = iota
	ifOperation
	forOperation
	includeOperation
)

type tpNode struct {
	opt       operation
	value1    *tpNode
	value2    *tpNode
	condition string
	content   string
	next      *tpNode //下一项
	prev      *tpNode //上一项
	parent    *tpNode //子项目
}

//返回内容值
func (t *tpNode) value(data interface{}) string {
	buf := &bytes.Buffer{}
	if t.opt == defaultOperation {
		buf.WriteString(t.content)
	} else {
		switch t.opt {
		case ifOperation:
			buf.WriteString(iFCase(t, data))
		}
	}
	if t.next != nil {
		buf.WriteString(t.next.value(data))
	}
	return buf.String()
}

// html与逻辑语法分离
// forEach(source.split(openTag), function (code) {
//     code = code.split(closeTag);
//     var $0 = code[0];
//     var $1 = code[1];
//     if (code.length === 1) {
//         mainCode += html($0);
//     } else {
//         mainCode += logic($0);
//         if ($1) {
//             mainCode += html($1);
//         }
//     }
// });

//分离语法
func separateHTML(source string) *tpNode {
	strs := strings.Split(source, "{{")
	nts := &tpNode{}
	topnode := nts
	for _, s := range strs {
		code := strings.Split(s, "}}")
		if len(code) == 1 {
			nts.opt = defaultOperation
			nts.content = code[0]
			nts.next = &tpNode{
				prev: nts,
			}
			nts = nts.next
		} else {
			cd0 := strings.Replace(code[0], `/^\s/`, "", -1)
			arr := strings.Split(cd0, " ")
			key := ""
			args := ""
			for index, v := range arr {
				if index == 0 {
					key = v
				} else {
					args += v + " "
				}
			}
			switch key {
			case "if":
				nts.opt = ifOperation
				nts.condition = args
				nts.value1 = &tpNode{
					content: code[1],
					opt:     defaultOperation,
					parent:  nts,
				}
				nts = nts.value1
			case "else":
				for nts.parent != nil {
					nts = nts.parent
					if nts.opt == ifOperation {
						tmpNode := &tpNode{
							content: code[1],
							opt:     defaultOperation,
							parent:  nts,
						}
						if len(arr) > 1 && arr[1] == "if" {
							nts = nts.value1
						} else {
							nts = nts.value2
						}
						nts = tmpNode
						break
					}
				}
			case "/if":
				for nts.parent != nil {
					nts = nts.parent
					if nts.opt == ifOperation {
						nts.next = &tpNode{
							prev: nts,
						}
						nts = nts.next
						break
					}
				}
			case "each":
			case "/each":
			case "echo":
			case "print":
			case "include":
			default:
				nts.opt = defaultOperation
				nts.content = "{{" + s
				nts.next = &tpNode{
					prev: nts,
				}
				nts = nts.next
			}
		}
	}
	return topnode
}

//获取对象值
func getValue(source string, data interface{}) interface{} {
	defer func() { //必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			fmt.Println(err) //这里的err其实就是panic传入的内容，55
		}
	}()
	sts := strings.SplitN(source, ".", 2)
	t := reflect.TypeOf(data)
	if t.Kind() == reflect.Map {
		v := reflect.ValueOf(data)
		result := v.MapIndex(reflect.ValueOf(sts[0]))
		if result.IsValid() {
			if len(sts) > 1 {
				return getValue(sts[1], result.Interface())
			}
			return result.Interface()
		}
		return nil
	}
	v := reflect.ValueOf(data)
	res := v.FieldByName(sts[0])
	if res.IsValid() {
		if len(sts) > 1 {
			return getValue(sts[1], res.Interface())
		}
		return res.Interface()
	}
	return nil
}

// 处理逻辑语句
//     function logic(code) {
//         var thisLine = line;
//         if (parser) {
//             // 语法转换插件钩子
//             code = parser(code, options);
//         } else if (debug) {
//             // 记录行号
//             code = code.replace(/\n/g, function () {
//                 line++;
//                 return "$line=" + line + ";";
//             });
//         }
//         // 输出语句. 编码: <%=value%> 不编码:<%=#value%>
//         // <%=#value%> 等同 v2.0.3 之前的 <%==value%>
//         if (code.indexOf('=') === 0) {
//             var escapeSyntax = escape && !/^=[=#]/.test(code);
//             code = code.replace(/^=[=#]?|[\s;]*$/g, '');
//             // 对内容编码
//             if (escapeSyntax) {
//                 var name = code.replace(/\s*\([^\)]+\)/, '');
//                 // 排除 utils.* | include | print
//                 if (!utils[name] && !/^(include|print)$/.test(name)) {
//                     code = "$escape(" + code + ")";
//                 }
//                 // 不编码
//             } else {
//                 code = "$string(" + code + ")";
//             }
//             code = replaces[1] + code + replaces[2];
//         }

//         if (debug) {
//             code = "$line=" + thisLine + ";" + code;
//         }

//         // 提取模板中的变量名
//         forEach(getVariable(code), function (name) {
//             // name 值可能为空，在安卓低版本浏览器下
//             if (!name || uniq[name]) {
//                 return;
//             }
//             var value;
//             // 声明模板变量
//             // 赋值优先级:
//             // [include, print] > utils > helpers > data
//             if (name === 'print') {
//                 value = print;
//             } else if (name === 'include') {
//                 value = include;
//             } else if (utils[name]) {
//                 value = "$utils." + name;
//             } else if (helpers[name]) {
//                 value = "$helpers." + name;
//             } else {
//                 value = "$data." + name;
//             }
//             headerCode += name + "=" + value + ",";
//             uniq[name] = true;
//         });
//         return code + "\n";
//     }
// };

func iFCase(node *tpNode, data interface{}) string {
	println(node.condition)
	return ""
}
