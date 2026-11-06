package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	dropPrompt     = "Drop your THP8 folder here or click to choose it!"
	dropZonePad    = 32
	dropZoneHeight = 260
	dropZoneWidth  = 320
	boxStroke      = 2
	boxCorner      = 8
	dropIconSize   = 64
)

type dropZone struct {
	widget.BaseWidget
	onTap      func()
	hovered    bool
	icon       fyne.CanvasObject
	prompt     *widget.Label
	statusIcon *widget.Icon
	status     *widget.Label
	detail     *widget.Label
	statusRow  *fyne.Container
	content    *fyne.Container
}

func newDropZone(onTap func()) *dropZone {
	z := &dropZone{onTap: onTap}
	icon := widget.NewIcon(theme.UploadIcon())
	z.icon = container.NewCenter(container.NewGridWrap(fyne.NewSize(dropIconSize, dropIconSize), icon))
	z.prompt = wrappedLabel(dropPrompt)
	z.prompt.Alignment = fyne.TextAlignCenter
	z.statusIcon = widget.NewIcon(nil)
	z.statusIcon.Hide()
	z.status = wrappedLabel("")
	z.status.Alignment = fyne.TextAlignCenter
	z.status.TextStyle = fyne.TextStyle{Bold: true}
	z.statusRow = container.NewBorder(nil, nil, z.statusIcon, nil, z.status)
	z.statusRow.Hide()
	z.detail = wrappedLabel("")
	z.detail.Alignment = fyne.TextAlignCenter
	z.detail.Hide()
	z.content = container.NewVBox(z.icon, z.prompt, z.statusRow, z.detail)
	z.ExtendBaseWidget(z)
	return z
}

func (z *dropZone) setMessage(icon fyne.Resource, status, detail string, busy bool) {
	if icon != nil {
		z.statusIcon.SetResource(icon)
	}
	showIf(z.statusIcon, icon != nil)
	z.status.SetText(status)
	showIf(z.statusRow, status != "")
	z.detail.SetText(detail)
	showIf(z.detail, detail != "")
	showIf(z.prompt, !busy)
	showIf(z.icon, !busy)
	z.Refresh()
}

func (z *dropZone) Tapped(_ *fyne.PointEvent) {
	z.onTap()
}

func (z *dropZone) MouseIn(_ *desktop.MouseEvent) {
	z.hovered = true
	z.Refresh()
}

func (z *dropZone) MouseMoved(_ *desktop.MouseEvent) {}

func (z *dropZone) MouseOut() {
	z.hovered = false
	z.Refresh()
}

func (z *dropZone) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (z *dropZone) CreateRenderer() fyne.WidgetRenderer {
	return newDropZoneRenderer(z)
}

type dropZoneRenderer struct {
	zone *dropZone
	box  *canvas.Rectangle
}

func newDropZoneRenderer(zone *dropZone) *dropZoneRenderer {
	box := canvas.NewRectangle(color.Transparent)
	box.StrokeWidth = boxStroke
	box.CornerRadius = boxCorner
	r := &dropZoneRenderer{zone: zone, box: box}
	r.paintBox()
	return r
}

func (r *dropZoneRenderer) Layout(size fyne.Size) {
	content := r.zone.content
	width := max(size.Width-2*dropZonePad, 0)
	for range 2 {
		content.Resize(fyne.NewSize(width, content.MinSize().Height))
	}
	content.Move(fyne.NewPos(dropZonePad, (size.Height-content.Size().Height)/2))
	r.box.Resize(size)
	r.paintBox()
}

func (r *dropZoneRenderer) paintBox() {
	name := theme.ColorNamePlaceHolder
	if r.zone.hovered {
		name = theme.ColorNamePrimary
	}
	r.box.StrokeColor = theme.Color(name)
	r.box.Refresh()
}

func (r *dropZoneRenderer) MinSize() fyne.Size {
	content := r.zone.content.MinSize()
	return fyne.NewSize(max(dropZoneWidth, content.Width+2*dropZonePad),
		max(dropZoneHeight, content.Height+2*dropZonePad))
}

func (r *dropZoneRenderer) Refresh() {
	r.paintBox()
	r.zone.content.Refresh()
	r.Layout(r.zone.Size())
}

func (r *dropZoneRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.box, r.zone.content}
}

func (r *dropZoneRenderer) Destroy() {}
