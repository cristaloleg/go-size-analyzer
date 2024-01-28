package pkg

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"github.com/Zxilly/go-size-analyzer/pkg/tool"
	"github.com/goretk/gore"
)

func loadSectionMap(gofile *gore.GoFile) (sections *SectionMap) {
	sections = &SectionMap{Sections: make(map[string]*Section)}

	switch f := gofile.GetParsedFile().(type) {
	case *pe.File:
		sections.loadFromPe(f)
	case *elf.File:
		sections.loadFromElf(f)
	case *macho.File:
		sections.loadFromMacho(f)
	default:
		panic("This should not happened :(")
	}

	return
}

type SectionMap struct {
	Sections map[string]*Section
}

func (s *SectionMap) GetSectionName(addr uint64) string {
	for _, section := range s.Sections {
		if addr >= section.Addr && addr < section.AddrEnd {
			return section.Name
		}
	}
	return ""
}

func (s *SectionMap) GetSection(addr, size uint64) *Section {
	for _, section := range s.Sections {
		if addr >= section.Addr && addr < section.AddrEnd && addr+size <= section.AddrEnd {
			return section
		}
	}
	return nil
}

func (s *SectionMap) AssertSize(size uint64) error {
	sectionsSize := uint64(0)
	for _, section := range s.Sections {
		sectionsSize += section.Size
	}

	if sectionsSize > size {
		return fmt.Errorf("section size %d > file size %d", sectionsSize, size)
	}

	return nil
}

func (s *SectionMap) loadFromPe(file *pe.File) {
	imageBase := tool.GetImageBase(file)

	for _, section := range file.Sections {
		s.Sections[section.Name] = &Section{
			Name:         section.Name,
			Size:         uint64(section.Size),
			Offset:       uint64(section.Offset),
			End:          uint64(section.Offset + section.Size),
			Addr:         imageBase + uint64(section.VirtualAddress),
			AddrEnd:      imageBase + uint64(section.VirtualAddress+section.VirtualSize),
			OnlyInMemory: false, // pe file didn't have an only-in-memory section
		}
	}
	return
}

func (s *SectionMap) loadFromElf(file *elf.File) {
	for _, section := range file.Sections {
		// not exist in binary
		if section.Type == elf.SHT_NULL || section.Size == 0 {
			continue
		}

		if section.Type == elf.SHT_NOBITS {
			// seems like .bss section
			s.Sections[section.Name] = &Section{
				Name:         section.Name,
				Addr:         section.Addr,
				AddrEnd:      section.Addr + section.Size,
				OnlyInMemory: true,
			}
			continue
		}

		s.Sections[section.Name] = &Section{
			Name:         section.Name,
			Size:         section.FileSize,
			Offset:       section.Offset,
			End:          section.Offset + section.FileSize,
			Addr:         section.Addr,
			AddrEnd:      section.Addr + section.Size,
			OnlyInMemory: false,
		}
	}

	return
}

func (s *SectionMap) loadFromMacho(file *macho.File) {
	for _, section := range file.Sections {
		if section.Size == 0 {
			continue
		}

		if section.Offset == 0 {
			// seems like .bss section
			s.Sections[section.Name] = &Section{
				Name:         section.Name,
				Addr:         section.Addr,
				AddrEnd:      section.Addr + section.Size,
				OnlyInMemory: true,
			}
			continue
		}

		s.Sections[section.Name] = &Section{
			Name:         section.Name,
			Size:         section.Size,
			Offset:       uint64(section.Offset),
			End:          uint64(section.Offset) + section.Size,
			Addr:         section.Addr,
			AddrEnd:      section.Addr + section.Size,
			OnlyInMemory: false,
		}
	}

	return
}
