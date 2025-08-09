# GO-PDF-Organizer

A command-line tool written in Go that automatically organizes your PDF files into categorized folders based on their content. It uses OCR (Optical Character Recognition) to extract text from the documents and matches the text against a configurable list of keywords.

## Features

- **Content-Based Classification**: Uses Tesseract OCR to read the content of PDF files.
- **Configurable Categories**: You can define your own categories and keywords in a simple `categories.conf` file.
- **Recursive Organization**: Scans a specified directory and all its subdirectories for PDF files.
- **Automatic Folder Creation**: Creates category folders automatically in the executable's directory.
- **Duplicate Handling**: Automatically renames files with duplicate names (e.g., `invoice (1).pdf`).
- **OCR Test Mode**: A dedicated flag (`-test-ocr`) to test the OCR functionality on a single PDF file and view the extracted text.
- **Verbose Mode**: Provides detailed logging of the organization process, including OCR output.
- **Flexible Matching**: Choose between matching any keyword in a category or requiring all keywords to be present for a classification.
- **Unclassified Files**: Files that do not match any category remain in their original location.

## Requirements

The program relies on external command-line tools to perform the OCR process. Please ensure these are installed on your system:

- **Tesseract OCR**: The primary OCR engine.
- **Tesseract Language Data**: You need to install the language data for the languages you plan to use (e.g., `tesseract-ocr-por` for Portuguese).
- **Poppler Utilities**: Specifically `pdftoppm`, which converts PDF pages into images for Tesseract to process.

You can install these on a Debian-based system with the following command:

```bash
sudo apt install tesseract-ocr tesseract-ocr-por poppler-utils
````

## Installation

### From Source

1.  Clone the repository:
    ```bash
    git clone https://github.com/nfcg/go-pdf-organizer.git
    cd go-pdf-organizer
    ```
2.  Build the executable:
    ```bash
    go build go-pdf-organizer.go
    ```
    This will create a `go-pdf-organizer` executable in your current directory.

### Configuration

The program uses a `categories.conf` file to define the classification rules. The format is straightforward:

  - Category names are enclosed in square brackets `[]`.
  - Keywords for each category are listed on new lines.
  - Lines starting with `#` are treated as comments.

Example `categories.conf`:

```ini
# This is a comment.
# The `[]` defines a new category.
[Invoices]
invoice
fatura
pagamento
compra

[Receipts]
receipt
recibo
```

Place this file in the same directory as the `go-pdf-organizer` executable.

## Usage

Run the program from the command line with various flags to control its behavior.

### Organizing a Folder

To organize all PDFs in the current directory and its subdirectories:

```bash
./go-pdf-organizer
```

To specify a different path:

```bash
./go-pdf-organizer -path /path/to/your/pdf/folder
```

### Options
**Flags**:

  * `-p, -path`: Path to the folder containing the PDFs to organize. (default: Executable's directory)
  * `-l, -lang`: OCR language code (e.g., `por`, `eng`, `spa`). (default: `por`)
  * `-c, -config`: Path to the categories configuration file. (default: `categories.conf`)
  * `-v, -verbose`: Enable verbose mode to see detailed OCR output. (default: `false`)
  * `-m, -matchall`: Require all keywords of a category to be present for classification. By default, it matches any single keyword. (default: `false`)
  * `-t, -test-ocr`: Path to a single PDF file to test OCR extraction and print the output. The program will exit after this.
  * `-h, -help`: Show the help message and exit.

### Example: OCR Test

To see what text the program extracts from a specific PDF, use the `-test-ocr` flag:

```bash
./go-pdf-organizer -test-ocr "path/to/your/document.pdf" -lang eng
```

This will print the extracted text directly to your console.

## How It Works

The program operates in the following steps:

1.  **Flag Parsing**: Reads command-line arguments to configure the run (e.g., path, language, verbosity).
2.  **Category Loading**: Parses the `categories.conf` file into an in-memory data structure.
3.  **Recursive File Walk**: Traverses the specified directory tree, looking for files with a `.pdf` extension.
4.  **Text Extraction**: For each PDF, it uses `pdftoppm` to convert the first page to a PNG image, then uses `tesseract` to perform OCR on the image and extract the text.
5.  **Categorization**: The extracted text is converted to lowercase and checked against the keywords of each defined category.
6.  **File Movement**: If a category match is found, the PDF is moved to a new folder named after the category, located in the executable's directory.
7.  **Error Handling**: Any errors during the process (e.g., file not found, OCR failure) are logged, but the program continues to process other files.


## Contributing

Contributions are welcome\! If you have suggestions for improvements or new features, please open an issue or submit a pull request.
