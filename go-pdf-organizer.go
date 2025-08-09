package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Category struct represents a document category with a name and a list of keywords.
type Category struct {
	Name     string
	Keywords []string
}

var (
	verbose     bool
	help        bool
	lang        string
	configPath  string
	execDir     string // Global variable to store the executable's directory.
	matchAll    bool   // New global variable for the "match all keywords" option.
	testOCRFile string // New global variable for the OCR test file path.
)

// main is the entry point of the application. It parses command-line flags and orchestrates the PDF organization or OCR test.
func main() {
	// Define command-line flags for various options.
	flag.BoolVar(&help, "help", false, "Show help message")
	flag.BoolVar(&help, "h", false, "Show help message (shorthand)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose mode (shows OCR output for organization, and for test-ocr)")
	flag.BoolVar(&verbose, "v", false, "Enable verbose mode (shorthand)")
	flag.StringVar(&lang, "lang", "por", "OCR language (e.g., por, eng, spa)")
	flag.StringVar(&lang, "l", "por", "OCR language (shorthand)")
	flag.StringVar(&configPath, "config", "categories.conf", "Path to categories config file")
	flag.StringVar(&configPath, "c", "categories.conf", "Path to categories config file (shorthand)")
	flag.BoolVar(&matchAll, "matchall", false, "Require all keywords of a category to be present for classification")
	flag.BoolVar(&matchAll, "m", false, "Require all keywords (shorthand)")
	flag.StringVar(&testOCRFile, "test-ocr", "", "Path to a specific PDF file to test OCR extraction")
	flag.StringVar(&testOCRFile, "t", "", "Path to a specific PDF file to test OCR extraction (shorthand)")

	var err error
	// Get the directory of the executable to use as the default path and destination for classified files.
	execDir, err = getDefaultPath()
	if err != nil {
		log.Fatal("Error getting executable path:", err)
	}

	pdfPath := flag.String("path", execDir, "Path to PDF folder to organize")
	pdfPathShort := flag.String("p", execDir, "Path to PDF folder (shorthand)")
	flag.Parse()

	// Handle the case where the shorthand path flag is used.
	if *pdfPath == execDir && *pdfPathShort != execDir {
		pdfPath = pdfPathShort
	}

	// If the help flag is set, print the help message and exit.
	if help {
		printHelp()
		return
	}

	// --- OCR Test Logic ---
	// If the test-ocr flag is set, perform an OCR test on the specified file and exit.
	if testOCRFile != "" {
		fmt.Printf("\n=== Testing OCR for: %s ===\n", testOCRFile)
		if _, err := os.Stat(testOCRFile); os.IsNotExist(err) {
			log.Fatalf("Error: File not found for OCR test: %s", testOCRFile)
		}

		content, err := extractTextFromPDF(testOCRFile, lang)
		if err != nil {
			log.Fatalf("Error extracting text from %s: %v", testOCRFile, err)
		}

		fmt.Println("\n--- OCR Extracted Text ---")
		fmt.Println(content)
		fmt.Println("--------------------------")
		fmt.Printf("Extracted %d characters.\n", len(content))
		return // Exit after performing the OCR test.
	}
	// --- End OCR Test Logic ---

	// If verbose mode is enabled, print a summary of the current settings.
	if verbose {
		log.Println("Starting PDF organizer in verbose mode")
		log.Printf("Version: 2.8 (Recursive, keeps unclassified, classified to exec dir, match all option, OCR test option, auto-rename duplicates)")
		log.Printf("Base path: %s", *pdfPath)
		log.Printf("OCR Language: %s", lang)
		log.Printf("Categories config: %s", configPath)
		log.Printf("Executable directory: %s", execDir)
		log.Printf("Match All Keywords: %t", matchAll)
	}

	fmt.Println("\n=== PDF Content Organizer with OCR ===")

	// Load the categories and their keywords from the configuration file.
	categories, err := loadCategories(configPath)
	if err != nil {
		log.Fatal("Error loading categories:", err)
	}

	if verbose {
		log.Printf("Loaded %d categories", len(categories))
	}

	// Start the recursive organization process from the specified path.
	err = organizeRecursively(*pdfPath, categories)
	if err != nil {
		log.Fatal("Organization error:", err)
	}

	fmt.Println("\nOrganization completed successfully!")
}

