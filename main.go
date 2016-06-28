package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

var emergencyDefaultContent = []byte(`##
# Host Database
#
# localhost is used to configure the loopback interface
# when the system is booting.  Do not change this entry.
##
127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost

`)

type Entry struct {
	networkAddress string
	hostname       string
	label          string
}

type EntryList struct {
	val []Entry
}

func (list *EntryList) appendEntry(e Entry) {
	for _, entry := range list.val {
		if checkForNetworkConflicts(e, entry) {
			fmt.Printf("New entry conflicts with: %s\n", entry)
			os.Exit(1)
		}
	}
	list.val = append(list.val, e)
}

func (list *EntryList) appendList(el *EntryList) {
	for _, entry := range el.val {
		list.appendEntry(entry)
	}
}

func (list *EntryList) toStringArray(includeComment bool) [][]string {
	stringArray := [][]string{}
	for _, entry := range list.val {
		stringEntry := []string{strings.TrimSpace(entry.networkAddress), entry.hostname}
		if entry.label != "" {
			if includeComment {
				stringEntry = append(stringEntry, "# "+entry.label)
			} else {
				stringEntry = append(stringEntry, entry.label)
			}
		} else {
			stringEntry = append(stringEntry, "")
		}
		stringArray = append(stringArray, stringEntry)
	}
	return stringArray
}

func (list *EntryList) prettyPrintList(output io.Writer) {
	table := tablewriter.NewWriter(output)
	table.SetHeader([]string{"Network Address", "Hostname", "Label"})
	for _, entry := range list.toStringArray(false) {
		table.Append(entry)
	}
	table.Render()
}

func (list *EntryList) formatListToString() string {
	lineConcatArray := []string{}
	for _, entry := range list.toStringArray(true) {
		lineConcatArray = append(lineConcatArray, strings.Join(entry, "	"))
	}
	return strings.Join(lineConcatArray, "\n")
}

type EntryFilterFunc func(Entry) bool

func (list *EntryList) Filter(filterFunc EntryFilterFunc) *EntryList {
	newList := &EntryList{}
	for _, entry := range list.val {
		if filterFunc(entry) {
			newList.val = append(newList.val, entry)
		}
	}
	return newList
}

func (list *EntryList) Write(path string) {
	hostsBytes := []byte(list.formatListToString())
	err := ioutil.WriteFile(path, append(emergencyDefaultContent, hostsBytes...), 0644)
	if err != nil {
		fmt.Printf("Error writing to hosts file: %s\n", err)
		if askForConfirmation("Would you like to restore to the default hosts file?") {
			restoreHostsWithDefault(path)
		} else {
			fmt.Println("Your hosts file is corrupted. Please restore it manually.")
		}
	}
}

func (list *EntryList) AddLabel(networkAddress, label string) bool {
	for idx, _ := range list.val {
		if list.val[idx].networkAddress == networkAddress {
			list.val[idx].label = label
			return true
		}
	}
	return false
}

func (check Entry) In(list *EntryList) bool {
	for _, entry := range list.val {
		if compareEntries(entry, check) {
			return true
		}
	}
	return false
}

// DEFAULT ENTRIES TO ALWAYS INCLUDE
// 127.0.0.1	localhost
// 255.255.255.255	broadcasthost
// ::1             localhost

var persistentEntries = &EntryList{
	val: []Entry{
		Entry{networkAddress: "127.0.0.1", hostname: "localhost", label: "default"},
		Entry{networkAddress: "255.255.255.255", hostname: "broadcasthost", label: "default"},
		Entry{networkAddress: "::1", hostname: "localhost", label: "default"},
	},
}

func compareEntries(base, compare Entry) bool {
	if base.hostname == compare.hostname && base.networkAddress == compare.networkAddress {
		return true
	}
	return false
}

func checkForNetworkConflicts(base, compare Entry) bool {
	if base.networkAddress == compare.networkAddress {
		return true
	}
	return false
}

