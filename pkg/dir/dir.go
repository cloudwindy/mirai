package dir

import (
	"os"
	"path"
)

func Is(p string) (bool, error) {
	info, err := os.Stat(p)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func Index(p string, index string) (string, bool, error) {
	ok, err := Is(p)
	if err != nil {
		return "", false, err
	}
	if ok {
		return path.Join(p, index), true, nil
	}
	return p, false, nil
}
