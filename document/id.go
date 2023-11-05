package document

import (
	"bitbucket.org/inceptionlib/pdfinject-go"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/signintech/gopdf"
	"github.com/ubavic/bas-celik/widgets"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type IdDocument struct {
	Loaded                 bool
	Photo                  image.Image
	DocumentNumber         string
	IssuingDate            string
	ExpiryDate             string
	IssuingAuthority       string
	PersonalNumber         string
	Surname                string
	GivenName              string
	ParentName             string
	Sex                    string
	PlaceOfBirth           string
	CommunityOfBirth       string
	StateOfBirth           string
	StateOfBirthCode       string
	DateOfBirth            string
	State                  string
	Community              string
	Place                  string
	Street                 string
	AddressNumber          string
	AddressLetter          string
	AddressEntrance        string
	AddressFloor           string
	AddressApartmentNumber string
	AddressDate            string
}

func (doc *IdDocument) formatName() string {
	return doc.GivenName + ", " + doc.ParentName + ", " + doc.Surname
}

func (doc *IdDocument) formatAddress() string {
	var address strings.Builder

	address.WriteString(doc.Street)
	address.WriteString(" ")
	address.WriteString(doc.AddressNumber)
	address.WriteString(doc.AddressLetter)

	if len(doc.AddressApartmentNumber) != 0 {
		address.WriteString("/")
		address.WriteString(doc.AddressApartmentNumber)
	}

	if len(doc.Community) > 0 {
		address.WriteString(", ")
		address.WriteString(doc.Community)
	}

	address.WriteString(", ")
	address.WriteString(doc.Place)

	return address.String()
}

func (doc *IdDocument) formatPlaceOfBirth() string {
	var placeOfBirth strings.Builder

	placeOfBirth.WriteString(doc.PlaceOfBirth)
	placeOfBirth.WriteString(", ")
	placeOfBirth.WriteString(doc.CommunityOfBirth)
	placeOfBirth.WriteString(", ")
	placeOfBirth.WriteString(doc.StateOfBirth)

	return placeOfBirth.String()
}

func (doc IdDocument) BuildUI(pdfHandler func(), statusBar *widgets.StatusBar) *fyne.Container {
	nameF := widgets.NewField("Ime, ime roditelja, prezime", doc.formatName(), 350)
	birthDateF := widgets.NewField("Datum rođenja", doc.DateOfBirth, 100)
	sexF := widgets.NewField("Pol", doc.Sex, 50)
	personalNumberF := widgets.NewField("JMBG", doc.PersonalNumber, 200)
	birthRow := container.New(layout.NewHBoxLayout(), sexF, birthDateF, personalNumberF)
	birthPlaceF := widgets.NewField("Mesto rođenja, opština i država", doc.formatPlaceOfBirth(), 350)
	addressF := widgets.NewField("Prebivalište i adresa stana", doc.formatAddress(), 350)
	addressDateF := widgets.NewField("Datum promene adrese", doc.AddressDate, 10)
	personInformationGroup := widgets.NewGroup("Podaci o građaninu", nameF, birthRow, birthPlaceF, addressF, addressDateF)

	issuedByF := widgets.NewField("Dokument izdaje", doc.IssuingAuthority, 10)
	documentNumberF := widgets.NewField("Broj dokumenta", doc.DocumentNumber, 100)
	issueDateF := widgets.NewField("Datum izdavanja", doc.IssuingDate, 100)
	expiryDateF := widgets.NewField("Važi do", doc.ExpiryDate, 100)
	docRow := container.New(layout.NewHBoxLayout(), documentNumberF, issueDateF, expiryDateF)
	docGroup := widgets.NewGroup("Podaci o dokumentu", issuedByF, docRow)
	colRight := container.New(layout.NewVBoxLayout(), personInformationGroup, docGroup)

	imgWidget := canvas.NewImageFromImage(doc.Photo)
	imgWidget.SetMinSize(fyne.Size{Width: 200, Height: 250})
	imgWidget.FillMode = canvas.ImageFillContain
	colLeft := container.New(layout.NewVBoxLayout(), imgWidget)
	cols := container.New(layout.NewHBoxLayout(), colLeft, colRight)

	saveButton := widget.NewButton("Štampaj", pdfHandler)
	buttonBar := container.New(layout.NewHBoxLayout(), statusBar, layout.NewSpacer(), saveButton)

	return container.New(layout.NewVBoxLayout(), cols, buttonBar)
}

func (doc *IdDocument) BuildSign() (*string, error) {
	executable, err := os.Executable() // Gets the path of the current executable.
	if err != nil {
		fmt.Println("Error getting executable path:", err)
	}

	form := map[string]interface{}{
		"field_politicalPartyName":         "Dosta je bilo, Suverensiti",
		"field_applicant":                  "Dosta je bilo, Suverensiti",
		"field_fullName":                   doc.formatName(),
		"field_personalNumber":             doc.PersonalNumber,
		"field_place":                      doc.Place,
		"field_streetHouseNumber":          doc.formatAddress(),
		"field_firstLastName":              doc.formatName(),
		"field_dateOfBirth":                doc.DateOfBirth,
		"field_placeStreetWithHouseNumber": doc.Place,
		"field_documentInfo":               doc.IssuingAuthority + ", " + doc.IssuingDate + ", " + doc.DocumentNumber,
		"field_authorizedCertifier":        doc.IssuingAuthority,
		"field_workingPlace":               doc.Place,
		"field_documentRegistryNo":         "",
		"field_location":                   "Beograd",
		"field_date":                       "17.10",
	}

	execPath := filepath.Dir(executable) // Finds the directory of the executable.
	formPath := filepath.Join(execPath, "templates/form-01.pdf")

	pdfInject := pdfinject.New()
	dest, err := pdfInject.FillWithDestFile(form, formPath, "tmp.pdf")
	if err != nil {
		fmt.Println("Error getting executable path:", err)
	}

	//err = os.RemoveAll("tmp.pdf")
	//if err != nil {
	//	return nil, err
	//}

	return dest, nil
}

func (doc *IdDocument) BuildPdf() ([]byte, string, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	pdf.AddPage()

	err := pdf.AddTTFFontData("liberationsans", font)
	if err != nil {
		return nil, "", fmt.Errorf("loading font: %w", err)
	}

	err = pdf.SetFont("liberationsans", "", 13.5)
	if err != nil {
		return nil, "", fmt.Errorf("setting font: %w", err)
	}

	// Color the page
	pdf.SetLineWidth(0.1)
	pdf.SetFillColor(124, 252, 0) //setup fill color
	pdf.RectFromUpperLeftWithStyle(50, 100, 400, 600, "FD")
	pdf.SetFillColor(0, 0, 0)

	pdf.SetXY(50, 50)

	// Import page 1
	tpl1 := pdf.ImportPage("templates/form-01.pdf", 1, "/MediaBox")
	pdf.SetLineType("solid")
	// Draw pdf onto page
	pdf.UseImportedTemplate(tpl1, 50, 100, 400, 0)

	fileName := strings.ToLower(doc.GivenName + "_" + doc.Surname + ".pdf")

	pdf.SetInfo(gopdf.PdfInfo{
		Title:        doc.GivenName + " " + doc.Surname,
		Author:       "Čitač",
		Subject:      "Lična karta",
		CreationDate: time.Now(),
	})

	return pdf.GetBytesPdf(), fileName, nil
}
