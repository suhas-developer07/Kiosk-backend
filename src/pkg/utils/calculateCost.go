package utils

import (
	"math"
	"strconv"
	"strings"
)

func CountPages(rangeStr string) int {
	parts := strings.Split(rangeStr, ",")
	total := 0

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.Contains(p, "-") {
			r := strings.Split(p, "-")
			start, _ := strconv.Atoi(strings.TrimSpace(r[0]))
			end, _ := strconv.Atoi(strings.TrimSpace(r[1]))
			total += (end - start + 1)
		} else {
			total++
		}
	}
	return total
}

func ParsePageLayout(up string) int {
	up = strings.TrimSpace(strings.ToLower(up))
	up = strings.TrimSuffix(up, "-up")
	n, err := strconv.Atoi(up)
	if err != nil {
		return 1
	}
	return n
}

func CalculatePrintJob(
	PageRange string,
	PageLayout string,
	PrintingSide string,
	PrintingMode string,
	Copies int,
) (int, int) {

	pages := CountPages(PageRange)
	paperSize := ParsePageLayout(PageLayout)

	pagesPerSheet := paperSize
	if PrintingSide == "double" {
		pagesPerSheet = paperSize * 2
	}

	sheets := int(math.Ceil(float64(pages) / float64(pagesPerSheet)))

	costPerPrintingSide := 1
	if PrintingMode == "PrintingMode" {
		costPerPrintingSide = 5
	}

	var totalPrintingSides int

	if PrintingSide == "single" {
		totalPrintingSides = sheets
	} else {
		fullSheets := pages / (paperSize * 2)
		remainingPages := pages % (paperSize * 2)

		if remainingPages == 0 {
			totalPrintingSides = fullSheets * 2
		} else if remainingPages <= paperSize {
			totalPrintingSides = fullSheets*2 + 1
		} else {
			totalPrintingSides = fullSheets*2 + 2
		}
	}

	totalPrintingSides = totalPrintingSides * Copies
	price := totalPrintingSides * costPerPrintingSide

	return sheets * Copies, price
}
