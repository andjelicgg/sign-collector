package helper

import (
	"encoding/csv"
	"fmt"
	"github.com/ubavic/bas-celik/widgets"
	"os"
	"os/exec"
	"time"
)

var statusBar *widgets.StatusBar

type Form map[string]interface{}

func AppendCSV(form Form) error {
	// Get the current date in dd-mm-yyyy format
	date := time.Now().Format("02-01-2006")
	fileName := date + ".csv"

	// Check if the file exists
	fileExists := false
	if _, err := os.Stat(fileName); err == nil {
		fileExists = true
	}

	// Open the file in append mode or create it if it doesn't exist
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a CSV writer
	w := csv.NewWriter(file)
	defer w.Flush()

	// Write the header if the file is new
	if !fileExists {
		header := make([]string, 0, len(form))
		for key := range form {
			header = append(header, key)
		}
		if err := w.Write(header); err != nil {
			return err
		}
	}

	// Write the form data
	record := make([]string, 0, len(form))
	for _, value := range form {
		// Convert the value to a string type
		strValue, ok := value.(string)
		if !ok {
			// If the value is not a string, use fmt.Sprint to convert it to a string
			strValue = fmt.Sprint(value)
		}
		record = append(record, strValue)
	}
	if err := w.Write(record); err != nil {
		// Here you should return the error after setting the status
		// gui.SetStatus should be a valid call if gui is properly imported and SetStatus is exported
		//gui.SetStatus("Error appending to CSV", err)
		return err
	}
	w.Flush()

	return nil
}

func PrintPDF(file string, pdfType string) {
	cmd := exec.Command("./print_doc", file)

	err := cmd.Start() // Use Start() instead of Run() if you want non-blocking execution
	if err != nil {
		fmt.Printf("Error starting command: %s\n", err)
		os.Exit(1)
	}

	err = cmd.Wait() // Wait for the command to finish
	if err != nil {
		fmt.Printf("Command finished with error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Document sent to the " + pdfType)
}
