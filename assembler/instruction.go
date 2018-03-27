package assembler

type instEncoder func(p *instructionParams) (machineCode uint16, ok bool)

type instruction struct {
	opCode          uint16
	setExtend       bool
	validateOperand operandValidator
	encoder         instEncoder
}

var standardInstructions = map[string]instruction{
	"TC":     {opCode: 000000, validateOperand: validateTCOperand},
	"RELINT": {opCode: 000003},
	"INHNT":  {opCode: 000004},
	"EXTEND": {opCode: 000006, setExtend: true},

	"CCS": {opCode: 010000, validateOperand: requireErasable},
	"TCF": {opCode: 010000, validateOperand: requireFixed},

	"DAS":  {opCode: 020000, validateOperand: requireErasable},
	"LXCH": {opCode: 022000, validateOperand: requireErasable},
	"INCR": {opCode: 024000, validateOperand: requireErasable},
	"ADS":  {opCode: 026000, validateOperand: requireErasable},

	"CA": {opCode: 030000, validateOperand: requireAnyMemoryOperand},

	"CS": {opCode: 040000, validateOperand: requireAnyMemoryOperand},

	"INDEX":  {opCode: 050000, validateOperand: validateINDEXOperand},
	"RESUME": {opCode: 050017},
	"DXH":    {opCode: 052000, validateOperand: requireErasable},
	"TS":     {opCode: 054000, validateOperand: requireErasable},
	"XCH":    {opCode: 056000, validateOperand: requireErasable},

	"AD": {opCode: 060000, validateOperand: requireAnyMemoryOperand},

	"MASK": {opCode: 070000, validateOperand: requireAnyMemoryOperand},

	//Implied Address codes
	"XXALQ":  {opCode: 000000},       //Replace with TC A
	"XLQ":    {opCode: 000001},       //Replace with TC L
	"RETURN": {opCode: 000002},       //Replace with TC Q
	"NOOP":   {encoder: noopEncoder}, //Replace with TCF if in Fixed, or CA if in Erasable
	"DDOUBL": {opCode: 020001},       //Replace with DAS A
	"ZL":     {opCode: 022007},       //Replace with LXCH
	"COM":    {opCode: 040000},       //Replace with CS A
	"DTCF":   {opCode: 052005},       //Replace with DXCH FB
	"DTCB":   {opCode: 052006},       //Replace with DXCH Z
	"OVSK":   {opCode: 054000},       //Replace with TS A
	"TCAA":   {opCode: 054005},       //Replace with TS Z
	"DOUBLE": {opCode: 060000},       //Replace with AD A
	"ZQ":     {opCode: 022007},       //Replace with QXCH
	"DCOM":   {opCode: 040001},       //Replace with DCS A
	"SQUARE": {opCode: 070000},       //Replace with MP A
}

var extendedInstructions = map[string]instruction{
	"READ":   {opCode: 000000, validateOperand: requireChannel},
	"WRITE":  {opCode: 001000, validateOperand: requireChannel},
	"RAND":   {opCode: 002000, validateOperand: requireChannel},
	"WAND":   {opCode: 003000, validateOperand: requireChannel},
	"ROR":    {opCode: 004000, validateOperand: requireChannel},
	"WOR":    {opCode: 005000, validateOperand: requireChannel},
	"RXOR":   {opCode: 006000, validateOperand: requireChannel},
	"EDRUPT": {opCode: 007000, validateOperand: requireChannel},

	"DV":  {opCode: 010000, validateOperand: requireErasable},
	"BZF": {opCode: 010000, validateOperand: requireFixed},

	"MSU":  {opCode: 020000, validateOperand: requireErasable},
	"QXCH": {opCode: 022000, validateOperand: requireErasable},
	"AUG":  {opCode: 024000, validateOperand: requireErasable},
	"DIM":  {opCode: 026000, validateOperand: requireErasable},

	"DCA": {opCode: 030000, validateOperand: requireAnyMemoryOperand},

	"DCS": {opCode: 040000, validateOperand: requireAnyMemoryOperand},

	"INDEX": {opCode: 050000, validateOperand: requireAnyMemoryOperand, setExtend: true},

	"SU":   {opCode: 060000, validateOperand: requireErasable},
	"BZMF": {opCode: 060000, validateOperand: requireFixed},

	"MP": {opCode: 070000, validateOperand: requireAnyMemoryOperand},
}

func noopEncoder(p *instructionParams) (uint16, bool) {
	if p.location.isErasable() {
		//Replace with CA A (030000) if in Erasable
		return 030000, true
	}

	//Replace with TCF [I+1] (010000 + (I+1)) if in Fixed
	nextLoc, err := p.location.nextValid()
	if err != nil {
		p.logger.LogErrorf("cannot implement %v at the end of fixed memory", p.instToken)
		return 0, false
	}

	// if nextLoc.isBeginingOfSwitchableBank() {
	// 	pl.LogWarningf("%v at this location may require a bank switch", p.instToken)
	// }

	return 010000 | nextLoc.asOperand(), true
}
