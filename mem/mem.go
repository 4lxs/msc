package mem

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type RegionType uint

const (
	CODE RegionType = iota
	STACK
	HEAP
	ANONYMOUS // obtained via mmap
)

type Memory struct {
	pid     int
	regions []MemoryRegion
}

type MemoryRegion struct {
	begin    uint64
	end      uint64
	filename string
	kind     RegionType
}

func NewMemory(pid int) *Memory {
	regions := getProgramRegions(pid)

	return &Memory{
		pid:     pid,
		regions: regions,
	}
}

func (m *Memory) Read(position int64, count uint64) []byte {
	memFileName := fmt.Sprintf("/proc/%v/mem", m.pid)
	memFile, err := os.Open(memFileName)
	if err != nil {
		log.Fatal("Failed to open mem file:", err)
	}

	buf := make([]byte, count)
	_, err = memFile.ReadAt(buf, position)
	if err != nil {
		log.Fatal("Failed to read mem file:", err)
	}

	return buf
}

func (m *Memory) Search(pattern string) (positions []uint64, err error) {
	memFileName := fmt.Sprintf("/proc/%v/mem", m.pid)
	memFile, err := os.Open(memFileName)
	if err != nil {
		return
	}

	for _, region := range m.regions {
		// if region.kind == CODE {
		// 	continue
		// }
		positions = append(positions, region.Search(pattern, memFile)...)
	}
	return
}

func (m *MemoryRegion) Search(pattern string, memFile *os.File) (positions []uint64) {
	// log.Printf("Searching %v: %x-%x", m.filename, m.begin, m.end)
	const bufSize uint64 = 4096
	buf := make([]byte, bufSize)
	var currentPos uint64 = 0

	reader := io.NewSectionReader(memFile, int64(m.begin), int64(m.end-m.begin))

	for {
		n, err := reader.Read(buf)
		if err != nil && n == 0 {
			break
		}

		positions = append(positions, findAllSubstrings(string(buf), pattern)...)
		currentPos += uint64(n)
	}
	// log.Println("Found", positions)

	return
}

func findAllSubstrings(s, token string) (positions []uint64) {
	curr := 0
	for {
		ind := strings.Index(s[curr:], token)
		if ind == -1 {
			break
		}
		positions = append(positions, uint64(curr+ind))
		curr += ind + 1
	}
	return
}

func (m *MemoryRegion) Size() uint64 {
	return m.end - m.begin
}

func getProgramRegions(pid int) (regions []MemoryRegion) {
	exeLink := fmt.Sprintf("/proc/%v/exe", pid)
	exeName, err := os.Readlink(exeLink)
	if err != nil {
		log.Fatalln(err)
	}
	mmapFileName := fmt.Sprintf("/proc/%v/maps", pid)
	mmapFile, err := os.Open(mmapFileName)
	if err != nil {
		log.Fatalln(err)
	}

	/** maps file is looks like this:
	 * address           perms offset  dev   inode       pathname
	 * 00400000-00452000 r-xp 00000000 08:02 173521      /usr/bin/dbus-daemon
	 * 00651000-00652000 r--p 00051000 08:02 173521      /usr/bin/dbus-daemon
	 * 00652000-00655000 rw-p 00052000 08:02 173521      /usr/bin/dbus-daemon
	 * 00e03000-00e24000 rw-p 00000000 00:00 0           [heap]
	 * 00e24000-011f7000 rw-p 00000000 00:00 0           [heap]
	 * ...
	 * 35b1800000-35b1820000 r-xp 00000000 08:02 135522  /usr/lib64/ld-2.15.so
	 * ...
	 *
	 * note that the pathname can be empty (anonymous mmap)
	 * proc(5) manpage for more detail
	 */

	reader := bufio.NewReader(mmapFile)
	// skip first line (address perms ...)
	reader.ReadBytes('\n')

	var start, end uint64
	var read, write, exec, copyOnWrite rune
	var offset, devMajor, devMinor, inode int
	var pathname string

	for {
		_, err := fmt.Fscanf(reader, "%x-%x %c%c%c%c %x %x:%x %d", &start, &end, &read, &write, &exec, &copyOnWrite, &offset, &devMajor, &devMinor, &inode)
		if err != nil {
			break
		}
		pathname, _ = reader.ReadString('\n')
		pathname = strings.TrimSpace(pathname)
		// fmt.Println(start, end, read, write, exec, copyOnWrite, offset, devMajor, devMinor, inode, pathname)
		// fmt.Printf("pathname: '%v'\n", pathname)

		if read != 'r' {
			continue
		}

		var kind RegionType
		switch pathname {
		case exeName:
			kind = CODE
		case "[heap]":
			kind = HEAP
		case "[stack]":
			kind = STACK
		case "":
			kind = ANONYMOUS
		default:
			continue
		}

		regions = append(regions, MemoryRegion{start, end, pathname, kind})
	}
	return
}
