package midground

type batch struct {
	processes   []*Process
	concurrency int
	doneCount   int
	doneChan    chan struct{}
}

func (s *batch) done() {
	s.doneCount++
	if s.doneCount == len(s.processes) {
		close(s.doneChan)
	}
}

func (s *batch) findBlocker(process *Process) *Process {
	// TODO: redo this using batchIndex
	if len(s.processes) == 1 {
		return nil
	}
	var runningCount int
	for i, otherProcess := range s.processes {
		if otherProcess == process {
			if i == 0 {
				return nil
			}
			previousProcess := s.processes[i-1]
			if previousProcess.IsScheduled() || previousProcess.IsBlocked() || runningCount == s.concurrency {
				return previousProcess
			}
			return nil
		} else if otherProcess.IsRunning() {
			runningCount++
		}
	}
	panic("job not in batch")
}

func (s *batch) runningJobs() (running []*Process) {
	for _, job := range s.processes {
		if job.IsRunning() {
			running = append(running, job)
		}
	}
	return
}
