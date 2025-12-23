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

func initData(bugList *tview.List, bugs *[]Bug) {
	/*
		Initialize records of bugs and bugsList for TUI
	*/
	data, err := os.ReadFile("data.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &bugs); err != nil {
		panic(err)
	}

	for _, bug := range *bugs {
		bugList.AddItem(bug.Label, bug.Solution, 0, nil)
	}
}

func inputChange(text string, bugList *tview.List, bugs *[]Bug) {
	bugList.Clear()

	bugLabels := []string{}
	items := make(map[string]string)

	for _, bug := range *bugs {
		bugLabels = append(bugLabels, bug.Label)
		items[bug.Label] = bug.Solution
	}

	rankings := fuzzy.RankFindFold(text, bugLabels)
	sort.Sort(rankings)

	for _, rank := range rankings {
		bugList.AddItem(rank.Target, items[rank.Target], 0, nil)
	}
}

func writeToDisk(bugs []Bug) error {
	data, err := json.MarshalIndent(&bugs, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile("data.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	bugs := []Bug{}
	app := tview.NewApplication()
	flex := tview.NewFlex()
	addBugForm := tview.NewForm()
	bugList := tview.NewList()
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
				app.SetFocus(bugList)
			}
		})

	initData(bugList, &bugs)

	inputHints.
		SetDynamicColors(true).
		SetText(`[yellow]<Ctrl+N> Add new record     <Enter>  Enter results list    <Ctrl+Q> Quit`).
		SetBackgroundColor(tcell.Color19)

	inputField.
		SetLabel("Enter an error message: ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.Color26).
		SetChangedFunc(func(text string) {
			inputChange(text, bugList, &bugs)
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				app.SetFocus(bugList)
				inputHints.SetText("")
			}
		}).
		SetBorder(true).
		SetBackgroundColor(tcell.Color19)

	addBugForm.
		AddInputField("Label", "", 30, nil, nil).
		AddTextArea("Solution", "", 80, 0, 0, nil).
		AddTextView("", "<Enter> Submit form    <Esc> Exit prompt", 0, 0, false, true).
		SetBorder(true).
		SetTitle("Enter data").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				pages.HidePage("addRecordModal")
				app.SetFocus(inputField)

				bugLabelItem := addBugForm.GetFormItemByLabel("Label")
				bugLabelInput, ok := bugLabelItem.(*tview.InputField)

				if !ok {
					panic("Title is not an input field")
				}

				bugLabel := bugLabelInput.GetText()

				bugSolutionItem := addBugForm.GetFormItemByLabel("Solution")
				bugSolutionInput, ok := bugSolutionItem.(*tview.TextArea)

				if !ok {
					panic("Solution is not a TextArea")
				}

				bugSolution := bugSolutionInput.GetText()

				bug := Bug{
					Label:    bugLabel,
					Solution: bugSolution,
				}

				bugs = append(bugs, bug)
				bugList.AddItem(bug.Label, bug.Solution, 0, nil)

				bugLabelInput.SetText("")
				bugSolutionInput.SetText("", false)

				return nil
			}

			return event
		})

	bugList.
		SetMainTextColor(tcell.ColorWhite).
		SetSecondaryTextColor(tcell.ColorYellow).
		SetBorder(true).
		SetTitle("Results").
		SetTitleAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.Color17)

	bugList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		// show new modal for showing bug info
		pages.ShowPage("selectedRecordFlex")
		selectedRecordModal.SetText(fmt.Sprintf(`
				Label: %s,
				Solution: %s
			`, s1, s2))
		app.SetFocus(selectedRecordModal)
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlN:
			pages.ShowPage("addRecordModal")
			app.SetFocus(addBugForm)
			return nil
		case tcell.KeyEsc:
			pages.HidePage("addRecordModal")
			pages.HidePage("selectedRecordFlex")
			app.SetFocus(inputField)
			inputHints.SetText(`[yellow]<Ctrl+N> Add new record     <Enter>  Enter results list    <Ctrl+Q> Quit`)
			return nil
		case tcell.KeyCtrlQ:
			writeToDisk(bugs)
			app.Stop()
			return nil
		}

		return event
	})

	flex.SetDirection(tview.FlexRow).
		AddItem(inputField, 3, 0, true).
		AddItem(inputHints, 6, 0, false).
		AddItem(bugList, 0, 1, false).
		SetBorderColor(tcell.ColorWhite)

	addRecordModal.
		AddItem(nil, 0, 1, false).
		AddItem(addBugForm, 15, 1, true).
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
