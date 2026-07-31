package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

const userAgent = "VAdlp"

// httpGet fetches url with a VAdlp User-Agent and the shared timeout client.
func httpGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return httpClient.Do(req)
}

func downloadFileForce(ctx context.Context, url, dest string, progress func(pct int), force bool) error {
	if !force {
		if _, err := os.Stat(dest); err == nil {
			if progress != nil {
				progress(100)
			}
			return nil
		}
	}

	resp, err := httpGet(ctx, url)
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
