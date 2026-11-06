package ui

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	thpsRunPage = "https://thps.run"
	gitHubPage  = "https://github.com/TheAnastasis/AnaMod"
	pickPadding = 8
	linkGap     = 12
)

type edgeRow struct{}

func shownObjects(objects []fyne.CanvasObject) []fyne.CanvasObject {
	var shown []fyne.CanvasObject
	for _, object := range objects {
		if object.Visible() {
			shown = append(shown, object)
		}
	}
	return shown
}

func widestWidth(objects []fyne.CanvasObject) float32 {
	var widest float32
	for _, object := range objects {
		widest = max(widest, object.MinSize().Width)
	}
	return widest
}

func (edgeRow) MinSize(objects []fyne.CanvasObject) fyne.Size {
	shown := shownObjects(objects)
	if len(shown) == 0 {
		return fyne.NewSize(0, 0)
	}
	var tallest float32
	for _, object := range shown {
		tallest = max(tallest, object.MinSize().Height)
	}
	count := float32(len(shown))
	return fyne.NewSize(widestWidth(shown)*count+linkGap*(count-1), tallest)
}

func (edgeRow) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	shown := shownObjects(objects)
	widest := widestWidth(shown)
	for i, object := range shown {
		object.Resize(fyne.NewSize(widest, size.Height))
		switch {
		case i == 0:
			object.Move(fyne.NewPos(0, 0))
		case i == len(shown)-1:
			object.Move(fyne.NewPos(size.Width-widest, 0))
		default:
			object.Move(fyne.NewPos((size.Width-widest)/2, 0))
		}
	}
}

func (p *patcher) buildLinkRow() *fyne.Container {
	return container.New(edgeRow{},
		widget.NewButton("thps.run", func() { p.openLink(thpsRunPage) }),
		p.updateBtn,
		widget.NewButton("AnaMod GitHub", func() { p.openLink(gitHubPage) }),
	)
}

func (p *patcher) buildPickScreen() fyne.CanvasObject {
	box := container.New(layout.NewCustomPaddedLayout(0, pickPadding, 0, 0), p.zone)
	return container.New(
		layout.NewCustomPaddedLayout(pickPadding, pickPadding, pickPadding, pickPadding),
		container.NewBorder(nil, p.linkRow, nil, nil, box),
	)
}

func (p *patcher) openLink(link string) {
	page, err := url.Parse(link)
	if err == nil {
		err = p.openURL(page)
	}
	if err != nil {
		fmt.Fprintf(p.log, "Could not open %s: %v\n", link, err)
		p.showText("Link", "Couldn't open your browser. The page is at "+link)
	}
}
