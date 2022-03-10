package misc

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"github.com/thinkeridea/go-extend/exbytes"
	"github.com/thinkeridea/go-extend/exstrings"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

// StrArr2IntArr string数组转化int数组
func StrArr2IntArr(strArr []string) ([]int, error) {
	var intArr []int
	for _, value := range strArr {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		intArr = append(intArr, intValue)
	}
	return intArr, nil
}

// Str2Int string转int
func Str2Int(str string) int {
	intValue, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return intValue
}

// IntArr2StrArr int数组转string数组
func IntArr2StrArr(intArr []int) []string {
	var strArr []string
	for _, value := range intArr {
		strValue := strconv.Itoa(value)
		strArr = append(strArr, strValue)
	}
	return strArr
}

// Int2Str int转string
func Int2Str(Int int) string {
	return strconv.Itoa(Int)
}

// IsInStrArr 是否在string数组中
func IsInStrArr(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// IsInIntArr 是否在int数组中
func IsInIntArr(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// WriteLine 写入文件
func WriteLine(fileName string, byte []byte) error {
	//file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	_ = os.RemoveAll(fileName)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	//创建成功挂起关闭文件流,在函数结束前执行
	defer file.Close()
	//NewWriter创建一个以目标文件具有默认大小缓冲、写入w的*Writer。
	writer := bufio.NewWriter(file)
	//写入器将内容写入缓冲。返回写入的字节数。
	_, err = writer.Write(byte)
	//Flush方法将缓冲中的数据写入下层的io.Writer接口。缺少，数据将保留在缓冲区，并未写入io.Writer接口
	_ = writer.Flush()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return err
}

// ReadLine 读取行
func ReadLine(fileName string, handler func(string, bool)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		//line = FixLine(line)
		handler(line, true)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// ReadLineAll 读取文件返回字符串数组
func ReadLineAll(fileName string) []string {
	var strArr []string
	f, err := os.Open(fileName)
	if err != nil {
		return strArr
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	defer buf.Reset(buf)
	for {
		line, err := buf.ReadBytes('\n')
		line = FixLineBytes(line)
		strArr = append(strArr, exbytes.ToString(line))
		if err != nil {
			if err == io.EOF {
				return strArr
			}
			return strArr
		}
	}
}

// ReadLineStr 读取文件返回字符串
func ReadLineStr(fileName string) string {
	var str = new(strings.Builder)
	defer str.Reset()
	f, err := os.Open(fileName)
	if err != nil {
		return str.String()
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				str.Write(line)
			}
			break
		}
		str.Write(line)
	}
	return str.String()
}

// FixSpace 自定义空格数去空格
func FixSpace(line string, len int) string {
	switch len {
	case 0:
		line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte(" "), []byte(""), -1))
		return line
	default:
		var re, _ = regexp.Compile("\\s{" + Int2Str(len+1) + ",}")
		line = re.ReplaceAllString(line, " ")
		line = strings.TrimSpace(line)
		return line
	}
}

// FixLine 行去换行、制表符（不去空格）
func FixLine(line string) string {
	//line = strings.Replace(line, "\r", "", -1)
	////line = strings.Replace(line, " ", "", -1)
	//line = strings.Replace(line, "\t", "", -1)
	//line = strings.Replace(line, "\r", "", -1)
	//line = strings.Replace(line, "\n", "", -1)
	//line = strings.Replace(line, "\xc2\xa0", "", -1)
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\r"), []byte(""), -1))
	//line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte(" "), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\t"), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\n"), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\xc2\xa0"), []byte(""), -1))
	return line
}

// FixLineSpace 行去换行、制表符、去空格
func FixLineSpace(line string) string {
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\r"), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte(" "), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\t"), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\n"), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\xc2\xa0"), []byte(""), -1))
	return line
}

// FixLineBytes 字节去换行、制表符、空格
func FixLineBytes(bytes []byte) []byte {
	bytes = exbytes.Replace(bytes, []byte("\r"), []byte(""), -1)
	bytes = exbytes.Replace(bytes, []byte(" "), []byte(""), -1)
	bytes = exbytes.Replace(bytes, []byte("\t"), []byte(""), -1)
	bytes = exbytes.Replace(bytes, []byte("\n"), []byte(""), -1)
	bytes = exbytes.Replace(bytes, []byte("\xc2\xa0"), []byte(""), -1)
	return bytes
}

