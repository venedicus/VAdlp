package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadFileForce(url, dest string, progress func(pct int), force bool) error {
	if !force {
		if _, err := os.Stat(dest); err == nil {
			if progress != nil {
				progress(100)
			}
			return nil
		}
	}

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmp := dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	total := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			wn, wErr := f.Write(buf[:n])
			written += int64(wn)
			if wErr != nil {
				f.Close()
				os.Remove(tmp)
				return wErr
			}
			if progress != nil && total > 0 {
				progress(int(written * 100 / total))
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			f.Close()
			os.Remove(tmp)
			return readErr
		}
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
