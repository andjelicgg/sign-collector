package document

import (
	"bitbucket.org/inceptionlib/pdfinject-go"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/signintech/gopdf"
	"github.com/ubavic/bas-celik/helper"
	"github.com/ubavic/bas-celik/widgets"
	"image"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

type IdDocument struct {
	Loaded                 bool
	Portrait               image.Image
	DocumentNumber         string
	DocumentType           string
	DocumentSerialNumber   string
	IssuingDate            string
	ExpiryDate             string
	IssuingAuthority       string
	PersonalNumber         string
	Surname                string
	GivenName              string
	ParentGivenName        string
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
	Location               string
}

func (doc *IdDocument) formatName() string {
	return doc.GivenName + " " + doc.Surname
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

func savePdf(doc *IdDocument) func() {
	return func() {
		err := doc.BuildPdf()
		if err != nil {
			fmt.Printf("generating PDF: %w", err)
			return
		}

		_, err = doc.BuildSign()
		if err != nil {
			return
		}
	}
}

func savePdfLocal(doc *IdDocument) func() {
	return func() {
		_, err := doc.BuildSignLocal()
		if err != nil {
			return
		}
	}
}

func (doc IdDocument) BuildUI(statusBar *widgets.StatusBar, enableManualUI func(), reloadCard *widget.Button) *fyne.Container {
	nameF := widgets.NewField("Ime", doc.GivenName, 100)
	parentF := widgets.NewField("Ime roditelja", doc.ParentGivenName, 100)
	surnameF := widgets.NewField("Prezime roditelja", doc.Surname, 100)
	fullNameRow := container.New(layout.NewHBoxLayout(), nameF, parentF, surnameF)
	birthDateF := widgets.NewField("Datum rođenja", doc.DateOfBirth, 100)
	sexF := widgets.NewField("Pol", doc.Sex, 50)
	personalNumberF := widgets.NewField("JMBG", doc.PersonalNumber, 200)
	birthRow := container.New(layout.NewHBoxLayout(), sexF, birthDateF, personalNumberF)
	birthPlaceF := widgets.NewField("Mesto rođenja, opština i država", doc.formatPlaceOfBirth(), 350)
	addressF := widgets.NewField("Prebivalište i adresa stana", doc.formatAddress(), 350)
	addressDateF := widgets.NewField("Datum promene adrese", doc.AddressDate, 10)
	personInformationGroup := widgets.NewGroup("Podaci o građaninu", fullNameRow, birthRow, birthPlaceF, addressF, addressDateF)

	issuedByF := widgets.NewField("Dokument izdaje", doc.IssuingAuthority, 100)
	documentNumberF := widgets.NewField("Broj dokumenta", doc.DocumentNumber, 100)
	issueDateF := widgets.NewField("Datum izdavanja", doc.IssuingDate, 100)
	expiryDateF := widgets.NewField("Važi do", doc.ExpiryDate, 100)
	docRow := container.New(layout.NewHBoxLayout(), documentNumberF, issueDateF, expiryDateF)
	docGroup := widgets.NewGroup("Podaci o dokumentu", issuedByF, docRow)
	colRight := container.New(layout.NewVBoxLayout(), personInformationGroup, docGroup)

	imgWidget := canvas.NewImageFromImage(doc.Portrait)
	imgWidget.SetMinSize(fyne.Size{Width: 200, Height: 250})
	imgWidget.FillMode = canvas.ImageFillContain
	colLeft := container.New(layout.NewVBoxLayout(), imgWidget)
	cols := container.New(layout.NewHBoxLayout(), colLeft, colRight)

	pdfHandler := savePdf(&doc)
	pdfHandlerLocal := savePdfLocal(&doc)

	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	var isLoading int32

	var saveButton *widget.Button
	var saveButtonLocal *widget.Button

	enableManualInput := widget.NewButton("Unesi ručno", func() {
		enableManualUI()
	})

	buttonAction := func() {
		if atomic.LoadInt32(&isLoading) == 1 {
			fmt.Println("Already loading")
			return
		}
		atomic.StoreInt32(&isLoading, 1)
		progress.Show()
		saveButton.SetText("Štampanje...")
		saveButton.Disable()
		reloadCard.Disable()
		enableManualInput.Disable()

		go func() {
			pdfHandler()
			time.Sleep(10 * time.Second) // Simulate a long-running task

			progress.Hide()
			saveButton.SetText("Štampaj")
			enableManualInput.Enable()
			saveButton.Enable()
			reloadCard.Enable()
			atomic.StoreInt32(&isLoading, 0)
		}()
	}

	buttonActionLocal := func() {
		if atomic.LoadInt32(&isLoading) == 1 {
			fmt.Println("Already loading")
			return
		}
		atomic.StoreInt32(&isLoading, 1)
		progress.Show()
		saveButtonLocal.SetText("Štampanje...")
		saveButtonLocal.Disable()
		saveButton.Disable()
		reloadCard.Disable()
		enableManualInput.Disable()

		go func() {
			pdfHandlerLocal()
			time.Sleep(10 * time.Second) // Simulate a long-running task

			progress.Hide()
			saveButton.SetText("Štampaj BG")
			enableManualInput.Enable()
			saveButtonLocal.Enable()
			saveButton.Enable()
			reloadCard.Enable()
			atomic.StoreInt32(&isLoading, 0)
		}()
	}

	saveButton = widget.NewButton("Štampaj", buttonAction)
	saveButtonLocal = widget.NewButton("Štampaj BG", buttonActionLocal)

	buttonBar := container.New(layout.NewHBoxLayout(), statusBar, layout.NewSpacer(), saveButton, saveButtonLocal, enableManualInput, reloadCard)

	return container.New(layout.NewVBoxLayout(), cols, buttonBar)
}

func (doc *IdDocument) BuildSign() (*string, error) {
	executable, err := os.Executable() // Gets the path of the current executable.
	if err != nil {
		fmt.Println("Error getting executable path:", err)
	}

	currentTime := time.Now()

	form := map[string]interface{}{
		"field_fullName":                   doc.formatName(),
		"field_personalNumber":             doc.PersonalNumber,
		"field_place":                      doc.Place,
		"field_streetHouseNumber":          doc.formatAddress(),
		"field_firstLastName":              doc.formatName(),
		"field_dateOfBirth":                doc.DateOfBirth,
		"field_placeStreetWithHouseNumber": doc.Place,
		"field_documentInfo":               doc.IssuingAuthority + ", " + doc.IssuingDate + ", " + doc.DocumentNumber,
		"field_authorizedCertifier":        "",
		"field_workingPlace":               "",
		"field_documentRegistryNo":         "",
		"field_location":                   "",
		"field_date":                       currentTime.Format("02.01"),
	}

	execPath := filepath.Dir(executable)
	formPath := filepath.Join(execPath, "templates/form-01.pdf")

	err = helper.AppendCSV(form, "parlament")
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return nil, err
	}

	pdfInject := pdfinject.New()
	dest, err := pdfInject.FillWithDestFile(form, formPath, "tmp.pdf")
	if err != nil {
		fmt.Println("Error getting executable path:", err)
	}

	helper.PrintPDF("tmp.pdf", "Parlament")
	helper.PrintPDF("tmp.pdf", "Parlament X2")

	return dest, nil
}

func (doc *IdDocument) BuildSignLocal() (*string, error) {
	executable, err := os.Executable() // Gets the path of the current executable.
	if err != nil {
		fmt.Println("Error getting executable path:", err)
	}

	currentTime := time.Now()

	form := map[string]interface{}{
		"field_fullName":                   doc.formatName(),
		"field_personalNumber":             doc.PersonalNumber,
		"field_place":                      doc.Place,
		"field_streetHouseNumber":          doc.formatAddress(),
		"field_firstLastName":              doc.formatName(),
		"field_dateOfBirth":                doc.DateOfBirth,
		"field_placeStreetWithHouseNumber": doc.Place,
		"field_documentInfo":               doc.IssuingAuthority + ", " + doc.IssuingDate + ", " + doc.DocumentNumber,
		"field_authorizedCertifier":        "",
		"field_workingPlace":               "",
		"field_documentRegistryNo":         "",
		"field_location":                   "",
		"field_date":                       currentTime.Format("02.01"),
	}

	execPath := filepath.Dir(executable)
	formPath := filepath.Join(execPath, "templates/form-02.pdf")

	err = helper.AppendCSV(form, "beograd")
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return nil, err
	}

	pdfInject := pdfinject.New()
	dest, err := pdfInject.FillWithDestFile(form, formPath, "tmp-02.pdf")
	if err != nil {
		fmt.Println("Error getting executable path:", err)
	}

	helper.PrintPDF("tmp-02.pdf", "BG")
	helper.PrintPDF("tmp-02.pdf", "BG X2")
	return dest, nil
}