// FixLineNotWrap 去制表符、空格（不去换行）
func FixLineNotWrap(line string) string {
	//line = strings.Replace(line, "\r", "", -1)
	//line = strings.Replace(line, " ", "", -1)
	//line = strings.Replace(line, "\t", "", -1)
	//line = strings.Replace(line, "\r", "", -1)
	//line = strings.Replace(line, "\xc2\xa0", "", -1)
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\r"), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte(" "), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\t"), []byte(""), -1))
	line = exbytes.ToString(exbytes.Replace(exstrings.UnsafeToBytes(line), []byte("\xc2\xa0"), []byte(""), -1))
	return line
}

// CheckIllegal 检测是否存在命令注入的特殊符号
func CheckIllegal(cmd string) bool {
	if strings.Contains(cmd, "&") || strings.Contains(cmd, "|") || strings.Contains(cmd, ";") ||
		strings.Contains(cmd, "$") || strings.Contains(cmd, "'") || strings.Contains(cmd, "`") ||
		strings.Contains(cmd, "(") || strings.Contains(cmd, ")") || strings.Contains(cmd, "\"") {
		return true
	}
	return false
}

// Str2ArrByWarp string转字符串数组
func Str2ArrByWarp(str string) []string {
	strNoWarp := FixLineNotWrap(str)
	arr := strings.Split(strNoWarp, "\n")
	arr = RemoveDuplicatesAndEmpty(arr)
	return arr
}

// RemoveDuplicatesAndEmpty 数组去重去空
func RemoveDuplicatesAndEmpty(ss []string) (ret []string) {
	result := make([]string, 0, len(ss))
	temp := map[string]struct{}{}
	for _, item := range ss {
		if _, ok := temp[item]; !ok && len(item) > 0 {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// FileIsExist 文件是否存在
func FileIsExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

// Base64Encode base64编码
func Base64Encode(keyword string) string {
	input := []byte(keyword)
	encodeString := base64.StdEncoding.EncodeToString(input)
	return encodeString
}

// Base64Decode base64解码
func Base64Decode(encodeString string) string {
	decodeBytes, err := base64.StdEncoding.DecodeString(encodeString)
	if err != nil {
		//global.Log.Println(err)
		return ""
	}
	return Bytes2Str(decodeBytes)
}

// CloneMap 克隆map
func CloneMap(strMap map[string]interface{}) map[string]interface{} {
	newStrMap := make(map[string]interface{})
	for k, v := range strMap {
		newStrMap[k] = v
	}
	return newStrMap
}

// CloneStrMap 克隆string map
func CloneStrMap(strMap map[string]string) map[string]string {
	newStrMap := make(map[string]string)
	for k, v := range strMap {
		newStrMap[k] = v
	}
	return newStrMap
}

// CloneIntMap 克隆int map
func CloneIntMap(intMap map[int]string) map[int]string {
	newIntMap := make(map[int]string)
	for k, v := range intMap {
		newIntMap[k] = v
	}
	return newIntMap
}

// ToMap map[string]string 转 map[string]interface{}
func ToMap(i map[string]string) map[string]interface{} {
	m := make(map[string]interface{}, len(i))
	for k, v := range i {
		m[k] = v
	}
	return m
}

// Str2Bytes 高效率str to bytes
func Str2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// Bytes2Str 高效率bytes to str
func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StrInArr 高效率判断字符串是否在数组中
func StrInArr(ss []string, s string) bool {
	index := sort.SearchStrings(ss, s)
	if index < len(ss) && ss[index] == s {
		return true
	}
	return false
}

// MergeSlice 合并字符串数组
func MergeSlice(s1 []string, s2 []string) []string {
	slice := make([]string, len(s1)+len(s2))
	copy(slice, s1)
	copy(slice[len(s1):], s2)
	return slice
}

// DoZlibCompress 压缩（字符串压缩）
func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	_, _ = w.Write(src)
	_ = w.Close()
	return in.Bytes()
}

// DoZlibUnCompress 解压（字符串解压）
func DoZlibUnCompress(compressSrc []byte) (unzip []byte, err error) {
	b := bytes.NewReader(compressSrc)
	var out bytes.Buffer
	r, err := zlib.NewReader(b)
	if err != nil {
		return
	}
	_, _ = io.Copy(&out, r)
	unzip = out.Bytes()
	return
}
