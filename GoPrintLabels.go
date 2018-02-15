package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

var logging bool

type Settings struct {
	FinalPrintCodeFolder   string
	InterimPrintCodeFolder string
}

var settingsFile string
var settings = Settings{}

func main() {

	var finalPath string
	var interimPath string

	if runtime.GOOS == "windows" {
		settingsFile = `C:\GoPrintLabels\settings.json`
		finalPath = filepath.Join("C:\\", "GoPrintLabels", "PrintCode", "Final")
		interimPath = filepath.Join("C:\\", "GoPrintLabels", "PrintCode", "Interim")
	} else {
		settingsFile = `./settings.json`
		finalPath = "./PrintCode/Final"
		interimPath = "./PrintCode/Interim"
	}

	_, err := os.Stat(finalPath)
	if err != nil {
		fmt.Println("Creating final print code storage location.")
		os.MkdirAll(finalPath, 0777)
	}

	_, err = os.Stat(interimPath)
	if err != nil {
		fmt.Println("Creating interim print code storage location.")
		os.MkdirAll(interimPath, 0777)
	}

	_, err = os.Stat(settingsFile)
	if err != nil {
		fmt.Println("creating default settings file.")
		settings.FinalPrintCodeFolder = finalPath
		settings.InterimPrintCodeFolder = interimPath
		updateSettings()
	}

	settingsData, err := ioutil.ReadFile(settingsFile)

	json.Unmarshal(settingsData, &settings)

	args := os.Args

	if len(args) > 1 {
		action := args[1]

		switch action {
		case "log":
			logging = true
			Listen()
		case "help":
			fmt.Println(helpFile)
		case "plcMapping":
			fmt.Println(plcMapping)
		case "string":
			fmt.Println(stringMapping)
		case "config":
			fmt.Println("")
			fmt.Println("Final label print code location: ", settings.FinalPrintCodeFolder)
			fmt.Println("Interim label print code location: ", settings.InterimPrintCodeFolder)
			fmt.Println("")
		case "final":
			if len(args) > 2 {
				setFinalFolder(args[2])
			} else {
				fmt.Println("")
				fmt.Println("Missing argument. Final print code folder name expected.")
				fmt.Println("usage: fl path/to/printcode")
				fmt.Println("You can get help by using help flag.")
				fmt.Println("")
			}
		case "interim":
			if len(args) > 2 {
				setInterimFolder(args[2])
			} else {
				fmt.Println("")
				fmt.Println("Missing argument. Iterim print code folder name expected.")
				fmt.Println("usage: fl path/to/printcode")
				fmt.Println("You can get help by using help flag.")
				fmt.Println("")
			}
		default:

		}

		return
	}

	Listen()

}

func updateSettings() {

	settingData, err := json.Marshal(&settings)
	if err != nil {
		fmt.Println("Could not create settings file.")
	}

	writeSettingsToFile(settingData)

}

func writeSettingsToFile(settingData []byte){

	err := ioutil.WriteFile(settingsFile, settingData, 0777)
	if err != nil {
		fmt.Println("could not save settings.json")
	}
}

func setFinalFolder(loc string) {

	_, err := os.Stat(loc)
	if err != nil {
		fmt.Println("Creating final print code storage location.")

		input := getUserInput("Folder does not exist. Create it? ")

		if len(input) > 0 {
			if input[:1] == "y" || input[:1] == "Y" {
				os.MkdirAll(loc, 0777)
				settings.FinalPrintCodeFolder = loc
				updateSettings()
			}
		}
		return
	}

	settings.FinalPrintCodeFolder = loc
	updateSettings()

}

func setInterimFolder(loc string) {
	_, err := os.Stat(loc)
	if err != nil {
		fmt.Println("Creating interim print code storage location.")

		input := getUserInput("Folder does not exist. Create it? ")

		if len(input) > 0 {
			if input[:1] == "y" || input[:1] == "Y" {
				os.MkdirAll(loc, 0777)
				settings.FinalPrintCodeFolder = loc
				updateSettings()
			}
		}
		return
	}
	settings.FinalPrintCodeFolder = loc
	updateSettings()
}

func getUserInput(prompt string) string {
	fmt.Println("")
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)

	return input
}
