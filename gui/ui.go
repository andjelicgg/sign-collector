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
	"time"
)

var statusBar *widgets.StatusBar
var window *fyne.Window
var verbose bool
var startPageOn bool // possible data races
var startPage *widgets.StartPage

func StartGui() {
	startPageOn = true

	appW := app.New()
	win := appW.NewWindow("DJB Potpisi")
	window = &win

	appW.Settings().SetTheme(MyTheme{})

	statusBar = widgets.NewStatusBar()

	startPage = widgets.NewStartPage()
	test := widget.NewButton("Unesi ručno", func() {
		enableManualUI()
	})

	var loadCardButton *widget.Button
	loadCardButton = widget.NewButton("Učitaj karticu", func() {
		loadCardButton.Disable()
		test.Disable()
		loadCardButton.SetText("Učitavanje..")
		LoadCard()
	})

	startPage.SetStatus("", "", false)

	valueText := widget.NewLabelWithStyle("Odaberite opciju", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	valueText.Wrapping = fyne.TextWrapWord
	valueText.Resize(fyne.NewSize(300, valueText.MinSize().Height))

	win.SetContent(
		container.New(
			layout.NewVBoxLayout(),
			valueText,
			test,
			loadCardButton,
		),
	)
	win.Resize(fyne.NewSize(740, 400))

	win.ShowAndRun()
}

func formatName(firstName string, lastName string, middleName string) string {
	var st strings.Builder

	st.WriteString(firstName)
	if len(middleName) > 0 {
		st.WriteString(", " + middleName + ", ")
	} else {
		st.WriteString(" ")
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

func LoadCard() {
	ctx, err := scard.EstablishContext()
	if err != nil {
		fmt.Printf("Error establishing context: %s", err)
		return
	}

	defer ctx.Release()
	Pooler(ctx)
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

	readCardButton := widget.NewButton("Učitaj karticu", func() {
		LoadCard()
	})

	// Manual print Parlament
	submitButton := widget.NewButton("Štampaj", func() {
		executable, err := os.Executable() // Gets the path of the current executable.
		if err != nil {
			fmt.Println("Error getting executable path:", err)
		}

		fullName := formatName(givenName.Text, givenSurname.Text, parentGivenName.Text)
		currentTime := time.Now()

		form := map[string]interface{}{
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
			"field_date":                       currentTime.Format("02.01"),
		}

		execPath := filepath.Dir(executable) // Finds the directory of the executable.
		formPath := filepath.Join(execPath, "templates/form-01.pdf")

		err = helper.AppendCSV(form, "parlament")
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
		helper.PrintPDF("tmp.pdf", "Parlament X2")
	})

	// Manual print Local
	submitLocal := widget.NewButton("Štampaj BG", func() {
		executable, err := os.Executable() // Gets the path of the current executable.
		if err != nil {
			fmt.Println("Error getting executable path:", err)
		}

		fullName := formatName(givenName.Text, givenSurname.Text, parentGivenName.Text)
		currentTime := time.Now()

		form := map[string]interface{}{
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
			"field_date":                       currentTime.Format("02.01"),
		}

		execPath := filepath.Dir(executable) // Finds the directory of the executable.
		formPath := filepath.Join(execPath, "templates/form-02.pdf")

		err = helper.AppendCSV(form, "beograd")
		if err != nil {
			fmt.Println("Error getting executable path:", err)
			SetStatus("Greska prilikom dodavanja", err)
			return
		}

		pdfInject := pdfinject.New()
		_, err = pdfInject.FillWithDestFile(form, formPath, "tmp-02.pdf")
		if err != nil {
			fmt.Println("Error getting executable path:", err)
		}
		helper.PrintPDF("tmp-02.pdf", "Beograd")
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
	)

	buttons := container.New(layout.NewHBoxLayout(), submitButton, submitLocal, readCardButton)
	flexSpacer := layout.NewSpacer()
	formWrap := container.New(layout.NewVBoxLayout(), form, buttons, flexSpacer)

	fixedSpacer := NewCustomSpacer(20, 20)
	paddedContainer := container.New(
		layout.NewBorderLayout(fixedSpacer, fixedSpacer, fixedSpacer, fixedSpacer),
		fixedSpacer,
		fixedSpacer,
		fixedSpacer,
		fixedSpacer,
		formWrap,
	)

	(*window).Resize(fyne.NewSize(740, 400))
	(*window).SetContent(paddedContainer)

	startPageOn = false
}

func setUI(doc document.Document) {
	reloadCard := widget.NewButton("Učitaj sledeću karticu", func() {
		LoadCard()
	})
	ui := doc.BuildUI(statusBar, enableManualUI, reloadCard)
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

	if len(explanation) > 5 {
		loadCardButton := widget.NewButton("Učitaj sledeću karticu", func() {
			LoadCard()
		})

		explainLabel := widget.NewLabel("Da li je kartica prisutna?")

		container := container.New(layout.NewVBoxLayout(), explainLabel, loadCardButton)
		(*window).SetContent(container)

		(*window).Resize(fyne.NewSize(740, 400))
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
