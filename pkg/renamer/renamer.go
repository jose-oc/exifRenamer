package renamer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jose-oc/imagesExifRenamer/pkg/metadata"
)

// Options holds configuration settings for the renaming process.
type Options struct {
	Destination        string
	Recursive          bool
	DryRun             bool
	LowercaseExtension bool
	Verbose            bool
	FolderPattern      string
	FilePattern        string
	DefaultPrefix      string
	DefaultSuffix      string
}

// Renamer performs the sequential renaming and moving of media files.
type Renamer struct {
	Options Options
}

// New creates a new Renamer with the given Options.
func New(opts Options) *Renamer {
	return &Renamer{Options: opts}
}

// Supported media file extensions (case-insensitive)
var supportedExtensions = map[string]bool{
	// Images
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".heic": true,
	".heif": true,
	".tiff": true,
	".tif":  true,
	".webp": true,
	".avif": true,
	// Camera RAW formats
	".cr2":  true,
	".cr3":  true,
	".nef":  true,
	".arw":  true,
	".rw2":  true,
	".dng":  true,
	".crw":  true,
	// Videos
	".mp4":  true,
	".mov":  true,
	".m4v":  true,
}

// IsLegacyPattern checks if the pattern is in legacy format.
// It returns true if the pattern contains uniquely legacy tokens.
func IsLegacyPattern(pattern string) bool {
	legacyOnly := []string{"%D", "%h", "%P", "%C", "%F"}
	for _, tok := range legacyOnly {
		if strings.Contains(pattern, tok) {
			return true
		}
	}
	return false
}

// TranslateLegacyPattern translates a legacy pattern format to standard format.
func TranslateLegacyPattern(pattern string) string {
	return TranslateLegacyToStandard(pattern)
}

// TranslateLegacyToStandard converts a legacy pattern format to standard format.
func TranslateLegacyToStandard(pattern string) string {
	var result strings.Builder
	runes := []rune(pattern)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '%' && i+1 < len(runes) {
			next := runes[i+1]
			switch next {
			case 'Y':
				result.WriteString("%Y")
			case 'M':
				result.WriteString("%m") // Legacy Month -> Standard Month
			case 'D':
				result.WriteString("%d") // Legacy Day -> Standard Day
			case 'h':
				result.WriteString("%H") // Legacy Hour -> Standard Hour
			case 'm':
				result.WriteString("%M") // Legacy Minute -> Standard Minute
			case 's':
				result.WriteString("%S") // Legacy Second -> Standard Second
			case 'P':
				result.WriteString("%p") // Legacy Prefix -> Standard Prefix
			case 'S':
				result.WriteString("%s") // Legacy Suffix -> Standard Suffix
			case 'C':
				result.WriteString("%c") // Legacy Counter -> Standard Counter
			case 'F':
				result.WriteString("%e") // Legacy Extension -> Standard Extension
			default:
				result.WriteRune('%')
				result.WriteRune(next)
			}
			i++
		} else {
			result.WriteRune(runes[i])
		}
	}
	return result.String()
}

// TranslatePattern converts an input pattern (legacy or standard) into the standard pattern format.
func TranslatePattern(pattern string) string {
	if IsLegacyPattern(pattern) {
		return TranslateLegacyToStandard(pattern)
	}
	return pattern
}

// FormatStandardPattern replaces standard placeholders in a standard pattern string.
func FormatStandardPattern(pattern string, t time.Time, prefix, suffix, collision, ext string) string {
	var result strings.Builder
	runes := []rune(pattern)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '%' && i+1 < len(runes) {
			next := runes[i+1]
			switch next {
			case 'Y':
				result.WriteString(t.Format("2006"))
			case 'y':
				result.WriteString(t.Format("06"))
			case 'm':
				result.WriteString(t.Format("01"))
			case 'd':
				result.WriteString(t.Format("02"))
			case 'H':
				result.WriteString(t.Format("15"))
			case 'M':
				result.WriteString(t.Format("04"))
			case 'S':
				result.WriteString(t.Format("05"))
			case 'p':
				result.WriteString(prefix)
			case 's':
				result.WriteString(suffix)
			case 'c':
				result.WriteString(collision)
			case 'e':
				result.WriteString(ext)
			default:
				result.WriteRune('%')
				result.WriteRune(next)
			}
			i++
		} else {
			result.WriteRune(runes[i])
		}
	}
	return result.String()
}

