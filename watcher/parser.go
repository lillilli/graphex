package watcher

import (
	"strconv"
	"strings"
)

type FileData struct {
	Values [][2]float64 `json:"values"`
}

func parseFile(b []byte) *FileData {
	stringifiedFile := string(b)
	res := &FileData{Values: make([][2]float64, 0)}
	rows := strings.Split(stringifiedFile, "\r\n")

	for i, row := range rows {
		if i == 0 || strings.TrimSpace(row) == "" {
			continue
		}

		values := strings.Split(row, " ")
		if len(values) < 2 {
			continue
		}

		x, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			continue
		}

		y, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			continue
		}

		res.Values = append(res.Values, [2]float64{x, y})
	}

	return res
}
