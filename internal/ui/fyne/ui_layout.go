package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type layoutShell struct {
	holder            *fyne.Container
	splitH            *container.Split
	splitV            *container.Split
	activeSplit       *container.Split
	mode              string
	leftTabs          *container.AppTabs
	activityAccordion *widget.Accordion
	activitySplit     *container.Split
	topBar            *topBarHolder
	formatUI          *DownloadFormatUI
	profileActions    *fyne.Container
	commandPreview    *widget.Entry
	batchURLEntry     *widget.Entry
	extraArgsEntry    *widget.Entry
	queueScroll       *container.Scroll
	statusBadge       *StatusBadge
	phaseBadge        *PhaseBadge
	shortCollapsed    bool
	savedOffset       float64

	lastBucket        uint64
	lastSplitOffset   float64
	lastActivitySplit float64
	tabBottom         bool
	topCompact        bool
	profileCols       int
	queueMinH         float32
	cmdRows           int
	batchRows         int
	extraRows         int
	badgeCompact      bool
}

type topBarHolder struct {
	root    *fyne.Container
	wide    fyne.CanvasObject
	narrow  fyne.CanvasObject
	compact bool
}

func newLayoutShell(
	leftTabs *container.AppTabs,
	activityPanel fyne.CanvasObject,
	savedOffset float64,
) *layoutShell {
	splitH := container.NewHSplit(leftTabs, activityPanel)
	splitV := container.NewVSplit(leftTabs, activityPanel)
	holder := container.NewStack(splitH)
	return &layoutShell{
		holder:      holder,
		splitH:      splitH,
		splitV:      splitV,
		activeSplit: splitH,
		mode:        "h",
		leftTabs:    leftTabs,
		savedOffset: savedOffset,
	}
}

func (s *layoutShell) Root() fyne.CanvasObject {
	return s.holder
}

func newTopBar(
	journalBtn, openFolderBtn fyne.CanvasObject,
	statusBadge *StatusBadge,
	phaseBadge *PhaseBadge,
	stopBtn, runBtn fyne.CanvasObject,
) *topBarHolder {
	wide := container.NewHBox(
		journalBtn,
		openFolderBtn,
		layout.NewSpacer(),
		statusBadge.Root,
		phaseBadge.Root,
		stopBtn,
		runBtn,
	)
	narrow := container.NewVBox(
		container.NewHBox(journalBtn, openFolderBtn, layout.NewSpacer(), stopBtn, runBtn),
		container.NewHBox(statusBadge.Root, phaseBadge.Root),
	)
	root := container.NewStack(wide)
	return &topBarHolder{root: root, wide: wide, narrow: narrow}
}

func (t *topBarHolder) SetCompact(compact bool) {
	if t == nil || t.root == nil {
		return
	}
	if t.compact == compact {
		return
	}
	t.compact = compact
	if compact {
		t.root.Objects = []fyne.CanvasObject{t.narrow}
	} else {
		t.root.Objects = []fyne.CanvasObject{t.wide}
	}
	t.root.Refresh()
}

func (t *topBarHolder) Object() fyne.CanvasObject {
	if t == nil {
		return nil
	}
	return t.root
}

func newActivityLogPanel(logHeader fyne.CanvasObject, logs fyne.CanvasObject) fyne.CanvasObject {
	logScroll := container.NewVScroll(logs)
	logScroll.SetMinSize(fyne.NewSize(220, 200))
	return container.NewBorder(logHeader, nil, nil, nil, logScroll)
}

func newActivityBody(cmdAccordion, progressBlock, logPanel fyne.CanvasObject) (*container.Split, fyne.CanvasObject) {
	upper := container.NewVBox(cmdAccordion, progressBlock)
	split := container.NewVSplit(upper, logPanel)
	split.SetOffset(0.34)
	return split, split
}
