package main

import (
	// "github.com/gdamore/tcell/v2"
	// "fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
)

type Record struct {
	ErrorMessage string `json:"title"`
	Solution     string `json:"solution"`
}

var records = []Record{
	{ErrorMessage: "Null pointer dereference", Solution: "solution"},
	{ErrorMessage: "Race condition in goroutine", Solution: "solution"},
	{ErrorMessage: "Race condition for mutex", Solution: "solution"},
	{ErrorMessage: "Channel closed twice", Solution: "solution"},
	{ErrorMessage: "Deadlock in mutex", Solution: "solution"},
}

var exampleList = tview.NewList()

func initData() {
	for _, rec := range records {
		exampleList.AddItem(rec.ErrorMessage, rec.Solution, 0, nil)
	}
}

func inputChange(text string) {
	exampleList.Clear()

	errorMessages := []string{}

	for _, rec := range records {
		errorMessages = append(errorMessages, rec.ErrorMessage)
	}

	rankings := fuzzy.RankFindFold(text, errorMessages)
	sort.Sort(rankings)

	for _, rank := range rankings {
		exampleList.AddItem(rank.Target, "", 0, nil)
	}

}

func main() {
	app := tview.NewApplication()
	flex := tview.NewFlex()
	initData()

	inputField := tview.NewInputField()
	inputField.
		SetLabel("Enter an issue: ").
		SetChangedFunc(inputChange)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyDown, tcell.KeyTab:
			app.SetFocus(exampleList)
		case tcell.KeyEsc:
			app.Stop()
		default:
			app.SetFocus(inputField)
		}

		return event
	})

	flex.SetDirection(tview.FlexRow).
		AddItem(inputField, 0, 1, true).
		AddItem(exampleList, 0, 1, true)

	if err := app.SetRoot(flex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
