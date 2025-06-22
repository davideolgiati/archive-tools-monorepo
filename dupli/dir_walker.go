package main

import (
	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
	"errors"
	"os"
	"path"
)

type dirWalkerConfiguration struct {
	filterDirectory   func(string) bool
	fileCallback      func(File)
	directoryCallback func()
	skipEmpty         bool
}

type dirWalkerState struct {
	currentDirectory string
	currentFile      string
	directoriesQueue datastructures.Queue[string]
}

type dirwalkerStatistics struct {
	sizeProcessed   int64
	fileSeen        int
	directoriesSeen int
}

type dirWalker struct {
	configuration dirWalkerConfiguration
	state         dirWalkerState
	stats         dirwalkerStatistics
}

func NewWalker(skipEmpty bool) *dirWalker {
	walker := dirWalker{}

	walker.stats = dirwalkerStatistics{}
	walker.stats.fileSeen = 0
	walker.stats.directoriesSeen = 0
	walker.stats.sizeProcessed = 0

	walker.configuration = dirWalkerConfiguration{}
	walker.configuration.skipEmpty = skipEmpty
	walker.configuration.directoryCallback = nil
	walker.configuration.filterDirectory = nil
	walker.configuration.fileCallback = nil

	walker.state = dirWalkerState{}
	walker.state.directoriesQueue = datastructures.Queue[string]{}
	walker.state.directoriesQueue.Init()
	walker.state.currentDirectory = ""
	walker.state.currentFile = ""

	return &walker
}

func (walker *dirWalker) SetEntryPoint(directory string) {
	walker.state.directoriesQueue.Push(directory)
}

func (walker *dirWalker) SetDirectoryFilter(filterFn func(string) bool) {
	walker.configuration.filterDirectory = filterFn
}

func (walker *dirWalker) SetFileCallback(callback func(File)) {
	walker.configuration.fileCallback = callback
}

func (walker *dirWalker) SetDirectoryCallback(callback func()) {
	walker.configuration.directoryCallback = callback
}

func (walker *dirWalker) Walk() {
	var formattedSize commons.FileSize
	var objects []os.DirEntry
	var err error

	ui.AddNewNamedLine("directory-line", "Directories seen: %6d")
	ui.AddNewNamedLine("file-line", "Files seen: %12d")
	ui.AddNewNamedLine("size-line", "Processed: %10d %2s")

	for !walker.state.directoriesQueue.Empty() {
		walker.state.currentDirectory, err = walker.state.directoriesQueue.Pop()

		if err != nil {
			panic(err)
		}

		objects, err = os.ReadDir(walker.state.currentDirectory)

		if err == nil {
			walker.processDirectoryItems(&objects)

			formattedSize, err = commons.FormatFileSize(walker.stats.sizeProcessed)
			if err != nil {
				panic(err)
			}

			ui.UpdateNamedLine("directory-line", walker.stats.directoriesSeen)
			ui.UpdateNamedLine("file-line", walker.stats.fileSeen)
			ui.UpdateNamedLine("size-line", formattedSize.Value, *formattedSize.Unit)
		} else if !errors.Is(err, os.ErrPermission) {
			panic(err)
		}

		walker.configuration.directoryCallback()
	}
}

func (walker *dirWalker) processDirectoryItems(objects *[]os.DirEntry) {
	for _, obj := range *objects {
		walker.state.currentFile = path.Join(walker.state.currentDirectory, obj.Name())

		if obj.IsDir() {
			walker.processDirectoryEntry(&walker.state.currentFile)
		} else {
			walker.processFileEntry(&obj)
		}
	}
}

func (walker *dirWalker) processDirectoryEntry(directory *string) {
	if !walker.configuration.filterDirectory(*directory) {
		return
	}

	walker.stats.directoriesSeen += 1
	walker.state.directoriesQueue.Push(*directory)
}

func (walker *dirWalker) processFileEntry(obj *os.DirEntry) {
	infos, err := (*obj).Info()

	if err != nil {
		panic(err)
	}

	file := File{
		infos: infos,
		path:  walker.state.currentFile,
	}

	if !file.IsAllowed() {
		return
	}

	if walker.configuration.skipEmpty && file.infos.Size() == 0 {
		return
	}

	walker.stats.fileSeen += 1
	walker.stats.sizeProcessed += file.infos.Size()
	walker.configuration.fileCallback(file)
}
