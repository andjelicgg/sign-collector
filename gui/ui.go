package gui

import (
	"bitbucket.org/inceptionlib/pdfinject-go"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ebfe/scard"
	"github.com/ubavic/bas-celik/document"
	"github.com/ubavic/bas-celik/helper"
	"github.com/ubavic/bas-celik/widgets"
	"os"
	"path/filepath"
	"strings"
)

var statusBar *widgets.StatusBar
var window *fyne.Window
var verbose bool
var startPageOn bool // possible data races
var startPage *widgets.StartPage

func StartGui(ctx *scard.Context, verbose bool) {
	startPageOn = true

	appW := app.New()
	win := appW.NewWindow("DJB Potpisi")
	window = &win
	go pooler(ctx)

	appW.Settings().SetTheme(MyTheme{})

	statusBar = widgets.NewStatusBar()

	startPage = widgets.NewStartPage()
	startPage.SetStatus("", "", false)

	win.SetContent(
		container.New(
			layout.NewPaddedLayout(),
			startPage,
		),
	)
	win.ShowAndRun()
}

func formatName(firstName string, lastName string, middleName string) string {
	var st strings.Builder

	st.WriteString(firstName)
	st.WriteString(" ")
	if len(middleName) > 0 {
		st.WriteString(", " + middleName + ", ")
	}
	st.WriteString(lastName)

	return st.String()
}

func enableManualUI() {
	givenName := widget.NewEntry()
	givenName.SetPlaceHolder("Ime")

	givenSurname := widget.NewEntry()
	givenSurname.SetPlaceHolder("Prezime")

	parentGivenName := widget.NewEntry()
	parentGivenName.SetPlaceHolder("Srednje ime")

	personalNumber := widget.NewEntry()
	personalNumber.SetPlaceHolder("JMBG")

	city := widget.NewEntry()
	city.SetPlaceHolder("Grad")

	address := widget.NewEntry()
	address.SetPlaceHolder("Ulica")

	addressNo := widget.NewEntry()
	addressNo.SetPlaceHolder("Broj")

	DateOfBirth := widget.NewEntry()
	DateOfBirth.SetPlaceHolder("DateOfBirth")

	issuingAuthority := widget.NewEntry()
	issuingAuthority.SetPlaceHolder("Dokument izdao")

	issuingDate := widget.NewEntry()
	issuingDate.SetPlaceHolder("Datum izdavanja")

	documentNumber := widget.NewEntry()
	documentNumber.SetPlaceHolder("Broj dokumenta")

	authorizedCertifierName := widget.NewEntry()
	authorizedCertifierName.SetPlaceHolder("Ime i prezime overitelja")

	authorizedCertifierAddress := widget.NewEntry()
	authorizedCertifierAddress.SetPlaceHolder("Ulica i broj overitelja")

	place := widget.NewEntry()
	place.SetPlaceHolder("Lokacija")

	readCardButton := widget.NewButton("Procitaj karticu", func() {
		setStartPage("Čitam sa kartice...", "", nil)
	})
	submitButton := widget.NewButton("Submit", func() {
		executable, err := os.Executable() // Gets the path of the current executable.
		if err != nil {
			fmt.Println("Error getting executable path:", err)
		}

		fullName := formatName(givenName.Text, givenSurname.Text, parentGivenName.Text)

		form := map[string]interface{}{
			"field_politicalPartyName":         "Dosta je bilo, Suverensiti",
			"field_applicant":                  "Dosta je bilo, Suverensiti",
			"field_fullName":                   fullName,
			"field_personalNumber":             personalNumber.Text,
			"field_place":                      city.Text,
			"field_streetHouseNumber":          address.Text + " " + addressNo.Text,
			"field_firstLastName":              fullName,
			"field_dateOfBirth":                DateOfBirth.Text,
			"field_placeStreetWithHouseNumber": city.Text + " " + place.Text,
			"field_documentInfo":               issuingAuthority.Text + ", " + issuingDate.Text + ", " + documentNumber.Text,
			"field_authorizedCertifier":        authorizedCertifierName.Text,
			"field_workingPlace":               authorizedCertifierAddress.Text,
			"field_documentRegistryNo":         "",
			"field_location":                   place.Text,
			"field_date":                       "17.10",
		}

		execPath := filepath.Dir(executable) // Finds the directory of the executable.
		formPath := filepath.Join(execPath, "templates/form-01.pdf")

		err = helper.AppendCSV(form)
		if err != nil {
			fmt.Println("Error getting executable path:", err)
			SetStatus("Greska prilikom dodavanja", err)
			return
		}

		pdfInject := pdfinject.New()
		_, err = pdfInject.FillWithDestFile(form, formPath, "tmp.pdf")
		if err != nil {
			fmt.Println("Error getting executable path:", err)
		}
		helper.PrintPDF("tmp.pdf", "Parlament")
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Ime", givenName),
			widget.NewFormItem("Srednje ime", parentGivenName),
			widget.NewFormItem("Prezime", givenSurname),
			widget.NewFormItem("Datum rođenja", DateOfBirth),
			widget.NewFormItem("Grad", city),
			widget.NewFormItem("Ulica", address),
			widget.NewFormItem("Kućni broj", addressNo),
			widget.NewFormItem("Ime i prezime overitelja", authorizedCertifierName),
			widget.NewFormItem("Ulica i broj overitelja", authorizedCertifierAddress),
			widget.NewFormItem("JMBG", personalNumber),
			widget.NewFormItem("Broj dokumenta", documentNumber),
			widget.NewFormItem("Dokument izdao", issuingAuthority),
			widget.NewFormItem("Datum izdavanja", issuingDate),
		),
		submitButton,
		readCardButton,
	)

	(*window).SetContent(form)

	startPageOn = false
}

func setUI(doc document.Document) {
	ui := doc.BuildUI(statusBar, enableManualUI)
	columns := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), ui, layout.NewSpacer())
	container := container.New(layout.NewPaddedLayout(), columns)
	(*window).SetContent(container)

	(*window).Resize(container.MinSize())
	startPageOn = false
}

func setStartPage(status, explanation string, err error) {
	isError := false
	if err != nil {
		isError = true
	}

	if verbose && isError {
		fmt.Println(err)
	}

	startPage.SetStatus(status, explanation, isError)
	startPage.Refresh()

	if !startPageOn {
		(*window).SetContent(container.New(layout.NewPaddedLayout(), startPage))
		startPageOn = true
	}

}

func SetStatus(status string, err error) {
	isError := false
	if err != nil {
		isError = true
	}

	if verbose && isError {
		fmt.Println(err)
	}

	statusBar.SetStatus(status, isError)
	statusBar.Refresh()
}
