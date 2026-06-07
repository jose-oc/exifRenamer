# Image Exif Renamer Specification

This document details the specification for a Go-based Command Line Interface (CLI) tool designed to replicate the file renaming and organization behavior of the macOS application **ExifRenamer** based on a specific user configuration.

## 1. Goal
Create a Go CLI tool using **Cobra** and **Viper** that reads image metadata (EXIF data) and automatically renames and reorganizes (moves) images into a structured directory tree according to defined formatting rules.

---

## 2. Configuration Specifications

### A. Base Destination Folder
* **Default Directory:** `$HOME/Pictures/Sorted`
* **Configuration:** Must be fully configurable via:
  1. CLI Flags (e.g., `--dest` / `-d`)
  2. YAML Configuration file (managed by Viper)

### B. Directory Hierarchy & Organization
Files are grouped and moved into nested subfolders dynamically created based on the photo's creation date (extracted from EXIF metadata):
* **Folder Structure Pattern:** `[Year]/[Year]-[Month]/[Year]-[Month]-[Day]-[Suffix]`
* **Format String:** `%P%Y/%Y-%M/%Y-%M-%D%S` *(or standard pattern, see below)*
* **Components:**
  * `%P` (Prefix): Optional prefix (defaults to empty).
  * `%Y` (Year): 4-digit year (e.g., `2023`).
  * `%M` (Month): 2-digit month (e.g., `08`).
  * `%D` (Day): 2-digit day (e.g., `03`).
  * `%S` (Suffix): Suffix appended to the leaf folder name (e.g., `Playa`, resulting in `-Playa` when preceded by a hyphen).
* **Example Path:**
  `/Users/joseortizcastano/Pictures/Sorted/2023/2023-08/2023-08-03-Playa/`

### C. File Renaming Conventions
* **Format String:** `%P%Y%M%D-%h%m%s%S%C%F` *(or standard pattern, see below)*
* **Components:**
  * `%P` (Prefix): Optional prefix (defaults to empty).
  * `%Y%M%D` (Date): 4-digit Year, 2-digit Month, 2-digit Day (e.g., `20230803`).
  * `-` (Separator): Literal hyphen.
  * `%h%m%s` (Time): Hour, Minute, Second in 24-hour format (e.g., `194000`).
  * `%S` (Suffix): Suffix appended to the filename (e.g., `Playa`).
  * `%C` (Collision Counter): If a file with the same name already exists in the destination folder, an incrementing counter (`_1`, `_2`, etc.) is appended to ensure uniqueness.
  * `%F` (Extension): The file extension (e.g., `.jpg`, `.jpeg`, `.png`, `.heic`). The tool must support a configuration option (via config file or CLI flag) to change the extension to **lowercase** or preserve the original case.
* **Example Filename:** `20230803-194000-Playa.jpg`

---

## 3. Pattern / Format Token Standards
To follow common standards (similar to `strftime`), we can support the following placeholder mapping for format strings:

| Standard Token | ExifRenamer Token | Description | Example |
|---|---|---|---|
| `%Y` | `%Y` | 4-digit Year | `2023` |
| `%y` | — | 2-digit Year | `23` |
| `%m` | `%M` | 2-digit Month | `08` |
| `%d` | `%D` | 2-digit Day | `03` |
| `%H` | `%h` | 2-digit Hour (24h) | `19` |
| `%M` | `%m` | 2-digit Minute | `40` |
| `%S` | `%s` | 2-digit Second | `00` |
| `%p` | `%P` | Prefix | `Vacation-` |
| `%s` | `%S` | Suffix | `-Playa` |
| `%c` | `%C` | Collision Counter | `_1` |
| `%e` | `%F` | Extension (dot included) | `.jpg` |

The implementation should support these standard tokens to ensure modern developers are familiar with the patterns, while also being capable of parsing/converting the old ExifRenamer tokens.

## 4. Tool Architecture

* **Language:** Go (Golang)
* **Libraries:**
  * **Cobra:** CLI interface command structure (e.g., input path argument, flags for dry-runs, suffixes, and custom destinations).
  * **Viper:** Configuration management (reading default directories, default suffixes, formatting rules from a config file like `config.yaml` or `.env`).
  * **EXIF Reader:** `github.com/evanoberholster/imagemeta` (a pure-Go, high-performance metadata parser that natively supports HEIC, JPEG, TIFF, PNG, and camera RAW formats without external binary dependencies like `exiftool`).
  * **Video Metadata Reader:** `github.com/abema/go-mp4` (to parse the movie header atom `mvhd` and extract the original video recording creation timestamp for MP4/MOV formats).

---

## 5. Execution Behavior & Features

