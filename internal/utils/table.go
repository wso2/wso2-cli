package utils

import (
	"strings"

	"github.com/olekukonko/tablewriter"
)

type TableConfig struct {
	ColSeparator    string
	AutoWrapText    bool
	HeaderAlignment int
	HeaderLine      bool
	Borders         tablewriter.Border
	CenterSeparator string
	RowSeparator    string
}

func CreateTable(data [][]string, headers []string, colSeparator string) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	if colSeparator == "" {
		colSeparator = "    "
	}
	table.SetColumnSeparator(colSeparator)
	table.SetRowSeparator("")
	table.SetAutoWrapText(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT) // Align header to left
	table.SetHeaderLine(false)                       // Remove line between header and first row
	if headers != nil {
		table.SetHeader(headers)
	}

	for _, row := range data {
		table.Append(row)
	}
	table.Render()
	return tableString.String()
}

func GenTable(data [][]string, headers []string, config *TableConfig) string {
	if config == nil {
		config = &TableConfig{
			ColSeparator:    "    ",
			AutoWrapText:    false,
			HeaderAlignment: tablewriter.ALIGN_LEFT,
			HeaderLine:      false,
			Borders:         tablewriter.Border{Left: true, Top: false, Right: false, Bottom: false},
			CenterSeparator: "",
			RowSeparator:    "",
		}
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetBorders(config.Borders)
	table.SetCenterSeparator(config.CenterSeparator)
	table.SetColumnSeparator(config.ColSeparator)
	table.SetRowSeparator(config.RowSeparator)
	table.SetAutoWrapText(config.AutoWrapText)
	table.SetHeaderAlignment(config.HeaderAlignment)
	table.SetHeaderLine(config.HeaderLine)
	if headers != nil {
		table.SetHeader(headers)
	}

	for _, row := range data {
		table.Append(row)
	}
	table.Render()
	return tableString.String()
}
