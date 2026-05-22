package fyneui

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/core"
	"vadlp/internal/downloader"
	"vadlp/internal/i18n"
)

func showFormatPicker(w fyne.Window, cfg core.Config, onPick func(formatID string)) {
	if strings.TrimSpace(cfg.URL) == "" {
		dialog.ShowError(fmt.Errorf("%s", i18n.T("format.need_url", nil)), w)
		return
	}

	progBody := container.NewVBox(
		widget.NewLabel(i18n.T("format.fetching", nil)),
		widget.NewProgressBarInfinite(),
	)
	prog := dialog.NewCustomWithoutButtons(i18n.T("format.title", nil), progBody, w)
	prog.Show()

	go func() {
		result, err := downloader.Probe(cfg)
		uiExec(func() {
			prog.Hide()
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			openFormatDialog(w, result, onPick)
		})
	}()
}

func openFormatDialog(w fyne.Window, result downloader.ProbeResult, onPick func(formatID string)) {
	if len(result.Entries) == 0 {
		dialog.ShowInformation(i18n.T("format.title", nil), i18n.T("format.none", nil), w)
		return
	}

	var formatIDs []string

	infoLabel := widget.NewLabel("")
	infoLabel.Wrapping = fyne.TextWrapWord

	thumb := canvas.NewImageFromResource(nil)
	thumb.FillMode = canvas.ImageFillContain
	thumbMin := DialogSize(w, 320, 180)
	thumb.SetMinSize(thumbMin)

	formatSelect := widget.NewSelect([]string{}, nil)

	applyEntry := func(idx int) {
		entry := result.Entries[idx]
		infoLabel.SetText(fmt.Sprintf("%s\n%s · %s", entry.Title, entry.Uploader, entry.Duration))
		loadThumbnail(entry.Thumbnail, thumb)

		labels := make([]string, 0, len(entry.Formats))
		formatIDs = formatIDs[:0]
		for _, f := range entry.Formats {
			if f.ID == "" {
				continue
			}
			labels = append(labels, downloader.FormatLabel(f))
			formatIDs = append(formatIDs, f.ID)
		}
		if len(labels) == 0 {
			formatSelect.Options = []string{i18n.T("format.no_formats", nil)}
			formatSelect.ClearSelected()
			formatSelect.Disable()
			return
		}
		formatSelect.Enable()
		formatSelect.Options = labels
		formatSelect.SetSelected(labels[0])
	}

	applyEntry(0)

	top := container.NewHBox(thumb, infoLabel)
	body := container.NewVBox(top)
	if len(result.Entries) > 1 {
		names := make([]string, len(result.Entries))
		for i, e := range result.Entries {
			names[i] = fmt.Sprintf("%d. %s", i+1, e.Title)
		}
		entrySelect := widget.NewSelect(names, func(s string) {
			for i, n := range names {
				if n == s {
					applyEntry(i)
					break
				}
			}
		})
		entrySelect.SetSelected(names[0])
		body.Add(widget.NewForm(widget.NewFormItem(i18n.T("form.playlist_item", nil), entrySelect)))
	}
	body.Add(widget.NewForm(widget.NewFormItem(i18n.T("form.format", nil), formatSelect)))

	pick := func() string {
		if formatSelect.Disabled() {
			return ""
		}
		idx := formatSelect.SelectedIndex()
		if idx < 0 || idx >= len(formatIDs) {
			return ""
		}
		return formatIDs[idx]
	}

	d := dialog.NewCustomConfirm(
		i18n.T("format.title", nil),
		i18n.T("format.use", nil),
		i18n.T("btn.close", nil),
		body,
		func(ok bool) {
			if ok {
				if id := pick(); id != "" {
					onPick(id)
				}
			}
		},
		w,
	)
	d.Resize(DialogSize(w, 720, 480))
	d.Show()
}

func loadThumbnail(url string, img *canvas.Image) {
	url = strings.TrimSpace(url)
	if url == "" {
		uiExec(func() {
			img.Resource = nil
			img.Refresh()
		})
		return
	}
	go func() {
		resp, err := http.Get(url) //nolint:noctx
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return
		}
		b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		if err != nil {
			return
		}
		decoded, _, err := image.Decode(bytes.NewReader(b))
		if err != nil {
			return
		}
		res := fyne.NewStaticResource("thumb", b)
		uiExec(func() {
			img.Image = decoded
			img.Resource = res
			img.Refresh()
		})
	}()
}
