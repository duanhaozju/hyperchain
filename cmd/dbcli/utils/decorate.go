package utils

import "strings"

//Decorate -- decorate json result with color when print in terminal
func Decorate(src string) string {
	desc := src
	desc = strings.Replace(desc, "[", "\033[30;1m[\033[0m", -1)
	desc = strings.Replace(desc, "]", "\033[30;1m]\033[0m", -1)
	desc = strings.Replace(desc, "{", "\033[30;1m{\033[0m", -1)
	desc = strings.Replace(desc, "}", "\033[30;1m}\033[0m", -1)
	desc = strings.Replace(desc, "\"", "\033[34;1m\"", -1)
	desc = strings.Replace(desc, "\033[34;1m\":", "\"\033[0m:", -1)
	desc = strings.Replace(desc, ": \033[34;1m\"", ": \033[32m\"", -1)
	desc = strings.Replace(desc, "\033[34;1m\",", "\"\033[0m,", -1)
	desc = strings.Replace(desc, "\033[34;1m\"\n", "\"\033[0m\n", -1)
	desc = strings.Replace(desc, ":", "\033[30;1m:\033[0m", -1)
	return desc
}
