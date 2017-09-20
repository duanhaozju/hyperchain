package gen

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type ContractContent struct {
	ContractVariablesContent []string
	StructContent            map[string]string
	EnumContent              map[string]string
	MapContent               map[string]string
}

func (contractContent *ContractContent) String() string {
	return fmt.Sprintf("ContractVariablesContent: %v. StructContent: %v. EnumContent: %v. MapContent: %v.", contractContent.ContractVariablesContent, contractContent.StructContent, contractContent.EnumContent, contractContent.MapContent)
}

func standardizeAllContracts(path string) string {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	buf := bufio.NewReader(file)
	var res string
	for {
		str, err := buf.ReadByte()
		if err != nil {
			if err != io.EOF {
				fmt.Println(err.Error())
				return ""
			}
			break
		}
		if str == '/' {
			str, err := buf.ReadByte()
			if err != nil {
				if err != io.EOF {
					fmt.Println(err.Error())
					return ""
				}
				break
			}
			if str == '/' {
				for {
					str, err := buf.ReadByte()
					if err != nil {
						if err != io.EOF {
							fmt.Println(err.Error())
							return ""
						}
						break
					}
					if str == '\n' {
						break
					}
				}
			}
			continue
		}
		if str == '{' || str == '}' || str == ';' || str == '=' || str == ',' || str == '(' || str == ')' {
			res = res + " " + string(str) + " "
		} else {
			res += string(str)
		}
	}
	res = strings.TrimSpace(res)
	re, err := regexp.Compile("\\s{2,}")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	res = re.ReplaceAllString(res, " ")
	return res
}

func getContentOfContract(res string) (string, *ContractContent) {
	contractContent := &ContractContent{}
	contractContent.ContractVariablesContent = make([]string, 0)
	contractContent.StructContent = make(map[string]string)
	contractContent.EnumContent = make(map[string]string)
	contractContent.MapContent = make(map[string]string)

	content := strings.Split(strings.TrimSpace(res), " ")
	var i int
	var name string
	for i = 0; i < len(content); i++ {
		if content[i] == "{" {
			break
		}
		name += content[i]
	}
	i++
	for i < len(content) {
		if content[i] == "function" {
			count := 0
			for {
				i++
				if content[i] == "{" {
					count++
					break
				}
			}
			for {
				i++
				if content[i] == "{" {
					count++
				} else if content[i] == "}" {
					count--
				}
				if count == 0 {
					break
				}
			}
			i++
		} else if content[i] == "mapping" {
			var mapVariable string
			for {
				if content[i] == ";" {
					break
				} else if content[i] != " " {
					mapVariable += content[i]
					i++
				}
			}
			i++
			lasIndex := strings.LastIndex(mapVariable, ")")
			contractContent.ContractVariablesContent = append(contractContent.ContractVariablesContent, mapVariable[:lasIndex+1]+" "+mapVariable[lasIndex+1:])
			contractContent.MapContent[mapVariable[lasIndex+1:]] = mapVariable[:lasIndex+1]
		} else if content[i] == "struct" {
			i++
			key := content[i]
			var value string
			i = i + 2
			for {
				if content[i] != "}" {
					value += content[i] + " " + content[i+1] + " "
					i = i + 3
				} else if content[i] == "}" {
					break
				}
			}
			contractContent.StructContent[key] = value[:len(value)-1]
			i++
		} else if content[i] == "enum" {
			i++
			key := content[i]
			var value string
			i = i + 2
			for {
				if content[i] != "}" {
					value += content[i]
				} else if content[i] == "}" {
					break
				}
				i++
			}
			contractContent.EnumContent[key] = value
			i++
		} else if content[i] != "}" {
			var tmp string
			var constantFlag bool
			for {
				if content[i] == "constant" {
					constantFlag = true
				}
				if content[i] != "public" && content[i] != "private" && content[i] != "internal" {
					tmp += content[i] + " "
				}
				i++
				if content[i] == ";" {
					break
				}
			}
			i++
			if !constantFlag {
				contractContent.ContractVariablesContent = append(contractContent.ContractVariablesContent, tmp[:len(tmp)-1])
			}
		} else {
			i++
		}
	}
	return name, contractContent
}

func GetContentOfAllContracts(path string) map[string]*ContractContent {
	var contentOfAllContracts map[string]*ContractContent
	contentOfAllContracts = make(map[string]*ContractContent)
	res := standardizeAllContracts(path)
	temps := strings.Split(res, "contract ")
	for i := 0; i < len(temps); i++ {
		if len(temps[i]) != 0 {
			name, con := getContentOfContract(temps[i])
			contentOfAllContracts[name] = con
		}
	}
	return contentOfAllContracts
}
