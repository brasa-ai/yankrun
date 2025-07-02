package helpers

import "fmt"

func ErrorHandler(err error, customMessage ...string) {
	if err != nil {
		message := ""
		if len(customMessage) > 0 {
			message = customMessage[0]
		}
		if message != "" {
			panic(fmt.Sprintf("%s: %v", message, err))
		}
		panic(err)
	}
}
