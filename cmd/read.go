package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/4lxs/msc/mem"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var readCmd = &cobra.Command{
	Use:   "read <program> <position> <count>",
	Short: "Search for PATTERN in memory of PROGRAM",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		program, p, c := args[0], args[1], args[2]

		position, err := strconv.ParseInt(p, 0, 0)
		if err != nil {
			log.Fatalln("failed to convert position")
		}
		count, err := strconv.ParseUint(c, 0, 0)
		if err != nil {
			log.Fatalln("failed to convert count")
		}

		pid, err := getPid(program)
		if err != nil {
			log.Fatalln("Unable to find process:", err)
		}
		memory := mem.NewMemory(pid)

		if memory == nil {
			log.Fatalln("failed to get memory")
		}

		buf := memory.Read(position, count)
		fmt.Println("memory:", string(buf))
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