// FormatPattern formats a pattern string (which can be standard or legacy)
// with the provided metadata and suffix/prefix values.
func FormatPattern(pattern string, t time.Time, prefix, suffix, collision, ext string) string {
	stdPattern := TranslatePattern(pattern)
	return FormatStandardPattern(stdPattern, t, prefix, suffix, collision, ext)
}

// Process traverses the sources sequentially and moves/renames the files accordingly.
func (r *Renamer) Process(sources []string) error {
	files, err := r.scanSources(sources)
	if err != nil {
		return err
	}

	takenPaths := make(map[string]bool)

	for _, srcFile := range files {
		ext := filepath.Ext(srcFile)
		extLower := strings.ToLower(ext)

		// Check supported extensions
		if !supportedExtensions[extLower] {
			if r.Options.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: skipping unsupported file type: %s\n", srcFile)
			}
			continue
		}

		// Get creation time and source
		creationTime, source, err := metadata.GetMediaCreationTime(srcFile)
		if err != nil {
			return fmt.Errorf("error reading metadata for %s: %w", srcFile, err)
		}

		// Warning for fallback
		if source == "creation" || source == "modification" {
			fmt.Fprintf(os.Stderr, "Warning: metadata not found for %s. Fell back to file %s time.\n", srcFile, source)
		}

		// Choose extension casing
		targetExt := ext
		if r.Options.LowercaseExtension {
			targetExt = extLower
		}

		prefix := r.Options.DefaultPrefix
		suffix := r.Options.DefaultSuffix

		folderPat := TranslatePattern(r.Options.FolderPattern)
		filePat := TranslatePattern(r.Options.FilePattern)

		subFolder := FormatStandardPattern(folderPat, creationTime, prefix, suffix, "", "")
		destDir := filepath.Join(expandHome(r.Options.Destination), subFolder)

		var targetPath string
		counterInt := 0
		for {
			var counterVal string
			if counterInt > 0 {
				counterVal = fmt.Sprintf("_%d", counterInt)
			}
			formattedName := FormatStandardPattern(filePat, creationTime, prefix, suffix, counterVal, targetExt)
			targetPath = filepath.Join(destDir, formattedName)

			_, statErr := os.Stat(targetPath)
			existsOnDisk := !os.IsNotExist(statErr)

			if !existsOnDisk && !takenPaths[targetPath] {
				takenPaths[targetPath] = true
				break
			}
			counterInt++
		}

		// Output planned action
		if r.Options.DryRun {
			fmt.Printf("[DRY-RUN] %s -> %s\n", srcFile, targetPath)
		} else {
			if r.Options.Verbose {
				fmt.Printf("Moving %s -> %s\n", srcFile, targetPath)
			} else {
				fmt.Printf("%s -> %s\n", srcFile, targetPath)
			}

			// Create target folder structure
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destDir, err)
			}

			// Move the file
			if err := moveFile(srcFile, targetPath); err != nil {
				return fmt.Errorf("failed to move file %s to %s: %w", srcFile, targetPath, err)
			}
		}
	}

	return nil
}

func (r *Renamer) scanSources(sources []string) ([]string, error) {
	var files []string
	for _, source := range sources {
		source = expandHome(source)
		fi, err := os.Stat(source)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			if r.Options.Recursive {
				err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}
					if strings.HasPrefix(info.Name(), ".") {
						return nil
					}
					files = append(files, path)
					return nil
				})
				if err != nil {
					return nil, err
				}
			} else {
				entries, err := os.ReadDir(source)
				if err != nil {
					return nil, err
				}
				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}
					if strings.HasPrefix(entry.Name(), ".") {
						continue
					}
					files = append(files, filepath.Join(source, entry.Name()))
				}
			}
		} else {
			if strings.HasPrefix(fi.Name(), ".") {
				continue
			}
			files = append(files, source)
		}
	}
	return files, nil
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fallback to copy and delete
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, si.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	in.Close()
	return os.Remove(src)
}
