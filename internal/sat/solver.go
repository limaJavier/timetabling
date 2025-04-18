package sat

type SATSolver interface {
	Solve(*SAT) (SATSolution, error) // Returns a solution of the SAT instance if satisfiable, else returns nil (these are valid outputs where error shall be nil)
}
