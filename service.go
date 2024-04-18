// package main

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"time"
//
//
// 

// 	"github.com/atotto/clipboard"
// )

// func getHistoryFilePath() string {
// 	// Get the system's temporary directory
// 	tempDir := os.TempDir()

// 	// Construct the path to the history file
// 	return filepath.Join(tempDir, "clipboard_history.txt")
// }

// func cacheClipboardHistory() {
// 	historyFile := getHistoryFilePath()
// 	if historyFile == "" {
// 		return
// 	}
// 	var lastClipboard string

// 	for {
// 		currentClipboard, err := clipboard.ReadAll()
// 		if err != nil {
// 			continue
// 		}

// 		if currentClipboard != "" {
// 			// Append current clipboard content to history file
// 			if currentClipboard == lastClipboard {
// 				continue
// 			}
// 			f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 			if err != nil {
// 				fmt.Println("Error opening history file:", err)
// 			} else {
// 				_, err := fmt.Fprintln(f, currentClipboard)
// 				if err != nil {
// 					fmt.Println("Error writing to history file:", err)
// 				}
// 				f.Close()
// 			}
// 		}
// 		time.Sleep(1 * time.Second)

// 		lastClipboard = currentClipboard
// 	}
// }

// func main() {
// 	// Start monitoring clipboard in background
// 	go cacheClipboardHistory()

// 	// Keep the main goroutine running
// 	select {}
// }

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
)

const (
	historyFile = "clipboard_history.txt"
	bufferSize  = 100 // Buffer size for the channel
)

func getHistoryFilePath() string {
	// Get the system's temporary directory
	tempDir := os.TempDir()

	// Construct the path to the history file
	return filepath.Join(tempDir, historyFile)
}

func cacheClipboardHistory(history chan<- string, done <-chan struct{}) {
	historyFile := getHistoryFilePath()
	if historyFile == "" {
		return
	}

	lastClipboard := ""

	// Open the history file for writing
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening history file:", err)
		return
	}
	defer f.Close()

	// Buffered writer for improved performance
	bufWriter := bufio.NewWriter(f)
	defer bufWriter.Flush()

	for {
		select {
		case <-done:
			return
		default:
			currentClipboard, err := clipboard.ReadAll()
			if err != nil {
				fmt.Println("Error reading clipboard:", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Check if clipboard content has changed
			if currentClipboard != lastClipboard {
				// Send current clipboard content to history channel
				select {
				case history <- currentClipboard:
					// Write clipboard content to file
					if _, err := bufWriter.WriteString(currentClipboard + "\n"); err != nil {
						fmt.Println("Error writing to history file:", err)
					}
					// Flush the buffer to ensure the content is written to the file
					if err := bufWriter.Flush(); err != nil {
						fmt.Println("Error flushing buffer:", err)
					}

					// Update last clipboard content
					lastClipboard = currentClipboard
				default:
					// Skip if the history channel is full
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	// Channel to communicate clipboard history
	history := make(chan string, bufferSize)

	// Channel for cancellation
	done := make(chan struct{})
	defer close(done)

	// Start monitoring clipboard in background
	go cacheClipboardHistory(history, done)

	// Keep the main goroutine running
	select {}
}
