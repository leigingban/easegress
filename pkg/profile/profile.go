package profile

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"

	"github.com/megaease/easegateway/pkg/logger"
	"github.com/megaease/easegateway/pkg/option"
)

// Profile is the Profile interface.
type Profile interface {
	Close(wg *sync.WaitGroup)
}

type profile struct {
	cpuFile *os.File
	opt     *option.Options
}

// New creates a profile.
func New(opt *option.Options) (Profile, error) {
	p := &profile{
		opt: opt,
	}

	err := p.startCPUProfile()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *profile) startCPUProfile() error {
	if p.opt.CPUProfileFile == "" {
		return nil
	}

	f, err := os.Create(p.opt.CPUProfileFile)
	if err != nil {
		return fmt.Errorf("create cpu profile failed: %v", err)
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		return fmt.Errorf("start cpu profile failed: %v", err)
	}

	p.cpuFile = f

	logger.Infof("cpu profile: %s", p.opt.CPUProfileFile)

	return nil
}

func (p *profile) memoryProfile() {
	if p.opt.MemoryProfileFile == "" {
		return
	}

	// to include every allocated block in the profile
	runtime.MemProfileRate = 1

	logger.Infof("memory profile: %s", p.opt.MemoryProfileFile)
	f, err := os.Create(p.opt.MemoryProfileFile)
	if err != nil {
		logger.Errorf("create memory profile failed: %v", err)
		return
	}

	runtime.GC()         // get up-to-date statistics
	debug.FreeOSMemory() // help developer when using outside monitor tool

	if err := pprof.WriteHeapProfile(f); err != nil {
		logger.Errorf("write memory file failed: %v", err)
		return
	}
	if err := f.Close(); err != nil {
		logger.Errorf("close memory file failed: %v", err)
		return
	}
}

func (p *profile) Close(wg *sync.WaitGroup) {
	defer wg.Done()

	if p.cpuFile != nil {
		pprof.StopCPUProfile()
		err := p.cpuFile.Close()
		if err != nil {
			logger.Errorf("close %s failed: %v", p.opt.CPUProfileFile, err)
		}
	}

	p.memoryProfile()
}
