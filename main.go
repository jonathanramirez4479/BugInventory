package main

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
	"os"
	"sort"
)

type Record struct {
	Bug      string `json:"bug"`
	Solution string `json:"solution"`
}

var records = []Record{}

var exampleList = tview.NewList()

func initData() {
	// Populare records with json data
	data, err := os.ReadFile("./data.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &records); err != nil {
		panic(err)
	}

	for _, rec := range records {
		exampleList.AddItem(rec.Bug, rec.Solution, 0, nil)
	}
}

func inputChange(text string) {
	exampleList.Clear()

	bugs := []string{}
	items := make(map[string]string)

	for _, rec := range records {
		bugs = append(bugs, rec.Bug)
		items[rec.Bug] = rec.Solution
	}

	rankings := fuzzy.RankFindFold(text, bugs)
	sort.Sort(rankings)

	for _, rank := range rankings {
		exampleList.AddItem(rank.Target, items[rank.Target], 0, nil)
	}
}

func main() {
	app := tview.NewApplication()
	flex := tview.NewFlex()
	form := tview.NewForm()
	inputField := tview.NewInputField()
	addRecordModal := tview.NewFlex().SetDirection(tview.FlexRow)
	selectedRecordFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	pages := tview.NewPages()
	inputHints := tview.NewTextView()
	selectedRecordModal := tview.NewModal().
		AddButtons([]string{"Close"}).
		SetTextColor(tcell.ColorYellow).
		SetBackgroundColor(tcell.ColorBlack).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Close" {
				pages.HidePage("selectedRecordFlex")
				app.SetFocus(exampleList)
			}
		})

	initData()

	inputHints.
		SetDynamicColors(true).
		SetText(`[yellow]<Ctrl+N> Add new record     <Enter>  Enter items list    <Ctrl+Q> Quit`).
		SetBackgroundColor(tcell.Color19)

	inputField.
		SetLabel("Enter an issue: ").
		SetLabelColor(tcell.ColorWhite).
		SetChangedFunc(inputChange).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				app.SetFocus(exampleList)
			}
		}).
		SetBorder(true).
		SetBackgroundColor(tcell.Color19)

	form.
		AddInputField("Title", "", 30, nil, nil).
		AddTextArea("Solution", "", 80, 0, 0, nil).
		AddTextView("", "<Enter> Submit form    <Esc> Exit prompt", 0, 0, false, true).
		SetBorder(true).
		SetTitle("Enter data").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				pages.HidePage("addRecordModal")
				app.SetFocus(inputField)

				titleItem := form.GetFormItemByLabel("Title")
				titleInput, ok := titleItem.(*tview.InputField)

				if !ok {
					panic("Title is not an input field")
				}

				title := titleInput.GetText()

				solutionItem := form.GetFormItemByLabel("Solution")
				solutionInput, ok := solutionItem.(*tview.TextArea)

				if !ok {
					panic("Solution is not a TextArea")
				}

				solution := solutionInput.GetText()

				item := Record{
					Bug:      title,
					Solution: solution,
				}

				records = append(records, item)
				exampleList.AddItem(item.Bug, item.Solution, 0, nil)

				return nil
			}

			return event
		})

	exampleList.
		SetMainTextColor(tcell.ColorWhite).
		SetSecondaryTextColor(tcell.ColorYellow).
		SetBorder(true).
		SetTitle("Results").
		SetTitleAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.Color17)

	exampleList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		// show new modal for showing bug info
		pages.ShowPage("selectedRecordFlex")
		selectedRecordModal.SetText(fmt.Sprintf(`
				Bug: %s,
				Solution: %s
			`, s1, s2))
		app.SetFocus(selectedRecordModal)
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlN:
			pages.ShowPage("addRecordModal")
			app.SetFocus(form)
			return nil
		case tcell.KeyEsc:
			pages.HidePage("addRecordModal")
			pages.HidePage("selectedRecordFlex")
			app.SetFocus(inputField)
			return nil
		case tcell.KeyCtrlQ:
			app.Stop()
			return nil
		}

		return event
	})

	flex.SetDirection(tview.FlexRow).
		AddItem(inputField, 3, 0, true).
		AddItem(inputHints, 6, 0, false).
		AddItem(exampleList, 0, 1, false).
		SetBorderColor(tcell.ColorWhite)

	addRecordModal.
		AddItem(nil, 0, 1, false).
		AddItem(form, 15, 1, true).
		AddItem(nil, 0, 1, false)

	selectedRecordFlex.
		AddItem(nil, 0, 1, false).
		AddItem(selectedRecordModal, 0, 5, true).
		AddItem(nil, 0, 1, false)

	pages.
		AddPage("main", flex, true, true).
		AddPage("addRecordModal", addRecordModal, true, false).
		AddPage("selectedRecordFlex", selectedRecordFlex, true, false)

	if err := app.SetRoot(pages, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
