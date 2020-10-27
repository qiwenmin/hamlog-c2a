package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	version = "0.0.1"
)

var (
	csvFilename     = flag.String("i", "HamLog.csv", "input CSV filename")
	stationCallsign = flag.String("c", "BG1REN", "station callsign")
	startNo         = flag.Int("s", 0, "start number")

	latestNo = 0
)

func main() {
	flag.Parse()

	file, err := os.Open(*csvFilename)

	if err != nil {
		log.Fatalf("Fail to open %s!", *csvFilename)
	}

	defer file.Close()

	printHeader()

	scanner := bufio.NewScanner(file)

	count := 0
	lcount := 0
	ocount := 0

	for scanner.Scan() {
		lcount++

		line := scanner.Text()
		cols := strings.SplitN(line, ",", 2)
		//summary := cols[0]
		fields := strings.Split(cols[1], "|")

		// fmt.Fprintf(os.Stderr, "Converting %v...\n", summary)

		n, err := writeRecord(fields, *startNo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error [%v]. Line: %v\n", err, line)
			continue
		}

		ocount += n
		count++
	}

	fmt.Fprintf(os.Stderr,
		"Converted %v record(s) from %v lines with %v error(s). Exported %v record(s).\n",
		count, lcount, lcount-count, ocount)
	fmt.Fprintf(os.Stderr, "Latest No: %v\n", latestNo)
}

func writeRecord(fields []string, fromNo int) (int, error) {
	if len(fields) != 15 {
		return 0, errors.New("field error")
	}

	recNo, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, err
	}

	if recNo > latestNo {
		latestNo = recNo
	}

	if recNo < fromNo {
		return 0, nil
	}

	dateStr := fields[1]
	timeBegin := fields[2]
	timeEnd := fields[3]
	freq := fields[4]
	mode := fields[5]
	power := fields[6]
	stationQth := decodeStringField(fields[7])
	timezone := fields[8]
	rstSent := fields[9]
	rstRecv := fields[10]
	callsign := fields[11]
	opname := decodeStringField(fields[12])
	qth := decodeStringField(fields[13])
	comment := decodeStringField(fields[14])

	// fmt.Println(dateStr, timeBegin, timeEnd, freq, mode, power, stationQth, timezone, rstSent, rstRecv, callsign, opname, qth, comment)

	if isEmptyField(dateStr) || isEmptyField(timeBegin) {
		return 0, errors.New("QSO_DATE or TIME_ON is missing")
	}

	qsoDate, timeOn, err := toUTC(dateStr, timeBegin, timezone, 0)

	if err != nil {
		return 0, err
	}

	printField("QSO_DATE", qsoDate)
	printField("TIME_ON", timeOn)

	if (!isEmptyField(timeEnd)) && (timeEnd != timeBegin) {
		offset := time.Duration(0)
		if timeEnd < timeBegin {
			offset = time.Hour * 24
		}

		qsoDateOff, timeOff, err := toUTC(dateStr, timeEnd, timezone, offset)
		if err != nil {
			return 0, err
		}

		if qsoDateOff != qsoDate {
			printField("QSO_DATE_OFF", qsoDateOff)
		}

		printField("TIME_OFF", timeOff)
	}

	printField("CALL", callsign)

	printField("FREQ", freq)

	printField("BAND", freqToBand(freq))

	printField("MODE", mode)

	printField("RST_RCVD", rstRecv)
	printField("RST_SENT", rstSent)

	printField("TX_PWR", power)

	printField("NAME", opname)
	printField("QTH", qth)

	printField("STATION_CALLSIGN", strings.ToUpper(*stationCallsign))

	printField("MY_CITY", stationQth)

	printField("COMMENT", comment)

	fmt.Println("<EOR>")

	return 1, nil
}

func decodeStringField(v string) string {
	return strings.ReplaceAll(v, "_C_", ",")
}

func isEmptyField(v string) bool {
	return v == "" || v == "-"
}