* **Execution Threading Model (Single-Threaded):** To prevent race conditions during duplicate filename collision checks, the renaming and moving of files is processed sequentially (single-threaded). This ensures that collision suffixes (`_1`, `_2`, etc.) are assigned deterministically and files are not corrupted or overwritten.
* **Dry-Run Mode (Must-have):** Print planned actions (original path -> target path) without actually moving/renaming any files.
* **Fallback Mechanisms:** If a file has no EXIF data (e.g. video files without `mvhd`, unsupported image formats, or images stripped of metadata):
  1. Fall back to **file creation time**.
  2. If creation time is not available, fall back to **file modification time (`mtime`)**.
  3. Log a warning in either fallback case.
* **Collision Resolution (Agreed):** Strategy for handling duplicate timestamps (incrementing `%C` starting at `_1`, `_2`, etc. if a conflict is found in the target directory).
* **Supported Media Types:**
  * **Images:** JPEG, PNG, HEIC, TIFF, WebP, AVIF, and camera RAW formats (CR2, CR3, NEF, etc.).
  * **Videos:** MP4, MOV (with metadata extraction); other video formats fall back to file creation/modification times.
* **Non-Media and System Files Handling:**
  * **System Files:** Hidden files starting with a dot (like `.DS_Store` or `.localized`) are **silently ignored** and not processed or moved.
  * **Unsupported Extensions:** Non-media file types (e.g. `.txt`, `.zip`, `.pdf`) are **skipped** entirely and remain in their original location. A warning is logged in standard/verbose mode.
* **Execution Options:** Batch processing directories, recursively searching subdirectories, case conversion toggle for extensions (e.g. force lowercase), and quiet/verbose logging output.


---

## 6. Testing Strategy

To ensure code stability, readability, and correctness:
1. **Unit Tests (Pattern Translators):** Table-driven tests to verify that both standard tokens (e.g., `%Y`, `%m`, `%d`) and legacy ExifRenamer tokens (e.g., `%M`, `%D`, `%h`) are correctly parsed and translated into formatting strings.
2. **Integration Tests (File Renaming & Moving):**
   * Use Go's `t.TempDir()` to construct sandboxed source and destination folders.
   * Verify file-moving and renaming logic under dry-run (no files changed) vs actual execution.
   * Verify recursive directory parsing and configuration file parsing.
3. **Collision Tests:** Generate multiple files with identical timestamps and confirm they resolve to chronological duplicates (e.g., `_1`, `_2`) deterministically.
4. **Fallback Tests:** Test with images stripped of EXIF data to ensure the program falls back to file creation/modification times and prints the appropriate warning logs.


---

## 7. Config File Layout & CLI Command Structure

### A. YAML Configuration File Layout (`config.yaml` / `$HOME/.exifrenamer.yaml`)
```yaml
# Destination directory where organized folders/files will be placed
destination: "~/Pictures/Sorted"

# General runtime preferences
recursive: true
dry_run: false
lowercase_extension: true
verbose: false

# Naming and formatting rules
naming:
  # Folder hierarchy pattern.
  # Standard token format for: [Year]/[Year]-[Month]/[Year]-[Month]-[Day]-[Suffix]
  folder_pattern: "%Y/%Y-%m/%Y-%m-%d%s"
  
  # File renaming pattern.
  # Standard token format for: [Prefix][Year][Month][Day]-[Hour][Minute][Second][Suffix][Counter][Extension]
  file_pattern: "%p%Y%m%d-%H%M%S%s%c%e"
  
  # Default fallback strings if none are provided via CLI flags
  default_prefix: ""
  default_suffix: ""
```

### B. CLI Command Structure (Cobra-based)

#### Basic Usage Syntax:
```bash
exifrenamer [source_paths...] [flags]
```

#### CLI Flags:
* **Config Overrides & File Selection:**
  * `-c, --config <path>`: Path to a specific configuration file (overrides default search paths).
  * `-r, --recursive`: Recursively traverse source directories (overrides `recursive` config).

* **Output & Behavior Flags:**
  * `-d, --dest <path>`: Base destination directory (overrides `destination` config).
  * `--dry-run`: Dry-run mode. Displays what actions would be taken without making changes (overrides `dry_run` config).
  * `-l, --lower-ext`: Force lowercase file extensions, e.g. `.JPG` -> `.jpg` (overrides `lowercase_extension` config).

* **Prefix / Suffix Injection:**
  * `-p, --prefix <string>`: Temporary prefix to insert via `%p` or `%P` token (overrides `naming.default_prefix`).
  * `-s, --suffix <string>`: Temporary suffix to insert via `%s` or `%S` token (overrides `naming.default_suffix`).

* **Logging Control:**
  * `-v, --verbose`: Enable verbose logging to show warnings (e.g. EXIF fallback triggers) and processing steps.
  * `-q, --quiet`: Silence output except for critical errors.


