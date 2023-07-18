package admin

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Files(base string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := c.Params("*")
		if file == "" {
			names := make([]string, 0)
			err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					path = strings.TrimPrefix(path, base+"/")
					names = append(names, path)
				}
				return err
			})
			if err != nil {
				return err
			}
			return c.JSON(names)
		}
		file = path.Join(base, file)
		switch c.Method() {
		case "GET":
			return c.SendFile(file)
		case "PUT":
			ensureDir(file)
			if err := os.WriteFile(file, c.Body(), 0o644); err != nil {
				return err
			}
			return c.SendString("ok")
		case "DELETE":
			if err := os.Remove(file); err != nil {
				return err
			}
			return c.SendString("ok")
		}
		return nil
	}
}

// Create directories recursively
func ensureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
}
