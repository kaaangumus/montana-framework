package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/peterh/liner"
)

var bannerFrames = []string{
	`
[32m M O N T A N A   F R A M E W O R K[0m
`,
	`
[32m M O N T 4 N A   F R 4 M E W O R K[0m
`,
	`
[32m M 0 N T A N A   F R A M E W 0 R K[0m
`,
	`
[32m M O N T A N A   F R A M E W O R K[0m
`,
}

func showBannerAnimation() {
	for i := 0; i < 2; i++ {
		for _, frame := range bannerFrames {
			clearScreen()
			fmt.Print(frame)
			time.Sleep(100 * time.Millisecond)
		}
	}
	clearScreen()
}

// --- Data Structures ---

// Exploit defines the structure for an entry in our exploit database.
type Exploit struct {
	ID       int      `json:"exploit_id"`
	Date     string   `json:"date"`
	Category string   `json:"category"`
	Platform string   `json:"platform"`
	Author   string   `json:"author"`
	CVE      []string `json:"cve"`
	Title    string   `json:"title"`
	OriginalLink string `json:"original_link"`
	Source   string
}

// NmapRun defines the structure for Nmap's XML output.
type NmapRun struct {
	Hosts []Host `xml:"host"`
}
type Host struct {
	Addresses []Address `xml:"address"`
	Ports     []Port    `xml:"ports>port"`
}
type Address struct {
	Addr string `xml:"addr,attr"`
	Type string `xml:"addrtype,attr"`
}
type Port struct {
	PortID   string  `xml:"portid,attr"`
	Protocol string  `xml:"protocol,attr"`
	Service  Service `xml:"service"`
}
type Service struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}



// --- Global Variables ---

var (
	exploitDB     []Exploit
	activeExploit *Exploit
	dataDir       string
)

const nmapOutputFile = "scan_result.xml"
var historyFile = filepath.Join(os.TempDir(), ".montana_history")

// --- Main Application Flow ---

func main() {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	if strings.Contains(executable, "/usr/local/bin") {
		dataDir = "/usr/local/share/montana-framework"
	} else {
		dataDir = "."
	}

	showBannerAnimation()
	loadExploits()
	runShell()
}

func loadExploits() {
	jsonPath := filepath.Join(dataDir, "index.json")
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		fmt.Println("[!] Error: index.json not found.")
		return
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &exploitDB)

	for i := range exploitDB {
		if strings.Contains(exploitDB[i].OriginalLink, "0day.today") {
			exploitDB[i].Source = "0day.today"
		}
	}

	fmt.Printf("[*] Loaded %d exploits into memory.\n\n", len(exploitDB))
}

