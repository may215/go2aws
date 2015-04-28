package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Read the configuration data from the config files
// for the running environment.
func getConfigData(env string) error {

	fmt.Println("config file : " + "./config/" + env + ".conf")

	file, err := os.Open("./config/" + env + ".conf")
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var conf = strings.Join(lines, " ")
	b := []byte(conf)
	m_err := json.Unmarshal(b, &configuration)
	if m_err != nil {
		return m_err
	}

	return nil
}
