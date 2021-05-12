package utils

import (
	"bufio"
	"os"
	"strings"
)

var propertyMap = make(map[string]string)

func ReadProperties() {
	properties, _ := os.Open("E:\\discord_bot\\resources\\bot.properties")

	scanner := bufio.NewScanner(properties)

	for scanner.Scan() {
		line := scanner.Text()

		//ignore comments
		if line[0] != '#' {
			keyValuePair := strings.Split(line, "=")

			propertyMap[keyValuePair[0]] = keyValuePair[1]
		}
	}
}

func GetProperty(name string) string {
	return propertyMap[name]
}
