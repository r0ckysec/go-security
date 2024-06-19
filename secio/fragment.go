/**
 * @Description
 * @Author r0cky
 * @Date 2022/1/7 16:26
 **/
package secio

import (
	"bytes"
	"fmt"
	"github.com/r0ckysec/go-security/misc"
	"github.com/thinkeridea/go-extend/exstrings"
	"os"
)

func Fragment(fileName string, shards int) (ret []string) {
	lines := misc.ReadLineAll(fileName)
	//if len(lines) <= shards {
	//	ret = append(ret, fileName)
	//	return
	//} else {
	index := 0
	nameTag := 1
	for {
		end := shards + index
		if end > len(lines) {
			end = len(lines)
		}
		shardsLine := lines[index:end]
		fileNameTemp := fmt.Sprintf("%s.fragment%d", fileName, nameTag)
		_ = misc.WriteLine(fileNameTemp, exstrings.Bytes(exstrings.Join(shardsLine, "\n")))
		ret = append(ret, fileNameTemp)
		if end == len(lines) {
			break
		}
		nameTag++
		index = end
	}
	return
	//}
}

func MergeFragment(files []string, out string) (err error) {
	ss := make([]string, 0, 4096)
	for _, file := range files {
		lineAll := misc.ReadLineAll(file)
		ss = append(ss, lineAll...)
	}
	ss = misc.RemoveDuplicatesAndEmpty(ss)
	buffer := bytes.Buffer{}
	defer buffer.Reset()
	buffer.WriteString(exstrings.Join(ss, "\n"))
	err = misc.WriteLine(out, buffer.Bytes())
	return
}

func ClearFragment(fileName []string) {
	if len(fileName) > 0 {
		for _, s := range fileName {
			_ = os.RemoveAll(s)
		}
	}
}
