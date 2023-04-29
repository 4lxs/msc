package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/4lxs/msc/mem"
	"github.com/4lxs/msc/proc"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search PROGRAM PATTERN",
	Short: "Search for PATTERN in memory of PROGRAM",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		p, pattern := args[0], args[1]
		pid, err := getPid(p)
		if err != nil {
			log.Fatalln("Failed to find process:", err)
		}
		memory := mem.NewMemory(pid)
		position, err := memory.Search(pattern)
		if err != nil {
			log.Fatalln("Failed to search memory:", err)
		}

		fmt.Printf("memory: %x\n", position)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func getPid(program string) (pid int, err error) {
	pid, err = strconv.Atoi(program)
	if err != nil {
		pid, err = proc.GetPid(program)
	}
	return
}