// getDefaultPath returns the directory where the executable is located.
func getDefaultPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

// printHelp displays the usage instructions and options for the program.
func printHelp() {
	fmt.Println("Usage: pdforganizer [options]")
	fmt.Println("\nOrganizes PDF files by content using OCR and defined categories.")
	fmt.Println("Unclassified documents remain in their original location.")
	fmt.Println("Classified documents and their category folders are moved to the executable's directory.")
	fmt.Println("If a file with the same name already exists at the destination, it will be automatically renamed (e.g., 'file (1).pdf').")
	fmt.Println("\nOptions:")
	fmt.Println("  -path, -p string    Path to PDF folder to organize (default: executable directory)")
	fmt.Println("  -lang, -l string    OCR language (default: por)")
	fmt.Println("  -config, -c string  Path to categories config (default: categories.conf)")
	fmt.Println("  -verbose, -v        Enable verbose mode (shows OCR output)")
	fmt.Println("  -matchall, -m       Require ALL keywords of a category to be present for classification (default: false, matches ANY keyword)")
	fmt.Println("  -test-ocr, -t string Path to a specific PDF file to test OCR extraction and output the text.")
	fmt.Println("  -help, -h           Show help message")
	fmt.Println("\nNote: Keyword matching is case-insensitive")
	fmt.Println("\nRequirements:")
	fmt.Println("  - Tesseract OCR (sudo apt install tesseract-ocr)")
	fmt.Println("  - Portuguese language data (sudo apt install tesseract-ocr-por)")
	fmt.Println("  - Poppler utilities (sudo apt install poppler-utils)")
}

// loadCategories reads a configuration file and parses it into a slice of Category structs.
func loadCategories(configPath string) ([]Category, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	var categories []Category
	var currentCategory Category

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// A line enclosed in brackets indicates a new category.
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if currentCategory.Name != "" {
				categories = append(categories, currentCategory)
			}
			currentCategory = Category{
				Name:     strings.Trim(line, "[]"),
				Keywords: []string{},
			}
		} else if currentCategory.Name != "" {
			// Lines that are not categories are treated as keywords for the current category.
			currentCategory.Keywords = append(currentCategory.Keywords, strings.ToLower(line))
		}
	}

	// Append the last category after the loop finishes.
	if currentCategory.Name != "" {
		categories = append(categories, currentCategory)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	return categories, nil
}