func runShell() {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		prompt := "montana > "
		if activeExploit != nil {
			title := activeExploit.Title
			if len(title) > 20 {
				title = title[:17] + "..."
			}
			prompt = fmt.Sprintf("montana(%s) > ", title)
		}

		input, err := line.Prompt(prompt)
		if err != nil {
			if err == io.EOF {
				break // Exit on Ctrl+D
			}
			fmt.Printf("[!] Error reading input: %v\n", err)
			continue
		}

		if input == "" {
			continue
		}

		line.AppendHistory(input)
		handleCommand(input)
	}

	if f, err := os.Create(historyFile); err != nil {
		fmt.Printf("[!] Error writing history file: %v\n", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}

// --- Command Handling ---

func handleCommand(command string) {
	parts := strings.Fields(command)
	cmd := parts[0]

	if activeExploit != nil {
		handleExploitContextCommands(cmd, parts)
		return
	}

	handleMainCommands(cmd, parts)
}

func handleMainCommands(cmd string, parts []string) {
	switch cmd {
	case "help":
		printHelp()
	case "list":
		listExploits()
	case "search":
		if len(parts) < 2 {
			fmt.Println("Usage: search <keyword1> ...")
			return
		}
		searchAndPrint(parts[1:])
	case "use":
		if len(parts) < 2 {
			fmt.Println("Usage: use <exploit_id>")
			return
		}
		useExploit(parts[1])
	case "nmap":
		if len(parts) < 2 {
			fmt.Println("Usage: nmap <nmap args> <target>")
			return
		}
		runNmapScan(parts[1:])
	case "suggest":
		suggestExploits()
	case "nmap_results":
		listNmapResults(parts)
	case "exit":
		os.Exit(0)
	case "clear":
		clearScreen()
	case "os":
		runOsShell()
	default:
		fmt.Printf("Unknown command: %s. Type 'help'.\n", cmd)
	}
}

func runOsShell() {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	fmt.Println("\033[33mEntering OS mode. Type 'exit' to return.\033[0m")

	for {
		input, err := line.Prompt("os > ")
		if err != nil {
			if err == io.EOF {
				break // Exit on Ctrl+D
			}
			fmt.Printf("[!] Error reading input: %v\n", err)
			continue
		}

		if input == "" {
			continue
		}

		if input == "exit" {
			break
		}

		line.AppendHistory(input)

		parts := strings.Fields(input)
		if len(parts) > 0 && parts[0] == "ls" {
			parts = append(parts, "--color=always")
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Printf("\n[!] Command failed: %v\n", err)
		}
	}
	fmt.Println("Exiting OS mode.")
}

func clearScreen() {
	cmd := exec.Command("clear") // For Linux/macOS
	if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		cmd = exec.Command("cmd", "/c", "cls") // For Windows
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func handleExploitContextCommands(cmd string, parts []string) {
	switch cmd {
	case "info":
		showExploitInfo()
	case "get":
		copyExploitFile()
	case "back":
		activeExploit = nil
	case "help":
		printHelp()
	case "exit":
		os.Exit(0)
	case "os":
		runOsShell()
	default:
		fmt.Printf("Unknown command in exploit context: %s.\n", cmd)
	}
}

func printHelp() {
	if activeExploit != nil {
		fmt.Println("\n\033[33m--- Exploit Context Commands ---\033[0m")
		fmt.Println("  info          - Shows details of the selected exploit.")
		fmt.Println("  get           - Copies the exploit file to the current directory.")
		fmt.Println("  back          - Deselects the current exploit.")
		fmt.Println("  help          - Shows this help message.")
		fmt.Println("  exit          - Exits the framework.")
		fmt.Println("  os            - Enters OS shell mode.")
		fmt.Println()
		return
	}

	fmt.Println("\n\033[33m--- Workspace Commands ---\033[0m")
	fmt.Println("  nmap <args>   - Runs Nmap and saves results to an XML file.")
	fmt.Println("  suggest       - Suggests exploits based on the last Nmap scan.")
	fmt.Println("  nmap_results  - Lists all nmap scan results.")
	fmt.Println("\n\033[33m--- Exploit Database Commands ---\033[0m")
	fmt.Println("  search <k>    - Searches for exploits by keyword(s).")
	fmt.Println("  list          - Lists all available exploits.")
	fmt.Println("  use <id>      - Selects an exploit to interact with.")
	fmt.Println("\n\033[33m--- General Commands ---\033[0m")
	fmt.Println("  help          - Shows this help message.")
	fmt.Println("  exit          - Exits the framework.")
	fmt.Println("  clear         - Clears the screen.")
	fmt.Println("  os            - Enters OS shell mode.")
	fmt.Println()
}

// --- Workspace & Nmap Functions ---

const nmapScansDir = "nmap_scans"

func runNmapScan(args []string) {
	os.MkdirAll(nmapScansDir, os.ModePerm)
	files, _ := ioutil.ReadDir(nmapScansDir)
	scanNumber := len(files) + 1
	nmapOutputFile := filepath.Join(nmapScansDir, fmt.Sprintf("scan%d.xml", scanNumber))

	fmt.Printf("[*] Starting Nmap scan... Command: nmap -oX %s %s\n", nmapOutputFile, strings.Join(args, " "))
	cmdArgs := append([]string{"-oX", nmapOutputFile}, args...)
	cmd := exec.Command("nmap", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("\n[!] Nmap scan failed: %v\n", err)
		return
	}
	fmt.Println("\n[*] Nmap scan finished. Results saved to", nmapOutputFile)
}


func listNmapResults(parts []string) {
	if len(parts) == 1 {
		files, err := ioutil.ReadDir(nmapScansDir)
		if err != nil {
			fmt.Printf("[!] Could not read nmap scans directory: %v\n", err)
			return
		}
		fmt.Println("\n--- Saved Nmap Scans ---")
		for _, file := range files {
			fmt.Println(file.Name())
		}
		fmt.Println()
	} else {
		scanFile := filepath.Join(nmapScansDir, parts[1])
		content, err := ioutil.ReadFile(scanFile)
		if err != nil {
			fmt.Printf("[!] Could not read scan file: %v\n", err)
			return
		}
		fmt.Println(string(content))
	}
}

func suggestExploits() {
	files, err := ioutil.ReadDir(nmapScansDir)
	if err != nil {
		fmt.Printf("[!] Could not read nmap scans directory: %v\n", err)
		return
	}
	if len(files) == 0 {
		fmt.Println("[!] No nmap scans found. Run an nmap scan first.")
		return
	}

	latestScan := files[len(files)-1].Name()
	scanFile := filepath.Join(nmapScansDir, latestScan)

	xmlFile, err := os.Open(scanFile)
	if err != nil {
		fmt.Printf("[!] Could not open Nmap results file: %v\n", err)
		return
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)
	var nmapRun NmapRun
	xml.Unmarshal(byteValue, &nmapRun)

	fmt.Println("\n[*] Analyzing services from latest Nmap scan for exploit suggestions...")
	foundSomething := false
	for _, host := range nmapRun.Hosts {
		for _, port := range host.Ports {
			if port.Service.Product != "" {
				keywords := strings.Fields(strings.ToLower(port.Service.Product + " " + port.Service.Version))
				if len(keywords) == 0 { continue }

				results := searchExploitsInternal(keywords)
				if len(results) > 0 {
					foundSomething = true
					fmt.Printf("\n[+] Suggestions for %s %s:\n", port.Service.Product, port.Service.Version)
					for _, exp := range results {
						fmt.Printf("    - ID: % -10d Title: %s\n", exp.ID, exp.Title)
					}
				}
			}
		}
	}

	if !foundSomething {
		fmt.Println("\n[*] No matching exploits found in the database for the discovered services.")
	}
}


// --- Exploit Handling Functions ---

func getUniqueCategories() []string {
	categorySet := make(map[string]struct{})
	for _, exp := range exploitDB {
		categorySet[exp.Category] = struct{}{}
	}

	categories := make([]string, 0, len(categorySet))
	for category := range categorySet {
		categories = append(categories, category)
	}
	return categories
}

func listExploits() {
	if len(exploitDB) == 0 {
		return
	}

	// 1. Get and display categories
	categories := getUniqueCategories()
	fmt.Println("\nAvailable Categories:")
	for _, category := range categories {
		fmt.Printf("- %s\n", category)
	}
	fmt.Println("- all")

	// 2. Prompt for category
	line := liner.NewLiner()
	defer line.Close()
	selectedCategory, err := line.Prompt("\nEnter a category to list (default: all): ")
	if err != nil {
		fmt.Println("\n[!] Error reading input:", err)
		return
	}

	// 3. Filter exploits
	var filteredExploits []Exploit
	if selectedCategory == "all" || selectedCategory == "" {
		filteredExploits = exploitDB
	} else {
		for _, exp := range exploitDB {
			if strings.EqualFold(exp.Category, selectedCategory) {
				filteredExploits = append(filteredExploits, exp)
			}
		}
	}

	if len(filteredExploits) == 0 {
		fmt.Println("\nNo exploits found for that category.")
		return
	}

	// 4. Prepare output for pager
	var output strings.Builder
	output.WriteString("\nID\t\tTitle\t\t\t\t\t\t\t\tCategory\n")
	output.WriteString("-----------------------------------------------------------------------------------------\n")
	for _, exp := range filteredExploits {
		title := exp.Title
		if len(title) > 45 {
			title = title[:42] + "..."
		}
		output.WriteString(fmt.Sprintf("% -15d % -50s % -15s\n", exp.ID, title, exp.Category))
	}
	output.WriteString("\n")

	// 5. Pipe to pager
	pagerCmd := exec.Command("less")
	pagerIn, err := pagerCmd.StdinPipe()
	if err != nil {
		fmt.Println("[!] Could not create pager pipe:", err)
		// Fallback to just printing
		fmt.Print(output.String())
		return
	}

	pagerCmd.Stdout = os.Stdout
	pagerCmd.Stderr = os.Stderr

	if err := pagerCmd.Start(); err != nil {
		fmt.Println("[!] Could not start pager:", err)
		// Fallback to just printing
		fmt.Print(output.String())
		return
	}

	_, err = io.WriteString(pagerIn, output.String())
	if err != nil {
		fmt.Println("[!] Could not write to pager:", err)
	}
	pagerIn.Close()

	pagerCmd.Wait()
}

func searchAndPrint(keywords []string) {
	results := searchExploitsInternal(keywords)
	if len(results) == 0 {
		fmt.Println("No exploits found matching your criteria.")
		return
	}

	fmt.Printf("\nFound %d matching exploits:\n", len(results))
	fmt.Println("\nID\t\tTitle\t\t\t\t\t\t\t\tCategory")
	fmt.Println("-----------------------------------------------------------------------------------------")
	for _, exp := range results {
		title := exp.Title
		if len(title) > 45 {
			title = title[:42] + "..."
		}
		fmt.Printf("% -15d % -50s % -15s\n", exp.ID, title, exp.Category)
	}
	fmt.Println()
}

func searchExploitsInternal(keywords []string) []Exploit {
	var results []Exploit
	for _, exp := range exploitDB {
		searchableText := strings.ToLower(fmt.Sprintf("%s %s %s %s", exp.Title, exp.Platform, exp.Author, strings.Join(exp.CVE, " ")))
		match := true
		for _, keyword := range keywords {
			if !strings.Contains(searchableText, strings.ToLower(keyword)) {
				match = false
				break
			}
		}
		if match {
			results = append(results, exp)
		}
	}
	return results
}

func useExploit(idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid exploit ID.")
		return
	}

	for i := range exploitDB {
		if exploitDB[i].ID == id {
			activeExploit = &exploitDB[i]
			fmt.Printf("[*] Exploit selected: %d - %s\n", activeExploit.ID, activeExploit.Title)
			return
		}
	}
	fmt.Println("[!] Exploit not found.")
}

func showExploitInfo() {
	if activeExploit == nil {
		return
	}
	path := filepath.Join(dataDir, "exploits", activeExploit.Category, fmt.Sprintf("%d.txt", activeExploit.ID))
	fmt.Println("\n--- Exploit Details ---")
	fmt.Printf("ID:       %d\n", activeExploit.ID)
	fmt.Printf("Source:   %s\n", activeExploit.Source)
	fmt.Printf("Title:    %s\n", activeExploit.Title)
	fmt.Printf("Author:   %s\n", activeExploit.Author)
	fmt.Printf("Date:     %s\n", activeExploit.Date)
	fmt.Printf("Category: %s\n", activeExploit.Category)
	fmt.Printf("Platform: %s\n", activeExploit.Platform)
	fmt.Printf("CVE:      %s\n", strings.Join(activeExploit.CVE, ", "))
	fmt.Printf("Path:     %s\n", path)
	fmt.Println()
}

func copyExploitFile() {
	if activeExploit == nil {
		return
	}
	sourcePath := filepath.Join(dataDir, "exploits", activeExploit.Category, fmt.Sprintf("%d.txt", activeExploit.ID))
	destPath := fmt.Sprintf("%d.txt", activeExploit.ID)

	source, err := os.Open(sourcePath)
	if err != nil {
		fmt.Printf("[!] Error opening source file: %v\n", err)
		return
	}
	defer source.Close()

	destination, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("[!] Error creating destination file: %v\n", err)
		return
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		fmt.Printf("[!] Error copying file: %v\n", err)
		return
	}

	fmt.Printf("[*] Exploit file copied to: %s\n", destPath)
}