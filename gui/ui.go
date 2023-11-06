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

// CustomSpacer creates a spacer with a fixed size.
type CustomSpacer struct {
	widget.BaseWidget
	size fyne.Size
}

// NewCustomSpacer creates a new spacer with the given width and height.
func NewCustomSpacer(width, height float32) *CustomSpacer {
	s := &CustomSpacer{size: fyne.NewSize(width, height)}
	s.ExtendBaseWidget(s)
	return s
}

// MinSize returns the minimum size of the spacer, effectively making it a fixed size.
func (s *CustomSpacer) MinSize() fyne.Size {
	return s.size
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (s *CustomSpacer) CreateRenderer() fyne.WidgetRenderer {
	return &customSpacerRenderer{spacer: s}
}

type customSpacerRenderer struct {
	spacer *CustomSpacer
}

func (r *customSpacerRenderer) MinSize() fyne.Size {
	return r.spacer.size
}

func (r *customSpacerRenderer) Layout(_ fyne.Size) {
	// No layout needed for spacer
}

func (r *customSpacerRenderer) ApplyTheme() {
	// No theme needed for spacer
}

func (r *customSpacerRenderer) Refresh() {
	// No refresh needed for spacer
}

func (r *customSpacerRenderer) Objects() []fyne.CanvasObject {
	return nil
}

func (r *customSpacerRenderer) Destroy() {
	// No destroy actions needed for spacer
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

	// Use the standard spacer for flexible space
	flexSpacer := layout.NewSpacer()

	// Use the custom spacer for fixed space
	fixedSpacer := NewCustomSpacer(20, 20)

	// Create a container with padding around the content
	paddedContainer := container.New(
		layout.NewBorderLayout(fixedSpacer, fixedSpacer, fixedSpacer, fixedSpacer),
		fixedSpacer, // top
		fixedSpacer, // bottom
		fixedSpacer, // left
		fixedSpacer, // right
		form,
		flexSpacer, // Use the flexible spacer to push content to the center
	)

	minWidth := float32(200)  // Minimum width in pixels
	minHeight := float32(100) // Minimum height in pixels, adjust as needed
	(*window).Resize(fyne.NewSize(minWidth, minHeight))
	(*window).SetContent(paddedContainer)

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
