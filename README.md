# ExifRenamer CLI

A high-performance Command Line Interface (CLI) utility in Go to automatically rename and organize your photo and video libraries into structured directories using EXIF and container metadata creation dates. 

This tool is designed to fully replicate and modernize the file-handling behavior of the classic macOS **ExifRenamer** application based on custom format strings.

---

## Features

- **Multi-Format Support:**
  - **Images:** JPEG, PNG, HEIC, TIFF, WebP, AVIF, and camera RAW formats (CR2, CR3, DNG, NEF, ARW, etc.) using `imagemeta`.
  - **Videos:** MP4 and QuickTime MOV formats. It extracts the original recording time from the Movie Header Atom (`mvhd`) using `go-mp4`.
- **Robust Fallbacks:** If EXIF data or video creation metadata is missing, it falls back to the file creation (birth) time, and then to the file modification time (`mtime`), logging a warning.
- **Single-Threaded Safety:** Renaming and moving operations are executed sequentially (single-threaded) to prevent race conditions during duplicate filename collision checks.
- **Collision Resolution:** Automatically appends incrementing counters (`_1`, `_2`, etc.) when files share the same timestamp.
- **Dry-Run Preview:** Simulates execution and outputs planned operations without touching the disk.
- **Flexible Configuration:** Merge options between default settings, configuration files (YAML), and Command Line flags.

---

## Installation

### Prerequisites
- Go 1.26 or higher

### Build from Source
Clone this repository and run:
```bash
go build -o exifrenamer main.go
```

To build with a specific version tag:
```bash
go build -ldflags "-X github.com/jose-oc/imagesExifRenamer/cmd.Version=v1.0.0" -o exifrenamer main.go
```

---

## CLI Flags

Usage:
```bash
exifrenamer [source_paths...] [flags]
```

| Flag | Shorthand | Description | Default / Override |
|---|---|---|---|
| `--config` | `-c` | Path to a specific YAML configuration file | Overrides default search paths |
| `--recursive`| `-r` | Recursively traverse input source directories | `true` |
| `--dest` | `-d` | Base target destination directory | `$HOME/Pictures/Sorted` |
| `--dry-run` | — | Preview planned actions without editing files | `false` |
| `--lower-ext`| `-l` | Force file extensions to lowercase (e.g. `.JPG` to `.jpg`) | `true` |
| `--prefix` | `-p` | Set prefix to inject via `%p` or `%P` token | None |
| `--suffix` | `-s` | Set suffix to inject via `%s` or `%S` token | None |
| `--verbose` | `-v` | Enable detailed processing and warning logs | `false` |
| `--quiet` | `-q` | Silence all console outputs except for critical errors | `false` |
| `--version` | — | Display the compiled version info | — |

---

## Configuration File

The application automatically searches for a `.exifrenamer.yaml` or `config.yaml` configuration file in the following locations:
1. The current working directory
2. The user's home directory (`$HOME/`)

### Example `config.yaml`
```yaml
# Target folder
destination: "~/Pictures/Sorted"

# Runtime preferences
recursive: true
dry_run: false
lowercase_extension: true
verbose: false

# Formatting rules
naming:
  # Folder structure: [Year]/[Year]-[Month]/[Year]-[Month]-[Day]-[Suffix]
  folder_pattern: "%Y/%Y-%m/%Y-%m-%d-%s"
  
  # Filename: [Prefix][Year][Month][Day]-[Hour][Minute][Second]-[Suffix][Counter][Extension]
  file_pattern: "%p%Y%m%d-%H%M%S-%s%c%e"
  
  # Default suffix/prefix
  default_prefix: ""
  default_suffix: ""
```

---

## Format Token Standards

The tool supports standard tokens (similar to `strftime`) and translates legacy ExifRenamer tokens automatically:

| Standard Token | Legacy Token | Description | Example |
|---|---|---|---|
| `%Y` | `%Y` | 4-digit Year | `2023` |
| `%y` | — | 2-digit Year | `23` |
| `%m` | `%M` | 2-digit Month (01-12) | `08` |
| `%d` | `%D` | 2-digit Day (01-31) | `03` |
| `%H` | `%h` | 2-digit Hour (24-hour clock) | `19` |
| `%M` | `%m` | 2-digit Minute | `40` |
| `%S` | `%s` | 2-digit Second | `05` |
| `%p` | `%P` | Prefix string | `Vacation-` |
| `%s` | `%S` | Suffix string | `-Playa` |
| `%c` | `%C` | Sequential collision counter | `_1` |
| `%e` | `%F` | Extension (dot included) | `.jpg` |

---

## License
[MIT](LICENSE)