func (doc *IdDocument) BuildPdf() (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case error:
				retErr = x
			default:
				retErr = errors.New("unknown panic")
			}
		}
	}()

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	pdf.AddPage()

	err := pdf.AddTTFFontData("liberationsans", font)
	if err != nil {
		fmt.Printf("loading font: %w", err)
		panic(fmt.Errorf("loading font: %w", err))
	}

	err = pdf.SetFont("liberationsans", "", 13.5)
	if err != nil {
		panic(fmt.Errorf("setting font: %w", err))
	}

	const leftMargin = 58.8
	const rightMargin = 535
	const textLeftMargin = 67.3

	line := func(width float64) {
		if width > 0 {
			pdf.SetLineWidth(width)
		}

		y := pdf.GetY()
		pdf.Line(leftMargin, y, rightMargin, y)
	}

	moveY := func(y float64) {
		pdf.SetXY(pdf.GetX(), pdf.GetY()+y)
	}

	cell := func(s string) {
		err := pdf.Cell(nil, s)
		if err != nil {
			panic(fmt.Errorf("putting text: %w", err))
		}
	}

	putData := func(label, data string) {
		y := pdf.GetY()

		pdf.SetX(textLeftMargin)
		texts, err := pdf.SplitTextWithWordWrap(label, 120)
		if err != nil {
			panic(err)
		}

		for i, text := range texts {
			cell(text)
			if i < len(texts)-1 {
				pdf.SetXY(textLeftMargin, pdf.GetY()+12)
			}
		}

		y1 := pdf.GetY()

		pdf.SetXY(textLeftMargin+128, y)
		texts, err = pdf.SplitTextWithWordWrap(data, 350)
		if err != nil {
			panic(err)
		}

		for i, text := range texts {
			cell(text)
			if i < len(texts)-1 {
				pdf.SetXY(textLeftMargin+128, pdf.GetY()+12)
			}
		}

		y2 := pdf.GetY()

		pdf.SetXY(textLeftMargin, math.Max(y1, y2)+24.67)
	}

	pdf.SetLineType("solid")
	pdf.SetY(59.041)
	line(0.83)

	pdf.SetXY(textLeftMargin+1.0, 68.5)
	err = pdf.SetCharSpacing(-0.2)
	if err != nil {
		panic(err)
	}
	cell("ČITAČ ELEKTRONSKE LIČNE KARTE: ŠTAMPA PODATAKA")

	err = pdf.SetCharSpacing(-0.1)
	if err != nil {
		panic(err)
	}

	pdf.SetY(88)
	line(0)

	err = pdf.ImageFrom(doc.Portrait, leftMargin, 102.8, &gopdf.Rect{W: 119.9, H: 159})
	if err != nil {
		panic(err)
	}
	pdf.SetLineWidth(0.48)
	pdf.SetFillColor(255, 255, 255)
	err = pdf.Rectangle(leftMargin, 102.8, 179, 262, "D", 0, 0)
	if err != nil {
		panic(err)
	}

	pdf.SetFillColor(0, 0, 0)

	pdf.SetY(276)
	line(1.08)
	moveY(8)
	pdf.SetXY(textLeftMargin, 284)
	err = pdf.SetFontSize(11.1)
	if err != nil {
		panic(err)
	}
	cell("Podaci o građaninu")
	moveY(16)
	line(0)
	moveY(9)

	putData("Prezime:", doc.Surname)
	putData("Ime:", doc.GivenName)
	putData("Ime jednog roditelja:", doc.ParentGivenName)
	putData("Datum rođenja:", doc.DateOfBirth)
	putData("Mesto rođenja, opština i država:", doc.formatPlaceOfBirth())
	putData("Prebivalište:", doc.formatAddress())
	putData("Datum promene adrese:", doc.AddressDate)
	putData("JMBG:", doc.PersonalNumber)
	putData("Pol:", doc.Sex)

	moveY(-8.67)
	line(0)
	moveY(9)
	cell("Podaci o dokumentu")
	moveY(16)

	line(0)
	moveY(9)
	putData("Dokument izdaje:", doc.IssuingAuthority)
	putData("Broj dokumenta:", doc.DocumentNumber)
	putData("Datum izdavanja:", doc.IssuingDate)
	putData("Važi do:", doc.ExpiryDate)

	moveY(-8.67)
	line(0)
	moveY(3)
	line(0)
	moveY(9)

	cell("Datum štampe: " + time.Now().Format("02.01.2006."))

	pdf.SetY(730.6)
	line(0.83)

	err = pdf.SetFontSize(9)
	if err != nil {
		panic(err)
	}

	pdf.SetXY(leftMargin, 739.7)
	cell("1. U čipu lične karte, podaci o imenu i prezimenu imaoca lične karte ispisani su na nacionalnom pismu onako kako su")
	pdf.SetXY(leftMargin, 749.9)
	cell("ispisani na samom obrascu lične karte, dok su ostali podaci ispisani latiničkim pismom.")
	pdf.SetXY(leftMargin, 759.7)
	cell("2. Ako se ime lica sastoji od dve reči čija je ukupna dužina između 20 i 30 karaktera ili prezimena od dve reči čija je")
	pdf.SetXY(leftMargin, 769.4)
	cell("ukupna dužina između 30 i 36 karaktera, u čipu lične karte izdate pre 18.08.2014. godine, druga reč u imenu ili prezimenu")
	pdf.SetXY(leftMargin, 779.1)
	cell("skraćuje se na prva dva karaktera")

	pdf.SetY(794.5)
	line(0)

	pdf.SetInfo(gopdf.PdfInfo{
		Title:        doc.GivenName + " " + doc.Surname,
		Author:       "Baš Čelik",
		Subject:      "Lična karta",
		CreationDate: time.Now(),
	})

	err = os.WriteFile("id.pdf", pdf.GetBytesPdf(), 0666)
	if err != nil {
		fmt.Printf("Failed to write PDF to file: %s\n", err)
		os.Exit(1)
	}

	helper.PrintPDF("id.pdf", "licna karata")

	return nil
}
