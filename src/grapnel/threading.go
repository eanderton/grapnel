package grapnel

type Signal chan bool
type Worker func(Signal)

func launchWorkers(count int, fn Worker) []Signal {
  quitSignals := make([]Signal, 0)
  for ii := 0; ii < count; ii++ {
    quit := make(Signal)
    quitSignals = append(quitSignals, quit) 
    go fn(quit)
  }
  return quitSignals
}

func signalAll(quitSignals []Signal) {
  for _, signal := range quitSignals {
    signal <- true
  }
}
