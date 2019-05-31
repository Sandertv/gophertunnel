package script

import (
	"bufio"
	"io"
	"strings"
)

// distributor is a global distributor used for distributing code lines.
var distributor *Distributor

// Distributor continuously scans the stdin for new script code written to it, and broadcasts it to any
// scripts that are subscribed to it.
type Distributor struct {
	subscribers []*Script
}

// Distribute starts distributing code lines read from the reader to all subscribes of the distributor.
func Distribute(reader io.Reader) *Distributor {
	distributor := &Distributor{}
	go func() {
		scanner := bufio.NewScanner(reader)
		// Continuously scan for new lines written to the stdin, and interpret them immediately.
		var codeStr string
		inBlock := false
		for scanner.Scan() {
			if scanner.Text() == "" {
				// No need to process empty lines.
				continue
			}
			if scanner.Text()[0] == '"' && !inBlock {
				// The very first character was a double quote, meaning a 'block' was started.
				codeStr = strings.Trim(scanner.Text()[1:], "\r\n 	")
				inBlock = true
			} else if inBlock {
				// If we're reading a block, just append the data to the code string with \n.
				codeStr += "\n" + strings.Trim(scanner.Text(), "\r\n 	")
			}
			if inBlock {
				if codeStr == "" {
					// Don't do anything if the code string is currently empty.
					continue
				}
				if codeStr[len(codeStr)-1] == '"' {
					// The last character of the code string is a double quote, so we end the block here,
					// execute it, and clear it for the next read.
					codeStr = codeStr[:len(codeStr)-1]
					for key, script := range distributor.subscribers {
						if script == nil {
							// Script was already removed from the distributor.
							continue
						}
						if script.closed {
							// The script was closed so we remove it from the distributor.
							distributor.subscribers[key] = nil
							continue
						}
						if err := script.state.DoString(codeStr); err != nil {
							script.ErrorLog.Println(err)
						}
					}
					codeStr = ""
					inBlock = false
				}
				continue
			}
			// We're not reading a block, so we just run the line immediately.
			for key, script := range distributor.subscribers {
				if script == nil {
					// Script was already removed from the distributor.
					continue
				}
				if script.closed {
					// The script was closed so we remove it from the distributor.
					distributor.subscribers[key] = nil
					continue
				}
				if err := script.state.DoString(scanner.Text()); err != nil {
					script.ErrorLog.Println(err)
				}
			}
		}
	}()
	return distributor
}

// Subscribe subscribes a script to all code distributed by the distributor.
func (distributor *Distributor) Subscribe(script *Script) {
	distributor.subscribers = append(distributor.subscribers, script)
}