func restoreHostsWithDefault(path string) {
	err := ioutil.WriteFile(path, emergencyDefaultContent, 0644)
	if err != nil {
		fmt.Println("Unable to restore hosts file to default! Your hosts file is in a bad state, manually restore it before further action.")
		os.Exit(1)
	}
}

func getHostEntries(filePath string) *EntryList {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open hosts file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	entries := &EntryList{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		// XOR
		if (!strings.HasPrefix(text, "#")) != (len(strings.Fields(text)) == 0) {

			entryFields := strings.Fields(text)
			entry := Entry{
				networkAddress: strings.TrimSpace(entryFields[0]),
				hostname:       strings.TrimSpace(entryFields[1]),
			}
			if len(entryFields) > 2 && entryFields[2] == "#" {
				entry.label = strings.Join(entryFields[3:], " ")
			}
			if !entry.In(persistentEntries) {
				entries.val = append(entries.val, entry)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read hosts file: %s\n", err)
		os.Exit(1)
	}
	return entries
}

func main() {
	app := cli.NewApp()
	app.Name = "goodhost"
	app.Version = "v0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "file, f",
			Value:  "/etc/hosts",
			Usage:  "your hosts file",
			EnvVar: "HOSTS_FILE",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "list all current entries",
			Action: func(c *cli.Context) error {
				currentEntries := getHostEntries(c.GlobalString("file"))
				currentEntries.prettyPrintList(c.App.Writer)
				return nil
			},
		},
		{
			Name:      "get",
			Usage:     "get an entry by network address",
			ArgsUsage: "192.168.1.1",
			Action: func(c *cli.Context) error {
				currentEntries := getHostEntries(c.GlobalString("file"))
				filteredEntries := currentEntries.Filter(func(e Entry) bool {
					if e.networkAddress == strings.TrimSpace(c.Args().First()) {
						return true
					}
					return false
				})
				filteredEntries.prettyPrintList(c.App.Writer)
				return nil
			},
		},
		{
			Name:      "set",
			Usage:     "set an entry",
			ArgsUsage: "myhost.com 192.168.16.1 \"my label\"",
			Action: func(c *cli.Context) error {
				if len(c.Args()) < 2 || len(c.Args()) > 3 {
					fmt.Println("Incorrect number of arguments. \"set\" only expects a hostname followed by a network address with an optional label as the third argument.")
					os.Exit(1)
				}
				newEntryList := getHostEntries(c.GlobalString("file"))
				newEntry := Entry{
					hostname:       strings.TrimSpace(c.Args()[0]),
					networkAddress: strings.TrimSpace(c.Args()[1]),
				}
				if len(c.Args()) == 3 {
					newEntry.label = strings.TrimSpace(c.Args()[2])
				}
				newEntryList.appendEntry(newEntry)
				newEntryList.Write(c.GlobalString("file"))
				return nil
			},
		},
		{
			Name:      "remove",
			Usage:     "remove an entry by network address",
			ArgsUsage: "192.168.1.1",
			Action: func(c *cli.Context) error {
				currentEntries := getHostEntries(c.GlobalString("file"))
				filteredEntries := currentEntries.Filter(func(e Entry) bool {
					if e.networkAddress == strings.TrimSpace(c.Args().First()) {
						return false
					}
					return true
				})
				filteredEntries.Write(c.GlobalString("file"))
				return nil
			},
		},
		{
			Name:      "label",
			Usage:     "label an existing entry by network address",
			ArgsUsage: "192.168.1.1 \"local dev\"",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 2 {
					fmt.Println("Incorrect number of arguments. \"label\" only expects a network address followed by a label")
					os.Exit(1)
				}
				currentEntries := getHostEntries(c.GlobalString("file"))
				if !currentEntries.AddLabel(c.Args()[0], c.Args()[1]) {
					fmt.Println("Entry not found in hosts file")
				} else {
					currentEntries.Write(c.GlobalString("file"))
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
