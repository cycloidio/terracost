// Package cost defines structures that represent cloud resources and states in a cloud-agnostic,
// as well as tool-agnostic way. The hierarchy of structs is as follows:
//
// - Component is the lowest level, it describes the cost of a single cloud entity (e.g. storage
// space or compute time).
//
// - Resource is a collection of components and directly correlates to cloud resources (e.g. VM instance).
//
// - State is a collection of resources that exist (or are planned to exist) at any given moment
// across one or multiple cloud providers.
//
// - Plan is a difference between two states. It includes the prior (current) state and a planned
// state and it can be used to retrieve a list of ResourceDiff's.
package cost