// organizeRecursively walks through a directory and its subdirectories, organizing any PDF files found.
func organizeRecursively(currentPath string, categories []Category) error {
	// Check if the specified path exists.
	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		return fmt.Errorf("specified folder doesn't exist: %s", currentPath)
	}

	// Read the contents of the current directory.
	files, err := ioutil.ReadDir(currentPath)
	if err != nil {
		return err
	}

	// Iterate through each item in the directory.
	for _, file := range files {
		filePath := filepath.Join(currentPath, file.Name())

		// If the item is a directory, call organizeRecursively on it.
		if file.IsDir() {
			if verbose {
				log.Printf("Entering directory: %s", filePath)
			}
			err := organizeRecursively(filePath, categories)
			if err != nil {
				log.Printf("Error processing directory %s: %v", filePath, err)
			}
			continue
		}

		// If the item is a PDF file, process it.
		if strings.ToLower(filepath.Ext(file.Name())) == ".pdf" {
			if verbose {
				log.Printf("\nProcessing file: %s", file.Name())
				log.Printf("Full path: %s", filePath)
				log.Printf("Size: %d bytes", file.Size())
			}

			// Extract text from the PDF using OCR.
			content, err := extractTextFromPDF(filePath, lang)
			if err != nil {
				log.Printf("Error processing %s: %v", file.Name(), err)
				continue
			}

			if verbose {
				log.Println("\nOCR Output:")
				log.Println("----------------------------------------")
				log.Println(content)
				log.Println("----------------------------------------")
				log.Printf("Extracted %d characters", len(content))
			}

			contentLower := strings.ToLower(content)
			// Determine the category of the PDF based on its content.
			categoryName := determineCategory(contentLower, categories, matchAll)

			// If no category is determined, the file remains in its original location.
			if categoryName == "" {
				fmt.Printf("Unclassified: %s (remains in original location)\n", file.Name())
				continue
			}

			if verbose {
				log.Printf("Assigned category: %s", categoryName)
			}

			// Create the destination folder for the category if it doesn't exist.
			categoryPath := filepath.Join(execDir, categoryName)
			if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
				err = os.Mkdir(categoryPath, 0755)
				if err != nil {
					return fmt.Errorf("error creating folder %s in executable directory: %v", categoryName, err)
				}
				if verbose {
					log.Printf("Created category folder: %s", categoryPath)
				}
			}

			// --- Start of Automatic Renaming Logic ---
			// Handle duplicate filenames by renaming them with a counter.
			baseName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			ext := filepath.Ext(file.Name())
			targetFileName := file.Name()
			counter := 0
			foundUniqueName := false

			for !foundUniqueName {
				newPath := filepath.Join(categoryPath, targetFileName)
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					// The new path does not exist, so it's a unique name.
					err = os.Rename(filePath, newPath)
					if err != nil {
						return fmt.Errorf("error moving %s to %s: %v", file.Name(), newPath, err)
					}
					fmt.Printf("Organized: %s → %s\n", file.Name(), newPath)
					foundUniqueName = true
				} else if err != nil {
					// An error occurred while checking the file, other than not existing.
					return fmt.Errorf("error checking destination file %s: %v", newPath, err)
				} else {
					// The file already exists, generate a new name.
					counter++
					targetFileName = fmt.Sprintf("%s (%d)%s", baseName, counter, ext)
					if verbose {
						log.Printf("Duplicate found, trying new name: %s", targetFileName)
					}
				}
			}
			// --- End of Automatic Renaming Logic ---
		}
	}

	return nil
}

// extractTextFromPDF uses external tools (pdftoppm and tesseract) to perform OCR on the first page of a PDF file.
func extractTextFromPDF(pdfPath, language string) (string, error) {
	// Create a temporary directory for intermediate files.
	tempDir, err := ioutil.TempDir("", "pdfocr")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Ensure the temporary directory is cleaned up.

	// Use pdftoppm to convert the first page of the PDF to a PNG image.
	outputPrefix := filepath.Join(tempDir, "page")
	cmd := exec.Command("pdftoppm", "-png", "-f", "1", "-l", "1", pdfPath, outputPrefix)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("pdftoppm error: %v, %s", err, stderr.String())
	}

	// Find the generated PNG file.
	pngFiles, err := filepath.Glob(filepath.Join(tempDir, "page-*.png"))
	if err != nil || len(pngFiles) == 0 {
		return "", fmt.Errorf("no PNG files generated")
	}
	pngPath := pngFiles[0]

	// Use tesseract to extract text from the PNG image.
	cmd = exec.Command("tesseract", pngPath, "stdout", "-l", language, "--psm", "3")
	cmd.Stderr = &stderr
	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("tesseract error: %v, %s", err, stderr.String())
	}

	return out.String(), nil
}

// determineCategory checks the OCR-extracted text against category keywords to find a match.
func determineCategory(contentLower string, categories []Category, matchAll bool) string {
	for _, category := range categories {
		if matchAll {
			// "Match all" logic: all keywords for a category must be present.
			allKeywordsFound := true
			for _, keyword := range category.Keywords {
				if !strings.Contains(contentLower, keyword) {
					allKeywordsFound = false
					break
				}
			}
			if allKeywordsFound {
				return category.Name
			}
		} else {
			// "Match any" logic: at least one keyword must be present.
			for _, keyword := range category.Keywords {
				if strings.Contains(contentLower, keyword) {
					return category.Name
				}
			}
		}
	}
	return "" // Return an empty string if no category matches.
}
