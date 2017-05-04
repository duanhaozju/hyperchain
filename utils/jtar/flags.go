package main

import "flag"

const NOT_EXIST = "not exist"

var dir = flag.String("C","", "Temporarily changes directories to dir while processing the following inputfiles argument")

type Args struct {
	Control string
	Source  string
	Target  string
	Options map[string]string
}

func parseArgs() Args {
	args := Args{
		Options: make(map[string]string),
	}

	idx := 0
	for n, arg := range flag.Args() {
		if isPrefix(arg) {
			parseOptions(n, &args)
		} else {
			assignArgs(idx, arg, &args)
			idx += 1
		}
	}
	return args
}

func isPrefix(arg string) bool {
	if arg == "-C" {
		return true
	} else {
		return false
	}
}

func assignArgs(idx int, arg string, args *Args) {
	switch idx {
	case 0:
		args.Control = arg
	case 1:
		args.Target = arg
	case 2:
		args.Source = arg
	}
}

func parseOptions(idx int, args *Args) {

}

