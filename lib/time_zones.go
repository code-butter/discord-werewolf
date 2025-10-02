package lib

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

func SystemTimeZoneName() (string, error) {
	// For Linux
	resolvePath, err := filepath.EvalSymlinks("/etc/localtime")
	if err != nil {
		return "", err
	}
	parts := strings.Split(resolvePath, "/")
	// TODO: some are three long. We need a better method.
	return strings.Join(parts[len(parts)-2:], "/"), nil
}

func SystemTimeZone() (*time.Location, error) {
	var err error
	systemTzName, err := SystemTimeZoneName()
	if err != nil {
		return nil, err
	}
	systemTz, err := time.LoadLocation(systemTzName)
	if err != nil {
		return nil, err
	}
	return systemTz, nil
}

func AllTimeZoneNames() []string {
	// For Linux.
	base := "/usr/share/zoneinfo/"
	var timezones []string
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		subPath := strings.Replace(path, base, "", 1)
		if strings.Contains(subPath, "/") {
			timezones = append(timezones, subPath)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return timezones
}
