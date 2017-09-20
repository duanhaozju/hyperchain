//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package consensus

// consensus algorithm name
const (
	RBFT = "RBFT"
	NBFT = "NBFT"
)

// key for consensus
const (
	CONSENSUS_ALGO = "consensus.algo"
)

// Subscription type definitions
const (
	FILTER_View_Change_Finish = iota
	FILTER_Nextthing
)
