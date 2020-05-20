package file_operations

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
)

func checkFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func CheckFuncExists(filePath string, listOfArgTypes []string) (bool, string) {
	var funcExists = false
	var funcName = ""
	if !checkFileExists(filePath) {
		return false, funcName
	}

	for j, str := range listOfArgTypes {
		lengthOfStr := len(str)
		for i := 0; i < lengthOfStr; i++ {
			if str[i] == '(' {
				str = str[:i] + "\\" + str[i:]
				i += 2
			}
			if str[i] == ')' {
				str = str[:i] + "\\" + str[i:]
				i += 2
			}
			if str[i] == '[' {
				str = str[:i] + "\\" + str[i:]
				i += 2
			}
			if str[i] == ']' {
				str = str[:i] + "\\" + str[i:]
				i += 2
			}
		}
		listOfArgTypes[j] = str
	}

	var target = ""
	length := len(listOfArgTypes)
	i := 0
	for ; i < length-2; i++ {
		target = fmt.Sprintf("%s argname_%d %s,", target, i+1, listOfArgTypes[i])
	}
	target = fmt.Sprintf("%s argname_%d %s", target, i+1, listOfArgTypes[i])
	target = fmt.Sprintf("%s\\) %s", target, listOfArgTypes[length-1])

	fmt.Printf("Finding %s in %s...\n", target, filePath)
	funcExists, funcName = matchFunc(filePath, target)

	return funcExists, funcName
}

func matchFunc(filePath, origin string) (bool, string) {
	var funcName = ""
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
		return false, funcName
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return false, funcName
			}
			panic(err)
			return false, funcName
		}

		if ok, _ := regexp.Match(origin, line); ok {
			fmt.Println("Function has been generated before!")
			funcName = getFuncNameFromLine(line)
			fmt.Println("Previous function name:", funcName)
			return true, funcName
		}
	}
}

func getFuncNameFromLine(line []byte) string {
	// line is like "func AddAB( argname_1 int, argname_2 int) int {"
	// then this func will match funcName which like "AddAB" in line
	expr := "func \\w+\\(" // regular expression
	reg, _ := regexp.Compile(expr)
	// matchRet is the result of regular expression match, it will like "func AddAB("
	matchRet := string(reg.Find(line))
	// funcName is like "AddAB"
	funcName := matchRet[5 : len(matchRet)-1]
	return funcName
}

func ensureFileExists(filePath string) (*os.File, error, bool) {
	var f *os.File
	var err error
	var exist = false
	if checkFileExists(filePath) {
		exist = true
		f, err = os.OpenFile(filePath, os.O_APPEND, 0666)
	} else {
		f, err = os.Create(filePath)
	}

	if err != nil {
		panic(err)
		return nil, err, exist
	}

	return f, err, exist
}

func WriteFuncToFile(filePath, packageName string, input []byte) error {
	f, err, exist := ensureFileExists(filePath)
	defer f.Close()
	if err != nil {
		panic(err)
		return err
	}

	writer := bufio.NewWriter(f)
	if !exist {
		tmpStr := packageName + "\n"
		tmpBuffer := []byte(tmpStr)
		var buffer bytes.Buffer
		buffer.Write(tmpBuffer)
		buffer.Write(input)
		input = buffer.Bytes()
	}
	_, err = writer.Write(input)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}
