package fyneui

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

const layoutSizeQuantum float32 = 64

func layoutSizeBucket(w, h float32) uint64 {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	bw := uint64(int(w / layoutSizeQuantum))
	bh := uint64(int(h / layoutSizeQuantum))
	return bw<<32 | bh
}

func presetColumnsForWidth(w float32) int {
	switch {
	case w < 560:
		return 1
	case w < 920:
		return 2
	default:
		return 3
	}
}

func profileColumnsForWidth(w float32) int {
	if w < 720 {
		return 2
	}
	return 3
}

func commandPreviewRows(h float32) int {
	switch {
	case h < 620:
		return 3
	case h < 760:
		return 4
	default:
		return 6
	}
}

func batchURLRows(h float32) int {
	if h < 700 {
		return 2
	}
	return 3
}

func extraArgsRows(h float32) int {
	if h < 700 {
		return 3
	}
	return 4
}

func queueMinHeight(h float32) float32 {
	switch {
	case h < 520:
		return 120
	case h < 680:
		return 160
	case h < 840:
		return 200
	default:
		return 220
	}
}

func splitOffsetForWidth(mode string, w float32, saved float64) float64 {
	off := saved
	if off <= 0.05 || off >= 0.95 {
		off = 0.4
	}
	if mode == "v" {
		if off > 0.72 {
			off = 0.62
		}
		if off < 0.45 {
			off = 0.52
		}
		return off
	}
	switch {
	case w < 820:
		return 0.22
	case w < 1024:
		return 0.30
	case w < 1200:
		if off > 0.38 {
			return 0.38
		}
		return off
	default:
		return off
	}
}

func setSplitOffset(split *container.Split, off float64, last *float64) {
	if split == nil || last == nil {
		return
	}
	if math.Abs(split.Offset-off) < 0.005 && math.Abs(*last-off) < 0.005 {
		return
	}
	split.SetOffset(off)
	*last = off
}

// adaptLayout updates responsive chrome when the window size bucket changes.
// Call with stable canvas sizes (ignore values under 200px while the window is opening).
func adaptLayout(canvasSize fyne.Size, shell *layoutShell, savedOffset float64) {
	if shell == nil {
		return
	}
	w, h := canvasSize.Width, canvasSize.Height
	if w < 200 || h < 200 {
		return
	}

	bucket := layoutSizeBucket(w, h)
	if bucket == shell.lastBucket && shell.lastBucket != 0 {
		return
	}
	shell.lastBucket = bucket

	if w < 720 {
		if shell.mode != "v" {
			shell.mode = "v"
			shell.holder.Objects = []fyne.CanvasObject{shell.splitV}
			shell.activeSplit = shell.splitV
			shell.lastSplitOffset = -1
			shell.holder.Refresh()
		}
	} else if shell.mode != "h" {
		shell.mode = "h"
		shell.holder.Objects = []fyne.CanvasObject{shell.splitH}
		shell.activeSplit = shell.splitH
		shell.lastSplitOffset = -1
		shell.holder.Refresh()
	}

	off := splitOffsetForWidth(shell.mode, w, savedOffset)
	setSplitOffset(shell.activeSplit, off, &shell.lastSplitOffset)

	tabBottom := w < 700
	if shell.leftTabs != nil && shell.tabBottom != tabBottom {
		shell.tabBottom = tabBottom
		if tabBottom {
			shell.leftTabs.SetTabLocation(container.TabLocationBottom)
		} else {
			shell.leftTabs.SetTabLocation(container.TabLocationLeading)
		}
	}

	topCompact := w < 900
	if shell.topBar != nil && shell.topCompact != topCompact {
		shell.topCompact = topCompact
		shell.topBar.SetCompact(topCompact)
	}

	if shell.profileActions != nil {
		cols := profileColumnsForWidth(w)
		if shell.profileCols != cols {
			shell.profileCols = cols
			shell.profileActions.Layout = layout.NewGridLayoutWithColumns(cols)
			shell.profileActions.Refresh()
		}
	}

	if shell.formatUI != nil {
		shell.formatUI.SetPresetColumns(presetColumnsForWidth(w))
	}

	if shell.activityAccordion != nil && len(shell.activityAccordion.Items) > 0 {
		if h < 600 && !shell.shortCollapsed {
			shell.activityAccordion.Close(0)
			shell.shortCollapsed = true
		} else if h >= 680 {
			shell.shortCollapsed = false
		}
	}

	cmdRows := commandPreviewRows(h)
	if shell.commandPreview != nil && shell.cmdRows != cmdRows {
		shell.cmdRows = cmdRows
		shell.commandPreview.SetMinRowsVisible(cmdRows)
	}

	batchRows := batchURLRows(h)
	if shell.batchURLEntry != nil && shell.batchRows != batchRows {
		shell.batchRows = batchRows
		shell.batchURLEntry.SetMinRowsVisible(batchRows)
	}

	extraRows := extraArgsRows(h)
	if shell.extraArgsEntry != nil && shell.extraRows != extraRows {
		shell.extraRows = extraRows
		shell.extraArgsEntry.SetMinRowsVisible(extraRows)
	}

	qh := queueMinHeight(h)
	if shell.queueScroll != nil && shell.queueMinH != qh {
		shell.queueMinH = qh
		shell.queueScroll.SetMinSize(fyne.NewSize(0, qh))
	}

	compact := w < 820
	if shell.badgeCompact != compact {
		shell.badgeCompact = compact
		if shell.statusBadge != nil {
			shell.statusBadge.SetCompact(compact)
		}
		if shell.phaseBadge != nil {
			shell.phaseBadge.SetCompact(compact)
		}
	}

	if shell.activitySplit != nil {
		actOff := 0.34
		if h < 620 {
			actOff = 0.28
		} else if h < 760 {
			actOff = 0.31
		}
		setSplitOffset(shell.activitySplit, actOff, &shell.lastActivitySplit)
	}
}
