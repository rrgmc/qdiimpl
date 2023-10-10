package main

import (
	"errors"
	"fmt"
	"unicode"

	"golang.org/x/tools/go/packages"
)

func pkgInfoFromPath(srcDir string, packageName string, mode packages.LoadMode, tags []string) (*packages.Package, error) {
	var patterns []string
	if packageName != "" {
		patterns = append(patterns, packageName)
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode:       mode,
		Dir:        srcDir,
		BuildFlags: tags,
	}, patterns...)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, errors.New("package not found")
	}
	if len(pkgs) > 1 {
		return nil, errors.New("found more than one package")
	}
	if errs := pkgs[0].Errors; len(errs) != 0 {
		if len(errs) == 1 {
			return nil, errs[0]
		}
		return nil, fmt.Errorf("%s (and %d more errors)", errs[0], len(errs)-1)
	}
	return pkgs[0], nil
}

func getUniqueName(defaultName string, f func(nameExists string) bool) string {
	ct := 0
	currentName := defaultName
	for {
		if !f(currentName) {
			return currentName
		}
		if ct == 0 {
			currentName = fmt.Sprintf("%sQDII", defaultName)
		} else {
			currentName = fmt.Sprintf("%sQDII%d", defaultName, ct)
		}
		ct++
	}
}

// initialToLower converts initial to lower.
func initialToLower(s string) string {
	for _, r := range s {
		u := string(unicode.ToLower(r))
		return u + s[len(u):]
	}

	return s
}

// initialToUpper converts initial to upper.
func initialToUpper(s string) string {
	for _, r := range s {
		u := string(unicode.ToUpper(r))
		return u + s[len(u):]
	}

	return ""
}
