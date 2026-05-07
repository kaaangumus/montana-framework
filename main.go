package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Exploit struct {
	ID           int      `json:"exploit_id"`
	Date         string   `json:"date"`
	Category     string   `json:"category"`
	Platform     string   `json:"platform"`
	Author       string   `json:"author"`
	CVE          []string `json:"cve"`
	Title        string   `json:"title"`
	OriginalLink string   `json:"original_link"`
}

type options struct {
	indexPath string
	query     string
	cve       string
	category  string
	platform  string
	id        int
	limit     int
}

func main() {
	opts := parseFlags()

	db, err := loadIndex(opts.indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch {
	case opts.id > 0:
		printByID(db, opts.id)
	case opts.cve != "":
		printResults(filter(db, []string{opts.cve}, opts), opts.limit)
	case opts.query != "":
		printResults(filter(db, strings.Fields(opts.query), opts), opts.limit)
	case opts.category != "" || opts.platform != "":
		printResults(filter(db, nil, opts), opts.limit)
	default:
		printUsage()
	}
}

func parseFlags() options {
	var opts options
	flag.StringVar(&opts.indexPath, "index", defaultIndexPath(), "path to index.json")
	flag.StringVar(&opts.query, "q", "", "search query, for example: wordpress rce")
	flag.StringVar(&opts.cve, "cve", "", "search by CVE, for example: CVE-2021-41773")
	flag.StringVar(&opts.category, "category", "", "filter by category")
	flag.StringVar(&opts.platform, "platform", "", "filter by platform")
	flag.IntVar(&opts.id, "id", 0, "show one record by exploit id")
	flag.IntVar(&opts.limit, "limit", 50, "maximum number of results")
	flag.Parse()

	if opts.query == "" && flag.NArg() > 0 {
		opts.query = strings.Join(flag.Args(), " ")
	}

	return opts
}

func defaultIndexPath() string {
	if envPath := os.Getenv("MONTANA_INDEX"); envPath != "" {
		return envPath
	}

	if exe, err := os.Executable(); err == nil {
		installed := filepath.Join(filepath.Dir(exe), "index.json")
		if fileExists(installed) {
			return installed
		}
	}

	sharePath := filepath.Join("/usr/local/share/montana", "index.json")
	if fileExists(sharePath) {
		return sharePath
	}

	return "index.json"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadIndex(path string) ([]Exploit, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", path, err)
	}
	defer file.Close()

	var db []Exploit
	if err := json.NewDecoder(file).Decode(&db); err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", path, err)
	}

	sort.Slice(db, func(i, j int) bool {
		return db[i].ID < db[j].ID
	})

	return db, nil
}

func filter(db []Exploit, terms []string, opts options) []Exploit {
	var results []Exploit
	for _, item := range db {
		if opts.category != "" && !containsFold(item.Category, opts.category) {
			continue
		}
		if opts.platform != "" && !containsFold(item.Platform, opts.platform) {
			continue
		}
		if matchesTerms(item, terms) {
			results = append(results, item)
		}
	}
	return results
}

func matchesTerms(item Exploit, terms []string) bool {
	if len(terms) == 0 {
		return true
	}

	haystack := strings.ToLower(strings.Join([]string{
		strconv.Itoa(item.ID),
		item.Date,
		item.Category,
		item.Platform,
		item.Author,
		strings.Join(item.CVE, " "),
		item.Title,
		item.OriginalLink,
	}, " "))

	for _, term := range terms {
		if !strings.Contains(haystack, strings.ToLower(term)) {
			return false
		}
	}
	return true
}

func containsFold(value, term string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(term))
}

func printByID(db []Exploit, id int) {
	for _, item := range db {
		if item.ID == id {
			printDetail(item)
			return
		}
	}
	fmt.Printf("no exploit metadata found for id %d\n", id)
}

func printResults(results []Exploit, limit int) {
	if limit <= 0 {
		limit = 50
	}
	if len(results) == 0 {
		fmt.Println("no matching exploit metadata found")
		return
	}

	if len(results) < limit {
		limit = len(results)
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDATE\tPLATFORM\tCATEGORY\tCVE\tTITLE")
	for _, item := range results[:limit] {
		fmt.Fprintf(
			writer,
			"%d\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Date,
			emptyDash(item.Platform),
			emptyDash(item.Category),
			emptyDash(strings.Join(item.CVE, ",")),
			item.Title,
		)
	}
	writer.Flush()

	if len(results) > limit {
		fmt.Printf("\nshowing %d of %d results; increase -limit to show more\n", limit, len(results))
	} else {
		fmt.Printf("\n%d result(s)\n", len(results))
	}
}

func printDetail(item Exploit) {
	fmt.Printf("ID:       %d\n", item.ID)
	fmt.Printf("Date:     %s\n", emptyDash(item.Date))
	fmt.Printf("Title:    %s\n", emptyDash(item.Title))
	fmt.Printf("Platform: %s\n", emptyDash(item.Platform))
	fmt.Printf("Category: %s\n", emptyDash(item.Category))
	fmt.Printf("Author:   %s\n", emptyDash(item.Author))
	fmt.Printf("CVE:      %s\n", emptyDash(strings.Join(item.CVE, ", ")))
	fmt.Printf("Source:   %s\n", emptyDash(item.OriginalLink))
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func printUsage() {
	fmt.Println("Montana is a local exploit metadata search tool.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  montana -q \"wordpress rce\"")
	fmt.Println("  montana apache 2.4")
	fmt.Println("  montana -cve CVE-2021-41773")
	fmt.Println("  montana -platform linux -category remote -q openssh")
	fmt.Println("  montana -id 33814")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}
