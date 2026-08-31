package forge

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func Inventory(root string) (InventoryReport, error) {
	report := InventoryReport{Schema: "gooo/language-delta-forge/inventory/v1", RootReadmeExcluded: true}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && (entry.Name() == ".git" || entry.Name() == "_artifact") {
				return filepath.SkipDir
			}
			if path != root {
				report.Directories++
			}
			return nil
		}
		if path == filepath.Join(root, "README.md") {
			return nil
		}
		if strings.Contains(filepath.Clean(path), string(filepath.Separator)+".git"+string(filepath.Separator)) {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		lines := physicalLines(raw)
		report.Files++
		report.PhysicalLines += lines
		switch filepath.Ext(path) {
		case ".go":
			report.GoFiles++
			report.GoLines += lines
		case ".gooo":
			report.GoooFiles++
			report.GoooLines += lines
		}
		return nil
	})
	return report, err
}

func physicalLines(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	lines := bytes.Count(raw, []byte{'\n'})
	if raw[len(raw)-1] != '\n' {
		lines++
	}
	return lines
}

func CountFiles(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && (entry.Name() == ".git" || entry.Name() == "_artifact") {
				return filepath.SkipDir
			}
			return nil
		}
		if path != filepath.Join(root, "README.md") {
			count++
		}
		return nil
	})
	return count, err
}
