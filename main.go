package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"os"

	"github.com/kovetskiy/godocs"
	"github.com/reconquest/ser-go"
	"github.com/seletskiy/hierr"
)

var (
	version = "[manual build]"
	usage   = "poke " + version + `

poke is summoned for analysing MySQL slow query logs, poke reads time and
query_time fields and adds additional field time_start (time - query_time);

poke outputs records in JSON format only.

Usage:
    poke [options]
    poke -h | --help
    poke --version

Options:
    -f --file <path>  Specify file location to read. [default: /dev/stdin]
    -s --sort <path>  Specify sort rules.
                       You can specify two or more rules as comma-separated
                       list, like as following: rows_read:desc,rows_sent:asc
                       [default: time_start:asc]
    -h --help         Show this screen.
    --version         Show version.
`
)

var (
	rules = map[string]string{
		"Time":                  `datetime`,
		"Schema":                `string`,
		"Query_time":            `time`,
		"Lock_time":             `time`,
		"Rows_sent":             `int`,
		"Rows_examined":         `int`,
		"Rows_affected":         `int`,
		"Rows_read":             `int`,
		"Bytes_sent":            `int`,
		"Tmp_tables":            `int`,
		"Tmp_disk_tables":       `int`,
		"Tmp_table_sizes":       `int`,
		"QC_Hit":                `bool`,
		"Full_scan":             `bool`,
		"Full_join":             `bool`,
		"Tmp_table":             `bool`,
		"Tmp_table_on_disk":     `bool`,
		"Filesort":              `bool`,
		"Filesort_on_disk":      `bool`,
		"Merge_passes":          `int`,
		"InnoDB_IO_r_ops":       `int`,
		"InnoDB_IO_r_bytes":     `int`,
		"InnoDB_IO_r_wait":      `time`,
		"InnoDB_rec_lock_wait":  `time`,
		"InnoDB_queue_wait":     `time`,
		"InnoDB_pages_distinct": `int`,
	}

	regexps = map[string]*regexp.Regexp{}
)

type Record map[string]interface{}

func compileRegexps() {
	for key, rule := range rules {
		var data string

		switch rule {
		case "datetime":
			data = `.*`
		case "string":
			data = `\w+`
		case "time":
			data = `[0-9\.]+`
		case "int":
			data = `\d+`
		case "bool":
			data = `\w+`
		default:
			panic("uknown rule: " + rule)
		}

		regexps[key] = regexp.MustCompile(
			`^# .*` + key + `: (` + data + `)`,
		)
	}
}

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	compileRegexps()

	file, err := os.Open(args["--file"].(string))
	if err != nil {
		hierr.Fatalf(
			err, "can't open file: %s", args["--file"].(string),
		)
	}

	var (
		reader  = bufio.NewReader(file)
		record  = Record{}
		records = []Record{}
	)

	var line string
	for {
		data, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}

			hierr.Fatalf(
				err, "can't read input data",
			)
		}

		if isPrefix {
			line += string(data)

			continue
		}

		line = string(data)

		if strings.HasPrefix(line, "# Time: ") {
			if len(record) > 0 {
				if record, ok := process(record); ok {
					records = append(records, record)
				}
			}

			record = Record{}
		}

		err = unmarshal(line, record)
		if err != nil {
			log.Println(err)
		}
	}

	if record, ok := process(record); ok {
		records = append(records, record)
	}

	for index, record := range records {
		records[index] = prepare(record)
	}

	sorts := strings.Split(args["--sort"].(string), ",")

	for _, sorting := range sorts {
		parts := strings.Split(sorting, ":")
		if len(parts) != 2 {
			hierr.Fatalf(
				err, "--sort flag has invalid syntax, "+
					"should be key:value,key2:value2",
			)
		}

		key, value := parts[0], parts[1]

		sorter := &sorter{
			records: records,
			key:     key,
			desc:    strings.ToLower(value) == "desc",
		}

		sort.Sort(sorter)
	}

	data, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		hierr.Fatalf(
			err, "unable to encode records to JSON",
		)
	}

	fmt.Println(string(data))
}

func process(record Record) (Record, bool) {
	if timeEnd, ok := record["time"].(time.Time); ok {
		if queryTime, ok := record["query_time"].(time.Duration); ok {
			record["time_start"] = timeEnd.Add(queryTime * -1)
		}
	} else {
		return record, false
	}

	if query, ok := record["query"].(string); ok {
		record["query_length"] = len(query)
	}

	return record, true
}

func prepare(record Record) Record {
	for key, value := range record {
		switch value := value.(type) {
		case time.Time:
			record[key] = value.Format("2006-01-02 15:04:05.00000000")

		case time.Duration:
			record[key] = value.Seconds()
		}
	}

	return record
}

func unmarshal(line string, record Record) error {
	if !strings.HasPrefix(line, "# ") {
		_, ok := record["query"]
		if ok {
			record["query"] = record["query"].(string) + line
			return nil
		}

		record["query"] = line
	}

	for key, rule := range rules {
		raw, ok := match(line, key)
		if !ok {
			continue
		}

		value, err := parse(raw, key, rule)
		if err != nil {
			return ser.Errorf(
				err, "unable to parse %s: %s",
				key, raw,
			)
		}

		record[strings.ToLower(key)] = value
	}

	return nil
}

func match(data, key string) (string, bool) {
	matches := regexps[key].FindStringSubmatch(data)
	if len(matches) > 0 {
		return matches[1], true
	}

	return "", false
}

func parse(raw, key, rule string) (interface{}, error) {
	switch rule {
	case "datetime":
		return time.Parse("060102 15:04:05.0000000000", raw)
	case "time":
		return time.ParseDuration(raw + "s")
	case "string":
		return raw, nil
	case "int":
		return strconv.ParseInt(raw, 10, 64)
	case "bool":
		switch raw {
		case "Yes":
			return true, nil
		case "No":
			return false, nil
		default:
			return false, errors.New("invalid syntax: expected Yes or No")
		}
	}

	return nil, nil
}