func toUTC(d string, t string, tz string, offset time.Duration) (string, string, error) {
	// https://play.golang.org/p/iIf25Ee8EIO

	if tz != "UTC+08:00" {
		return "", "", errors.New("only supports UTC+08:00 timezone")
	}

	dtStr := fmt.Sprintf("%s %s +0800", d, t)
	tm, err := time.Parse("2/1/2006 1504 -0700", dtStr)

	if err != nil {
		return "", "", err
	}

	tm = tm.Add(offset)
	tu := tm.UTC()
	ud := tu.Format("20060102")
	ut := tu.Format("1504")

	return ud, ut, nil
}

type BandFreqRange struct {
	Band     string
	freqFrom uint64
	freqTo   uint64
}

func freqToBand(freq string) string {
	freqInMHz, err := strconv.ParseFloat(freq, 64)
	if err != nil {
		return ""
	}

	freqInHz := uint64(freqInMHz * 1e6)
	/*
		http://adif.org.uk/311/ADIF_311.htm#Band_Enumeration

		2190m	.1357	.1378
		630m	.472	.479
		560m	.501	.504
		160m	1.8	2.0
		80m	3.5	4.0
		60m	5.06	5.45
		40m	7.0	7.3
		30m	10.1	10.15
		20m	14.0	14.35
		17m	18.068	18.168
		15m	21.0	21.45
		12m	24.890	24.99
		10m	28.0	29.7
		8m	40	45
		6m	50	54
		5m	54.000001	69.9
		4m	70	71
		2m	144	148
		1.25m	222	225
		70cm	420	450
		33cm	902	928
		23cm	1240	1300
		13cm	2300	2450
		9cm	3300	3500
		6cm	5650	5925
		3cm	10000	10500
		1.25cm	24000	24250
		6mm	47000	47200
		4mm	75500	81000
		2.5mm	119980	120020
		2mm	142000	149000
		1mm	241000
	*/

	bandMap := []BandFreqRange{
		{"2190M", 135700, 137800},
		{"630M", 472000, 479000},
		{"560M", 501000, 504000},
		{"160M", 1800000, 2000000},
		{"80M", 3500000, 4000000},
		{"60M", 5060000, 5450000},
		{"40M", 7000000, 7300000},
		{"30M", 10100000, 10150000},
		{"20M", 14000000, 14350000},
		{"17M", 18068000, 18168000},
		{"15M", 21000000, 21450000},
		{"12M", 24890000, 24990000},
		{"10M", 28000000, 29700000},
		{"8M", 40000000, 45000000},
		{"6M", 50000000, 54000000},
		{"5M", 54000001, 69900000},
		{"4M", 70000000, 71000000},
		{"2M", 144000000, 148000000},
		{"1.25M", 222000000, 225000000},
		{"70CM", 420000000, 450000000},
		{"33CM", 902000000, 928000000},
		{"23CM", 1240000000, 1300000000},
		{"13CM", 2300000000, 2450000000},
		{"9CM", 3300000000, 3500000000},
		{"6CM", 5650000000, 5925000000},
		{"3CM", 10000000000, 10500000000},
		{"1.25CM", 24000000000, 24250000000},
		{"6MM", 47000000000, 47200000000},
		{"4MM", 75500000000, 81000000000},
		{"2.5MM", 119980000000, 120020000000},
		{"2MM", 142000000000, 149000000000},
		{"1MM", 241000000000, 999999999999},
	}

	for _, m := range bandMap {
		if freqInHz >= m.freqFrom && freqInHz <= m.freqTo {
			return m.Band
		}
	}

	return ""
}

func printField(fname, fvalue string) {
	if !isEmptyField(fvalue) {
		fmt.Printf("<%s:%v>%s", fname, len(fvalue), fvalue)
	}
}

func printFieldLn(fname, fvalue string) {
	printField(fname, fvalue)
	fmt.Println()
}

func printHeader() {
	fmt.Printf("Generated on %v for %s\n",
		time.Now().Format("2006-01-02 at 15:04:05 MST"),
		strings.ToUpper(*stationCallsign))

	printFieldLn("ADIF_VER", "3.1.1")
	printFieldLn("PROGRAMID", "HamLog-C2A")
	printFieldLn("PROGRAMVERSON", version)
	fmt.Println("<EOH>")
}
